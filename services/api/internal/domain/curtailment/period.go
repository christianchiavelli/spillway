package curtailment

import (
	"fmt"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/domain"
)

// Kind is how long a period runs.
type Kind string

// The periods the pages read the data in.
const (
	Daily   Kind = "day"
	Weekly  Kind = "week"  // Monday to Sunday, as the calendar in Brazil runs
	Monthly Kind = "month" // a calendar month
)

// FieldPeriod and FieldDate name the inputs of a period, as the contract does.
const (
	FieldPeriod = "period"
	FieldDate   = "date"
)

// Bounds are how far the ONS's data goes.
type Bounds struct {
	First  time.Time // the first day of the first file
	Latest time.Time // the last half hour published, of the source that is furthest behind
}

// LastDay is the day of the last half hour published.
func (b Bounds) LastDay() time.Time { return Day(b.Latest) }

// Period is a day, a week or a month, read up to the latest day the ONS has
// published, and the span before it that it is compared with. Every field is a
// day, as Day returns it.
type Period struct {
	Kind  Kind
	Start time.Time
	End   time.Time // which may not be published yet
	// Through is the last day with data: End, or the last day published while
	// the period is still running.
	Through time.Time
	// PreviousStart and PreviousEnd are as many days of the period before as
	// this one has published, so a week so far is compared with the same days
	// of the week before. Both are zero when the data does not go back so far.
	PreviousStart, PreviousEnd time.Time
	// Earlier and Later are a day of the period before and of the period
	// after, for paging through them, or zero where the data ends.
	Earlier, Later time.Time
}

// HasPrevious reports whether the data goes back to the span compared with.
func (p Period) HasPrevious() bool { return !p.PreviousStart.IsZero() }

// NewPeriod returns the period of the given kind that holds anchor, a day.
// The zero anchor is the last day published. A day outside the data returns
// an error that matches domain.ErrNotFound.
func NewPeriod(kind Kind, anchor time.Time, b Bounds) (Period, error) {
	first, last := b.First, b.LastDay()
	if anchor.IsZero() {
		anchor = last
	}
	if anchor.Before(first) || anchor.After(last) {
		return Period{}, fmt.Errorf("%s is outside the data, from %s to %s: %w",
			anchor.Format(time.DateOnly), first.Format(time.DateOnly), last.Format(time.DateOnly), domain.ErrNotFound)
	}

	var p Period
	switch kind {
	case Daily:
		p = Period{Kind: kind, Start: anchor, End: anchor, PreviousStart: anchor.AddDate(0, 0, -1)}
	case Weekly:
		monday := anchor.AddDate(0, 0, -((int(anchor.Weekday()) + 6) % 7))
		p = Period{Kind: kind, Start: monday, End: monday.AddDate(0, 0, 6), PreviousStart: monday.AddDate(0, 0, -7)}
	case Monthly:
		start := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, time.UTC)
		p = Period{Kind: kind, Start: start, End: start.AddDate(0, 1, -1), PreviousStart: start.AddDate(0, -1, 0)}
	default:
		var vs domain.Violations
		vs.Add(FieldPeriod, domain.CodeInvalid)
		return Period{}, vs.Err()
	}

	p.Through = earliest(p.End, last)
	before := p.Start.AddDate(0, 0, -1) // the last day of the period before
	p.PreviousEnd = earliest(p.PreviousStart.AddDate(0, 0, DaysBetween(p.Start, p.Through)), before)
	if p.PreviousEnd.Before(first) {
		p.PreviousStart, p.PreviousEnd = time.Time{}, time.Time{}
	}
	if !before.Before(first) {
		p.Earlier = before
	}
	if next := p.End.AddDate(0, 0, 1); !next.After(last) {
		p.Later = next
	}
	return p, nil
}
