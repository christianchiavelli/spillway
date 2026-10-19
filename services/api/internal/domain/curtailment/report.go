package curtailment

import "time"

// Potential is what the plants could have generated: what they did, and what
// they were kept from.
func (e Energy) Potential() float64 { return e.CurtailedMWh + e.GeneratedMWh }

// Split is curtailed energy by source, in MWh.
type Split struct {
	Wind, Solar float64
}

// Add puts mwh under its source.
func (s *Split) Add(source Source, mwh float64) {
	switch source {
	case Wind:
		s.Wind += mwh
	case Solar:
		s.Solar += mwh
	}
}

// Slot is a stretch of time, a half hour or a day, and what each source lost
// in it. Start is an instant.
type Slot struct {
	Start     time.Time
	Curtailed Split
}

// SubsystemSplit is what each source lost in one subsystem.
type SubsystemSplit struct {
	Subsystem Subsystem
	Curtailed Split
}

// ReasonSplit is what each source lost for one reason.
type ReasonSplit struct {
	Reason    Reason
	Curtailed Split
}

// Overview is a period's curtailment, the ways the overview page reads it.
type Overview struct {
	Period Period
	Latest time.Time // the last half hour published
	Wind   Energy
	Solar  Energy
	// PreviousCurtailedMWh is the energy lost over the span the period is
	// compared with, when it has one.
	PreviousCurtailedMWh float64
	PlantsLimited        int // the plants that lost any energy in the period
	PlantsReporting      int
	// Series is the half hours of a day, or the days of a week or a month, up
	// to the last one published.
	Series      []Slot
	BySubsystem []SubsystemSplit // all four, in the contract's order
	ByReason    []ReasonSplit    // the reasons given in the period
}

// PlantFilter narrows the plants ranked. An empty field matches every plant.
type PlantFilter struct {
	Source    Source
	Subsystem Subsystem
}

// PlantTotal is one plant's period.
type PlantTotal struct {
	Plant
	Energy
	LimitedMinutes int
}

// PlantRanking is the plants the period's limits cost the most, one page of
// them, most energy lost first.
type PlantRanking struct {
	Period    Period
	Latest    time.Time
	Reporting int // the plants that reported in the period
	Limited   int // of those, the ones that lost any energy: the ranking's length
	Items     []PlantTotal
}

// MonthTotal is one month of the whole grid. A source is nil in the months
// before the ONS published it.
type MonthTotal struct {
	Start time.Time // its first day
	Wind  *Energy
	Solar *Energy
}

// History is every month the ONS has published.
type History struct {
	Latest time.Time
	Months []MonthTotal
}

// SubsystemSeries is the energy one source lost in one subsystem, a day at a
// time.
type SubsystemSeries struct {
	Source       Source
	Subsystem    Subsystem
	CurtailedMWh []float64
}

// Recent is the latest days published, by source and subsystem.
type Recent struct {
	From, To time.Time
	Series   []SubsystemSeries
}
