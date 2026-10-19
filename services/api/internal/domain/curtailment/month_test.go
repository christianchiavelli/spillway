package curtailment_test

import (
	"errors"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

func at(hour, minute int, d int) time.Time {
	return time.Date(2026, 10, d, hour, minute, 0, 0, curtailment.Brasilia)
}

func TestMonthBuilderSumsTheRecords(t *testing.T) {
	b := curtailment.NewMonthBuilder(curtailment.Wind, day("2026-10-16"))
	for _, r := range []curtailment.Record{
		{Plant: "CJU_A", PlantName: "Conj. A", State: "RN", Subsystem: curtailment.Northeast, At: at(13, 0, 16), GeneratedMW: 100, CurtailedMW: 40, Reason: curtailment.ReasonReliability, LimitedMinutes: 30},
		{Plant: "CJU_A", PlantName: "Conj. A II", State: "RN", Subsystem: curtailment.Northeast, At: at(13, 30, 16), GeneratedMW: 90, CurtailedMW: 20, Reason: curtailment.ReasonExternal, LimitedMinutes: 12},
		{Plant: "CJU_B", PlantName: "Conj. B", State: "RS", Subsystem: curtailment.South, At: at(13, 0, 16), GeneratedMW: 60},
		// Limited, but nothing lost: the limit sat above what the wind gave.
		// The ONS counted 59 minutes in a half hour, which holds thirty.
		{Plant: "CJU_B", PlantName: "Conj. B", State: "RS", Subsystem: curtailment.South, At: at(0, 0, 17), GeneratedMW: -1, Reason: curtailment.ReasonEnergy, LimitedMinutes: 59},
	} {
		if err := b.Add(r); err != nil {
			t.Fatal(err)
		}
	}

	m := b.Month()

	if m.Source != curtailment.Wind || !m.Start.Equal(day("2026-10-01")) || m.Records != 4 || !m.Latest.Equal(at(0, 0, 17)) {
		t.Errorf("month = %v %v %d records, latest %v", m.Source, m.Start, m.Records, m.Latest)
	}
	wantPlants := []curtailment.Plant{
		{ID: "CJU_A", Name: "Conj. A II", State: "RN", Subsystem: curtailment.Northeast, Source: curtailment.Wind, LastSeen: day("2026-10-16")},
		{ID: "CJU_B", Name: "Conj. B", State: "RS", Subsystem: curtailment.South, Source: curtailment.Wind, LastSeen: day("2026-10-17")},
	}
	if !slices.Equal(m.Plants, wantPlants) {
		t.Errorf("plants = %+v", m.Plants)
	}
	wantDays := []curtailment.PlantDay{
		{Plant: "CJU_A", Day: day("2026-10-16"), Subsystem: curtailment.Northeast, Energy: curtailment.Energy{CurtailedMWh: 30, GeneratedMWh: 95}, LimitedMinutes: 42},
		{Plant: "CJU_B", Day: day("2026-10-16"), Subsystem: curtailment.South, Energy: curtailment.Energy{GeneratedMWh: 30}},
		{Plant: "CJU_B", Day: day("2026-10-17"), Subsystem: curtailment.South, Energy: curtailment.Energy{GeneratedMWh: -0.5}, LimitedMinutes: 30},
	}
	if !slices.Equal(m.PlantDays, wantDays) {
		t.Errorf("plant days = %+v", m.PlantDays)
	}
	wantSubsystems := []curtailment.SubsystemDay{
		{Day: day("2026-10-16"), Subsystem: curtailment.Northeast, Energy: curtailment.Energy{CurtailedMWh: 30, GeneratedMWh: 95}},
		{Day: day("2026-10-16"), Subsystem: curtailment.South, Energy: curtailment.Energy{GeneratedMWh: 30}},
		{Day: day("2026-10-17"), Subsystem: curtailment.South, Energy: curtailment.Energy{GeneratedMWh: -0.5}},
	}
	if !slices.Equal(m.SubsystemDays, wantSubsystems) {
		t.Errorf("subsystem days = %+v", m.SubsystemDays)
	}
	wantReasons := []curtailment.ReasonDay{
		{Day: day("2026-10-16"), Reason: curtailment.ReasonReliability, CurtailedMWh: 20},
		{Day: day("2026-10-16"), Reason: curtailment.ReasonExternal, CurtailedMWh: 10},
	}
	if !slices.Equal(m.ReasonDays, wantReasons) {
		t.Errorf("reason days = %+v", m.ReasonDays)
	}
	wantHalfHours := []curtailment.HalfHour{
		{Start: at(13, 0, 16), CurtailedMWh: 20},
		{Start: at(13, 30, 16), CurtailedMWh: 10},
		{Start: at(0, 0, 17), CurtailedMWh: 0},
	}
	if !slices.EqualFunc(m.HalfHours, wantHalfHours, func(a, b curtailment.HalfHour) bool {
		return a.Start.Equal(b.Start) && a.CurtailedMWh == b.CurtailedMWh
	}) {
		t.Errorf("half hours = %+v", m.HalfHours)
	}
}

func TestMonthBuilderRefusesWhatTheONSCouldNotHaveMeant(t *testing.T) {
	good := curtailment.Record{Plant: "CJU_A", Subsystem: curtailment.Northeast, At: at(13, 0, 16), GeneratedMW: 10}
	for name, change := range map[string]func(*curtailment.Record){
		"no plant":          func(r *curtailment.Record) { r.Plant = "" },
		"unknown subsystem": func(r *curtailment.Record) { r.Subsystem = "SUL" },
		"unknown reason":    func(r *curtailment.Record) { r.Reason = "XYZ" },
		"another month":     func(r *curtailment.Record) { r.At = time.Date(2026, 9, 30, 23, 30, 0, 0, curtailment.Brasilia) },
		"the next month":    func(r *curtailment.Record) { r.At = time.Date(2026, 11, 1, 0, 0, 0, 0, curtailment.Brasilia) },
		"off the half hour": func(r *curtailment.Record) { r.At = at(13, 10, 16) },
		"curtailed below 0": func(r *curtailment.Record) { r.CurtailedMW = -1 },
		"not a number":      func(r *curtailment.Record) { r.GeneratedMW = math.NaN() },
	} {
		r := good
		change(&r)
		b := curtailment.NewMonthBuilder(curtailment.Wind, day("2026-10-01"))
		if err := b.Add(r); !errors.Is(err, curtailment.ErrBadRecord) {
			t.Errorf("%s: Add() = %v, want ErrBadRecord", name, err)
		}
	}
	b := curtailment.NewMonthBuilder(curtailment.Wind, day("2026-10-01"))
	if err := b.Add(good); err != nil {
		t.Errorf("a good record: %v", err)
	}
}
