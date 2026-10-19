// Package curtailment holds what the grid operator (ONS) publishes about the
// wind and solar energy it keeps from being generated: its words, the days and
// periods it is read in, and the sums Spillway keeps of it.
package curtailment

import (
	"errors"
	"time"
)

// ErrNoData is returned until the first of the ONS's files is loaded, which
// takes a few minutes after a fresh start.
var ErrNoData = errors.New("no curtailment data loaded yet")

// Source is where the curtailed energy would have come from.
type Source string

// The sources the ONS curtails and publishes.
const (
	Wind  Source = "wind"
	Solar Source = "solar"
)

// Sources lists every source, in the order the pages show them.
func Sources() []Source { return []Source{Wind, Solar} }

// Valid reports whether s is a source the ONS publishes.
func (s Source) Valid() bool { return s == Wind || s == Solar }

// Subsystem is one of the four parts of Brazil's interconnected grid, coded as
// the ONS codes them in its open data.
type Subsystem string

// The four subsystems.
const (
	North     Subsystem = "N"
	Northeast Subsystem = "NE"
	Southeast Subsystem = "SE" // the ONS joins the Centre-West to the Southeast
	South     Subsystem = "S"
)

// Subsystems lists every subsystem, in the contract's order.
func Subsystems() []Subsystem { return []Subsystem{North, Northeast, Southeast, South} }

// Valid reports whether s is one of the four subsystems.
func (s Subsystem) Valid() bool {
	switch s {
	case North, Northeast, Southeast, South:
		return true
	default:
		return false
	}
}

// Reason is why the ONS limited a plant, as it codes it.
type Reason string

// The reasons the ONS gives.
const (
	ReasonEnergy      Reason = "ENE" // the plants could make more than the grid was using
	ReasonReliability Reason = "CNF" // held back to keep the grid within its safety limits
	ReasonExternal    Reason = "REL" // equipment outside the plant, mostly transmission, could not carry it
	ReasonAccess      Reason = "PAR" // a limit set in the plant's terms of connection
)

// Valid reports whether r is a reason the ONS gives.
func (r Reason) Valid() bool {
	switch r {
	case ReasonEnergy, ReasonReliability, ReasonExternal, ReasonAccess:
		return true
	default:
		return false
	}
}

// Brasilia is the ONS's clock. Brazil has kept no summer time since 2019,
// before the first file, so one fixed offset reads every instant it publishes.
var Brasilia = time.FixedZone("BRT", -3*60*60)

// Day returns the calendar day t falls on in Brasília, as midnight UTC of that
// date: days compare, add and travel as dates with no zone to get wrong.
func Day(t time.Time) time.Time {
	y, m, d := t.In(Brasilia).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// StartOf returns the instant day begins in Brasília.
func StartOf(day time.Time) time.Time {
	y, m, d := day.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, Brasilia)
}

// DaysBetween counts the days from a to b: 0 for the same day.
func DaysBetween(a, b time.Time) int {
	return int(b.Sub(a).Hours() / 24)
}

func earliest(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
