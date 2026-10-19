package curtailment_test

import (
	"errors"
	"testing"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

// The data as it stood on 17 October 2026: wind from October 2021, up to the
// last half hour of Friday the 16th.
var bounds = curtailment.Bounds{
	First:  day("2021-10-01"),
	Latest: time.Date(2026, 10, 16, 23, 30, 0, 0, curtailment.Brasilia),
}

func TestNewPeriod(t *testing.T) {
	tests := []struct {
		name   string
		kind   curtailment.Kind
		anchor string
		want   curtailment.Period
	}{
		{
			name: "the latest day, by default",
			kind: curtailment.Daily,
			want: curtailment.Period{
				Kind: curtailment.Daily, Start: day("2026-10-16"), End: day("2026-10-16"), Through: day("2026-10-16"),
				PreviousStart: day("2026-10-15"), PreviousEnd: day("2026-10-15"),
				Earlier: day("2026-10-15"),
			},
		},
		{
			name:   "a week still running, against the same days of the week before",
			kind:   curtailment.Weekly,
			anchor: "2026-10-14",
			want: curtailment.Period{
				Kind: curtailment.Weekly, Start: day("2026-10-12"), End: day("2026-10-18"), Through: day("2026-10-16"),
				PreviousStart: day("2026-10-05"), PreviousEnd: day("2026-10-09"),
				Earlier: day("2026-10-11"),
			},
		},
		{
			name:   "a week from its Sunday",
			kind:   curtailment.Weekly,
			anchor: "2026-10-11",
			want: curtailment.Period{
				Kind: curtailment.Weekly, Start: day("2026-10-05"), End: day("2026-10-11"), Through: day("2026-10-11"),
				PreviousStart: day("2026-09-28"), PreviousEnd: day("2026-10-04"),
				Earlier: day("2026-10-04"), Later: day("2026-10-12"),
			},
		},
		{
			name:   "a month so far",
			kind:   curtailment.Monthly,
			anchor: "2026-10-02",
			want: curtailment.Period{
				Kind: curtailment.Monthly, Start: day("2026-10-01"), End: day("2026-10-31"), Through: day("2026-10-16"),
				PreviousStart: day("2026-09-01"), PreviousEnd: day("2026-09-16"),
				Earlier: day("2026-09-30"),
			},
		},
		{
			name:   "a whole month against a shorter one",
			kind:   curtailment.Monthly,
			anchor: "2026-03-31",
			want: curtailment.Period{
				Kind: curtailment.Monthly, Start: day("2026-03-01"), End: day("2026-03-31"), Through: day("2026-03-31"),
				PreviousStart: day("2026-02-01"), PreviousEnd: day("2026-02-28"),
				Earlier: day("2026-02-28"), Later: day("2026-04-01"),
			},
		},
		{
			name:   "the first week, with nothing before it",
			kind:   curtailment.Weekly,
			anchor: "2021-10-01",
			want: curtailment.Period{
				Kind: curtailment.Weekly, Start: day("2021-09-27"), End: day("2021-10-03"), Through: day("2021-10-03"),
				Later: day("2021-10-04"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var anchor time.Time
			if tt.anchor != "" {
				anchor = day(tt.anchor)
			}
			got, err := curtailment.NewPeriod(tt.kind, anchor, bounds)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("NewPeriod() =\n%s\nwant\n%s", show(got), show(tt.want))
			}
			if got.HasPrevious() != !tt.want.PreviousStart.IsZero() {
				t.Errorf("HasPrevious() = %v", got.HasPrevious())
			}
		})
	}
}

func TestNewPeriodRefusesADayOutsideTheData(t *testing.T) {
	for _, anchor := range []string{"2021-09-30", "2026-10-17"} {
		if _, err := curtailment.NewPeriod(curtailment.Daily, day(anchor), bounds); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("NewPeriod(%s) = %v, want domain.ErrNotFound", anchor, err)
		}
	}
	var invalid *domain.ValidationError
	if _, err := curtailment.NewPeriod("year", time.Time{}, bounds); !errors.As(err, &invalid) {
		t.Errorf("NewPeriod(year) = %v, want a *domain.ValidationError", err)
	}
}

func TestDayIsTheDateInBrasilia(t *testing.T) {
	// 01:30 UTC on the 17th is 22:30 on the 16th in Brasília.
	if got := curtailment.Day(time.Date(2026, 10, 17, 1, 30, 0, 0, time.UTC)); !got.Equal(day("2026-10-16")) {
		t.Errorf("Day() = %v", got)
	}
	if got := curtailment.StartOf(day("2026-10-16")); !got.Equal(time.Date(2026, 10, 16, 3, 0, 0, 0, time.UTC)) {
		t.Errorf("StartOf() = %v", got)
	}
}

func day(s string) time.Time {
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return d
}

func show(p curtailment.Period) string {
	f := func(t time.Time) string {
		if t.IsZero() {
			return "-"
		}
		return t.Format(time.DateOnly)
	}
	return string(p.Kind) + " " + f(p.Start) + " to " + f(p.End) + " through " + f(p.Through) +
		", previous " + f(p.PreviousStart) + " to " + f(p.PreviousEnd) +
		", earlier " + f(p.Earlier) + ", later " + f(p.Later)
}
