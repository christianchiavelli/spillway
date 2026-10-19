//go:build integration

package postgres_test

import (
	"errors"
	"slices"
	"testing"
	"time"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/postgres"
	"github.com/christianchiavelli/spillway/services/api/internal/app/ingest"
	"github.com/christianchiavelli/spillway/services/api/internal/app/reports"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

func oct(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }

func at(d, hour, minute int) time.Time {
	return time.Date(2026, 10, d, hour, minute, 0, 0, curtailment.Brasilia)
}

func monthOf(t *testing.T, source curtailment.Source, records ...curtailment.Record) curtailment.Month {
	t.Helper()
	b := curtailment.NewMonthBuilder(source, oct(1))
	for _, r := range records {
		if err := b.Add(r); err != nil {
			t.Fatal(err)
		}
	}
	return b.Month()
}

var (
	plantA = func(r curtailment.Record) curtailment.Record {
		r.Plant, r.PlantName, r.State, r.Subsystem = "CJU_A", "Conj. A", "RN", curtailment.Northeast
		return r
	}
	plantB = func(r curtailment.Record) curtailment.Record {
		r.Plant, r.PlantName, r.State, r.Subsystem = "CJU_B", "Conj. B", "RS", curtailment.South
		return r
	}
	plantC = func(r curtailment.Record) curtailment.Record {
		r.Plant, r.PlantName, r.State, r.Subsystem = "CJU_C", "Conj. C Solar", "BA", curtailment.Northeast
		return r
	}
	windFile  = ingest.File{Key: "wind/2026_10", Source: curtailment.Wind, Month: oct(1), ETag: "w1"}
	solarFile = ingest.File{Key: "solar/2026_10", Source: curtailment.Solar, Month: oct(1), ETag: "s1"}
)

// The curtailment tables are shared by every reader, so this test runs alone
// and in order: load, read, revise, read again.
func TestCurtailment(t *testing.T) {
	store := postgres.NewCurtailment(pool, time.Now)

	if _, err := store.Bounds(t.Context()); !errors.Is(err, curtailment.ErrNoData) {
		t.Fatalf("Bounds() before any file = %v, want ErrNoData", err)
	}

	wind := monthOf(t, curtailment.Wind,
		plantA(curtailment.Record{At: at(15, 13, 0), GeneratedMW: 200, CurtailedMW: 100, Reason: curtailment.ReasonReliability, LimitedMinutes: 30}),
		plantA(curtailment.Record{At: at(16, 13, 0), GeneratedMW: 200, CurtailedMW: 300, Reason: curtailment.ReasonReliability, LimitedMinutes: 30}),
		plantB(curtailment.Record{At: at(16, 14, 0), GeneratedMW: 50}),
		plantB(curtailment.Record{At: at(16, 23, 30), GeneratedMW: 40}),
	)
	solar := monthOf(t, curtailment.Solar,
		plantC(curtailment.Record{At: at(16, 12, 0), GeneratedMW: 60, CurtailedMW: 40, Reason: curtailment.ReasonExternal, LimitedMinutes: 20}),
		plantC(curtailment.Record{At: at(16, 23, 30)}),
	)
	for _, load := range []struct {
		f ingest.File
		m curtailment.Month
	}{{windFile, wind}, {solarFile, solar}} {
		if err := store.Replace(t.Context(), load.f, load.m); err != nil {
			t.Fatal(err)
		}
	}

	versions, err := store.Versions(t.Context())
	if err != nil || versions[windFile.Key] != "w1" || versions[solarFile.Key] != "s1" {
		t.Errorf("Versions() = %v, %v", versions, err)
	}
	b, err := store.Bounds(t.Context())
	if err != nil || !b.First.Equal(oct(1)) || !b.Latest.Equal(at(16, 23, 30)) {
		t.Fatalf("Bounds() = %+v, %v", b, err)
	}

	svc := reports.New(store)

	t.Run("a day", func(t *testing.T) {
		o, err := svc.Overview(t.Context(), curtailment.Daily, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		if o.Wind != (curtailment.Energy{CurtailedMWh: 150, GeneratedMWh: 145}) || o.Solar != (curtailment.Energy{CurtailedMWh: 20, GeneratedMWh: 30}) {
			t.Errorf("totals = %+v %+v", o.Wind, o.Solar)
		}
		if o.PreviousCurtailedMWh != 50 || o.PlantsLimited != 2 || o.PlantsReporting != 3 {
			t.Errorf("previous %v, limited %d of %d", o.PreviousCurtailedMWh, o.PlantsLimited, o.PlantsReporting)
		}
		want := []curtailment.Slot{
			{Start: at(16, 12, 0), Curtailed: curtailment.Split{Solar: 20}},
			{Start: at(16, 13, 0), Curtailed: curtailment.Split{Wind: 150}},
			{Start: at(16, 14, 0)},
			{Start: at(16, 23, 30)},
		}
		if !slices.EqualFunc(o.Series, want, func(a, b curtailment.Slot) bool { return a.Start.Equal(b.Start) && a.Curtailed == b.Curtailed }) {
			t.Errorf("series = %+v", o.Series)
		}
		wantSubsystems := []curtailment.SubsystemSplit{
			{Subsystem: curtailment.North},
			{Subsystem: curtailment.Northeast, Curtailed: curtailment.Split{Wind: 150, Solar: 20}},
			{Subsystem: curtailment.Southeast},
			{Subsystem: curtailment.South},
		}
		if !slices.Equal(o.BySubsystem, wantSubsystems) {
			t.Errorf("by subsystem = %+v", o.BySubsystem)
		}
		reasons := map[curtailment.Reason]curtailment.Split{}
		for _, r := range o.ByReason {
			reasons[r.Reason] = r.Curtailed
		}
		if len(reasons) != 2 || reasons[curtailment.ReasonReliability].Wind != 150 || reasons[curtailment.ReasonExternal].Solar != 20 {
			t.Errorf("by reason = %+v", o.ByReason)
		}
	})

	t.Run("a week so far", func(t *testing.T) {
		o, err := svc.Overview(t.Context(), curtailment.Weekly, oct(14))
		if err != nil {
			t.Fatal(err)
		}
		if o.Wind.CurtailedMWh != 200 || len(o.Series) != 2 || !o.Series[0].Start.Equal(curtailment.StartOf(oct(15))) {
			t.Errorf("wind %v, series %+v", o.Wind.CurtailedMWh, o.Series)
		}
		if !o.Period.HasPrevious() || o.PreviousCurtailedMWh != 0 {
			t.Errorf("the week before, with no data, = %v", o.PreviousCurtailedMWh)
		}
	})

	t.Run("the plants ranked", func(t *testing.T) {
		r, err := svc.Plants(t.Context(), curtailment.Weekly, oct(16), curtailment.PlantFilter{}, reports.Page{Limit: 20})
		if err != nil {
			t.Fatal(err)
		}
		if r.Reporting != 3 || r.Limited != 2 || len(r.Items) != 2 || r.Items[0].ID != "CJU_A" || r.Items[0].CurtailedMWh != 200 || r.Items[0].LimitedMinutes != 60 {
			t.Errorf("Plants() = %+v", r)
		}
		r, _ = svc.Plants(t.Context(), curtailment.Weekly, oct(16), curtailment.PlantFilter{Source: curtailment.Solar}, reports.Page{Limit: 20})
		if r.Reporting != 1 || len(r.Items) != 1 || r.Items[0].Name != "Conj. C Solar" {
			t.Errorf("solar only = %+v", r)
		}
		r, _ = svc.Plants(t.Context(), curtailment.Weekly, oct(16), curtailment.PlantFilter{Subsystem: curtailment.South}, reports.Page{Limit: 20})
		if r.Reporting != 1 || r.Limited != 0 || len(r.Items) != 0 {
			t.Errorf("the South, which lost nothing = %+v", r)
		}
		r, _ = svc.Plants(t.Context(), curtailment.Weekly, oct(16), curtailment.PlantFilter{}, reports.Page{Limit: 1, Offset: 1})
		if r.Limited != 2 || len(r.Items) != 1 || r.Items[0].ID != "CJU_C" {
			t.Errorf("the second page of one = %+v", r)
		}
	})

	t.Run("the months and the latest days", func(t *testing.T) {
		h, err := svc.History(t.Context())
		if err != nil || len(h.Months) != 1 || h.Months[0].Wind.CurtailedMWh != 200 || h.Months[0].Solar.CurtailedMWh != 20 {
			t.Errorf("History() = %+v, %v", h, err)
		}
		r, err := svc.Recent(t.Context(), 2)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range r.Series {
			if s.Source == curtailment.Wind && s.Subsystem == curtailment.Northeast && !slices.Equal(s.CurtailedMWh, []float64{50, 150}) {
				t.Errorf("wind in the Northeast = %v", s.CurtailedMWh)
			}
		}
	})

	t.Run("alerts fire only over their threshold", func(t *testing.T) {
		alerts := postgres.NewAlerts(pool)
		owner, other := newAccount(t), newAccount(t)
		over := watch(t, alerts, owner.ID, "Wind over 100", curtailment.Wind, 100)
		at150 := watch(t, alerts, owner.ID, "Wind at 150", curtailment.Wind, 150)
		solarOver := watch(t, alerts, owner.ID, "Solar over 10", curtailment.Solar, 10)
		watch(t, alerts, other.ID, "Someone else's", curtailment.Wind, 1)

		activity, err := alerts.Activity(t.Context(), owner.ID, oct(16).AddDate(0, 0, 1-alert.RecentDays), oct(16))
		if err != nil {
			t.Fatal(err)
		}
		if activity[over.ID] != (alert.Activity{LastFiredOn: oct(16), DaysFired: 1}) || activity[solarOver.ID].DaysFired != 1 {
			t.Errorf("activity = %+v", activity)
		}
		if _, fired := activity[at150.ID]; fired {
			t.Error("a day that lost exactly the threshold set the alert off")
		}

		firings, total, err := alerts.Firings(t.Context(), owner.ID, oct(1), oct(16), 10, 0)
		if err != nil {
			t.Fatal(err)
		}
		names := make([]string, len(firings))
		for i, f := range firings {
			names[i] = f.AlertName
		}
		// On the same day, the one furthest over its threshold first: 20 is
		// twice 10, and 150 is one and a half times 100.
		if total != 2 || !slices.Equal(names, []string{"Solar over 10", "Wind over 100"}) || firings[1].CurtailedMWh != 150 {
			t.Errorf("Firings() = %+v, total %d", firings, total)
		}
		if page, total, _ := alerts.Firings(t.Context(), owner.ID, oct(1), oct(16), 10, 5); len(page) != 0 || total != 2 {
			t.Errorf("a page past the end = %+v, total %d", page, total)
		}
	})

	t.Run("a revised file replaces its month", func(t *testing.T) {
		revised := monthOf(t, curtailment.Wind,
			plantA(curtailment.Record{At: at(16, 13, 0), GeneratedMW: 280, CurtailedMW: 20, Reason: curtailment.ReasonEnergy, LimitedMinutes: 10}),
			plantA(curtailment.Record{At: at(16, 23, 30), GeneratedMW: 100}),
		)
		f := windFile
		f.ETag = "w2"
		if err := store.Replace(t.Context(), f, revised); err != nil {
			t.Fatal(err)
		}

		o, err := svc.Overview(t.Context(), curtailment.Daily, oct(16))
		if err != nil {
			t.Fatal(err)
		}
		if o.Wind.CurtailedMWh != 10 || o.PlantsReporting != 2 || o.PreviousCurtailedMWh != 0 {
			t.Errorf("after the revision: wind %v, %d reporting, previous %v", o.Wind.CurtailedMWh, o.PlantsReporting, o.PreviousCurtailedMWh)
		}
		if len(o.ByReason) != 2 {
			t.Errorf("reasons after the revision = %+v", o.ByReason)
		}
		if versions, _ := store.Versions(t.Context()); versions[windFile.Key] != "w2" {
			t.Errorf("the ETag recorded = %q", versions[windFile.Key])
		}
	})
}

func watch(t *testing.T, alerts *postgres.Alerts, owner uuid.UUID, name string, source curtailment.Source, threshold int32) alert.Alert {
	t.Helper()
	a, err := alert.New(uuid.NewV7(), owner, alert.Draft{Name: name, Source: string(source), Subsystem: "NE", ThresholdMWh: threshold}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := alerts.Add(t.Context(), a, alert.MaxPerAccount); err != nil {
		t.Fatal(err)
	}
	return a
}
