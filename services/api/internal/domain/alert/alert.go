// Package alert holds the rules of an alert: a threshold on the energy the
// grid operator (ONS) keeps wind or solar plants from generating, in one
// subsystem, in one day.
package alert

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

// The limits of an alert. The API contract states the same numbers, and a test
// in the HTTP adapter fails if the two ever disagree.
const (
	NameMaxLength = 80
	ThresholdMin  = 1
	ThresholdMax  = 1_000_000
	MaxPerAccount = 20
)

// Field names, as the API contract spells them.
const (
	FieldName      = "name"
	FieldSource    = "source"
	FieldSubsystem = "subsystem"
	FieldThreshold = "thresholdMwh"
)

// ErrLimitReached is returned when an account already holds MaxPerAccount
// alerts. It matches domain.ErrConflict.
var ErrLimitReached = fmt.Errorf("an account holds at most %d alerts: %w", MaxPerAccount, domain.ErrConflict)

// ErrNameTaken is returned when the account already has an alert by that name,
// ignoring case.
var ErrNameTaken error = &domain.ConflictError{Field: FieldName, Code: domain.CodeTaken}

// RecentDays is how far back an alert's activity is counted: the days up to
// the latest the ONS has published.
const RecentDays = 30

// Alert is a rule someone keeps: tell me when curtailment in a subsystem goes
// past a threshold in one day.
type Alert struct {
	ID           uuid.UUID
	AccountID    uuid.UUID
	Name         string
	Source       curtailment.Source
	Subsystem    curtailment.Subsystem
	ThresholdMWh int32
	CreatedAt    time.Time
}

// Draft is an alert as it was asked for, before any rule is checked.
type Draft struct {
	Name         string
	Source       string
	Subsystem    string
	ThresholdMWh int32
}

// New checks a draft and returns the alert it describes, or a
// *domain.ValidationError with every rule the draft breaks.
func New(id, accountID uuid.UUID, d Draft, now time.Time) (Alert, error) {
	var vs domain.Violations

	name := strings.TrimSpace(d.Name)
	switch {
	case name == "":
		vs.Add(FieldName, domain.CodeRequired)
	case utf8.RuneCountInString(name) > NameMaxLength:
		vs.Add(FieldName, domain.CodeTooLong)
	case strings.ContainsFunc(name, unicode.IsControl):
		vs.Add(FieldName, domain.CodeInvalid)
	}

	source := curtailment.Source(d.Source)
	if !source.Valid() {
		vs.Add(FieldSource, domain.CodeInvalid)
	}

	subsystem := curtailment.Subsystem(d.Subsystem)
	if !subsystem.Valid() {
		vs.Add(FieldSubsystem, domain.CodeInvalid)
	}

	if d.ThresholdMWh < ThresholdMin || d.ThresholdMWh > ThresholdMax {
		vs.Add(FieldThreshold, domain.CodeOutOfRange)
	}

	if err := vs.Err(); err != nil {
		return Alert{}, err
	}
	return Alert{
		ID:           id,
		AccountID:    accountID,
		Name:         name,
		Source:       source,
		Subsystem:    subsystem,
		ThresholdMWh: d.ThresholdMWh,
		CreatedAt:    now.UTC().Truncate(time.Microsecond), // what Postgres keeps
	}, nil
}

// Activity is what an alert's threshold met in the data.
type Activity struct {
	LastFiredOn time.Time // the latest day it fired, or zero if it never has
	DaysFired   int       // the days it fired in the RecentDays to the latest published
}

// Firing is a day an alert's source and subsystem lost more than its
// threshold: as much as the threshold does not set it off. Firings are read
// from the data, not recorded as they happen, so a day the ONS revises can gain
// or lose one (ADR 0021).
type Firing struct {
	AlertID      uuid.UUID
	AlertName    string
	Day          time.Time
	CurtailedMWh float64
	ThresholdMWh int32
}
