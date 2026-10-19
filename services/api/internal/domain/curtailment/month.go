package curtailment

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"time"
)

// ErrBadRecord is returned for a record the ONS could not have meant: a file
// that holds one is kept out whole, never in part.
var ErrBadRecord = errors.New("a record the ONS could not have meant")

// Record is one plant's half hour, as the ONS publishes it. Power is in MWmed,
// the average over the half hour, so a half hour's energy in MWh is half of it.
type Record struct {
	Plant     string // the ONS's id of a plant, or of a group of plants on one connection
	PlantName string
	State     string // the two letters of the plant's state
	Subsystem Subsystem
	At        time.Time // when the half hour starts
	// GeneratedMW is what the plant generated. It can be below zero at night,
	// when a plant only draws power.
	GeneratedMW float64
	// CurtailedMW is the ONS's GNRa: what it estimates the plant could have
	// generated without the limit, less what it did, never below zero, and zero
	// when there was no limit.
	CurtailedMW float64
	Reason      Reason // why the plant was limited, or "" when it was not
	// LimitedMinutes is the ONS's count of the half hour's minutes under a
	// limit. A few in millions fall outside 0 to 30, so it is clamped there: a
	// half hour has thirty minutes, and one odd count is no reason to drop a month.
	LimitedMinutes int
}

// A half hour's energy, in MWh, is its average power, in MW, times this.
const halfHour = 0.5

// Energy is what plants generated and what the ONS kept them from generating,
// in MWh.
type Energy struct {
	CurtailedMWh float64
	GeneratedMWh float64
}

// Plant is a plant, or a group of plants on one connection, as last seen.
type Plant struct {
	ID        string
	Name      string
	State     string
	Subsystem Subsystem
	Source    Source
	LastSeen  time.Time // the last day it reported
}

// PlantDay is one plant's day.
type PlantDay struct {
	Plant     string
	Day       time.Time
	Subsystem Subsystem
	Energy
	LimitedMinutes int
}

// SubsystemDay is one source's day in one subsystem.
type SubsystemDay struct {
	Day       time.Time
	Subsystem Subsystem
	Energy
}

// ReasonDay is the energy one source lost in a day for one reason.
type ReasonDay struct {
	Day          time.Time
	Reason       Reason
	CurtailedMWh float64
}

// HalfHour is the energy one source lost in a half hour, the whole grid over.
type HalfHour struct {
	Start        time.Time
	CurtailedMWh float64
}

// Month is one of the ONS's files, a source's month, summed the ways Spillway
// reads it. Each list is in a stable order: by day, then by key.
type Month struct {
	Source        Source
	Start         time.Time // its first day
	Latest        time.Time // its last half hour, zero when the file holds none
	Records       int
	Plants        []Plant
	PlantDays     []PlantDay
	SubsystemDays []SubsystemDay
	ReasonDays    []ReasonDay
	HalfHours     []HalfHour
}

// MonthBuilder sums a month's records as they are read, so a file is never
// held whole in memory.
type MonthBuilder struct {
	source      Source
	start       time.Time
	from, until time.Time // the month's instants, in Brasília
	records     int
	latest      time.Time
	plants      map[string]Plant
	plantSeen   map[string]time.Time
	plantDays   map[plantDayKey]*PlantDay
	subsystems  map[subsystemDayKey]*SubsystemDay
	reasons     map[reasonDayKey]*ReasonDay
	halfHours   map[int64]*HalfHour
}

type (
	plantDayKey struct {
		plant string
		day   int64
	}
	subsystemDayKey struct {
		day       int64
		subsystem Subsystem
	}
	reasonDayKey struct {
		day    int64
		reason Reason
	}
)

// NewMonthBuilder starts the sums of source's month, given as any day in it.
func NewMonthBuilder(source Source, month time.Time) *MonthBuilder {
	start := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	return &MonthBuilder{
		source:     source,
		start:      start,
		from:       StartOf(start),
		until:      StartOf(start.AddDate(0, 1, 0)),
		plants:     map[string]Plant{},
		plantSeen:  map[string]time.Time{},
		plantDays:  map[plantDayKey]*PlantDay{},
		subsystems: map[subsystemDayKey]*SubsystemDay{},
		reasons:    map[reasonDayKey]*ReasonDay{},
		halfHours:  map[int64]*HalfHour{},
	}
}

// Add sums one record in, or says what is wrong with it.
func (b *MonthBuilder) Add(r Record) error {
	if err := b.check(r); err != nil {
		return fmt.Errorf("%s at %s: %w", r.Plant, r.At.In(Brasilia).Format(time.DateTime), err)
	}
	b.records++
	day := Day(r.At)
	key := day.Unix()
	curtailed := r.CurtailedMW * halfHour
	generated := r.GeneratedMW * halfHour

	pd := b.plantDays[plantDayKey{r.Plant, key}]
	if pd == nil {
		pd = &PlantDay{Plant: r.Plant, Day: day}
		b.plantDays[plantDayKey{r.Plant, key}] = pd
	}
	pd.Subsystem = r.Subsystem
	pd.CurtailedMWh += curtailed
	pd.GeneratedMWh += generated
	pd.LimitedMinutes += min(max(r.LimitedMinutes, 0), 30)

	sd := b.subsystems[subsystemDayKey{key, r.Subsystem}]
	if sd == nil {
		sd = &SubsystemDay{Day: day, Subsystem: r.Subsystem}
		b.subsystems[subsystemDayKey{key, r.Subsystem}] = sd
	}
	sd.CurtailedMWh += curtailed
	sd.GeneratedMWh += generated

	if r.Reason != "" && curtailed > 0 {
		rd := b.reasons[reasonDayKey{key, r.Reason}]
		if rd == nil {
			rd = &ReasonDay{Day: day, Reason: r.Reason}
			b.reasons[reasonDayKey{key, r.Reason}] = rd
		}
		rd.CurtailedMWh += curtailed
	}

	// Every half hour any plant reported gets a slot, so a half hour with no
	// curtailment reads as zero, and only one not published yet is missing.
	hh := b.halfHours[r.At.Unix()]
	if hh == nil {
		hh = &HalfHour{Start: r.At}
		b.halfHours[r.At.Unix()] = hh
	}
	hh.CurtailedMWh += curtailed

	// A plant is known by its latest record: names and states do change.
	if seen, ok := b.plantSeen[r.Plant]; !ok || !r.At.Before(seen) {
		b.plantSeen[r.Plant] = r.At
		b.plants[r.Plant] = Plant{ID: r.Plant, Name: r.PlantName, State: r.State, Subsystem: r.Subsystem, Source: b.source, LastSeen: day}
	}
	if r.At.After(b.latest) {
		b.latest = r.At
	}
	return nil
}

func (b *MonthBuilder) check(r Record) error {
	switch {
	case r.Plant == "":
		return fmt.Errorf("no plant id: %w", ErrBadRecord)
	case !r.Subsystem.Valid():
		return fmt.Errorf("subsystem %q: %w", r.Subsystem, ErrBadRecord)
	case r.Reason != "" && !r.Reason.Valid():
		return fmt.Errorf("reason %q: %w", r.Reason, ErrBadRecord)
	case r.At.Before(b.from) || !r.At.Before(b.until):
		return fmt.Errorf("outside %s: %w", b.start.Format("2006-01"), ErrBadRecord)
	case r.At.Sub(b.from)%(30*time.Minute) != 0:
		return fmt.Errorf("not on a half hour: %w", ErrBadRecord)
	case !finite(r.GeneratedMW) || !finite(r.CurtailedMW):
		return fmt.Errorf("a power that is not a number: %w", ErrBadRecord)
	case r.CurtailedMW < 0:
		return fmt.Errorf("curtailed %v MW, below zero: %w", r.CurtailedMW, ErrBadRecord)
	}
	return nil
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Month returns the sums.
func (b *MonthBuilder) Month() Month {
	m := Month{Source: b.source, Start: b.start, Latest: b.latest, Records: b.records}
	m.Plants = slices.SortedFunc(maps.Values(b.plants), func(a, b Plant) int { return cmp.Compare(a.ID, b.ID) })
	m.PlantDays = sorted(b.plantDays, func(a, b PlantDay) int {
		return cmp.Or(a.Day.Compare(b.Day), cmp.Compare(a.Plant, b.Plant))
	})
	m.SubsystemDays = sorted(b.subsystems, func(a, b SubsystemDay) int {
		return cmp.Or(a.Day.Compare(b.Day), cmp.Compare(a.Subsystem, b.Subsystem))
	})
	m.ReasonDays = sorted(b.reasons, func(a, b ReasonDay) int {
		return cmp.Or(a.Day.Compare(b.Day), cmp.Compare(a.Reason, b.Reason))
	})
	m.HalfHours = sorted(b.halfHours, func(a, b HalfHour) int { return a.Start.Compare(b.Start) })
	return m
}

func sorted[K comparable, V any](m map[K]*V, compare func(a, b V) int) []V {
	out := make([]V, 0, len(m))
	for _, v := range m {
		out = append(out, *v)
	}
	slices.SortFunc(out, compare)
	return out
}
