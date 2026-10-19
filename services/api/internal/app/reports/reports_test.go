package reports_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/app/reports"
	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

func day(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }

var bounds = curtailment.Bounds{First: time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC), Latest: time.Date(2026, 10, 16, 23, 30, 0, 0, curtailment.Brasilia)}

func TestOverviewReadsThePeriodItWorksOut(t *testing.T) {
	store := &fakeStore{bounds: bounds}
	o, err := reports.New(store).Overview(t.Context(), curtailment.Weekly, day(10, 14))
	if err != nil {
		t.Fatal(err)
	}
	if !store.asked.Start.Equal(day(10, 12)) || !store.asked.Through.Equal(day(10, 16)) {
		t.Errorf("read %+v, want the week from Monday the 12th to the 16th", store.asked)
	}
	if o.Period != store.asked || !o.Latest.Equal(bounds.Latest) {
		t.Errorf("Overview() = %+v", o)
	}
}

func TestReportsPassOnWhyTheyCannotAnswer(t *testing.T) {
	svc := reports.New(&fakeStore{boundsErr: curtailment.ErrNoData})
	if _, err := svc.Overview(t.Context(), curtailment.Daily, time.Time{}); !errors.Is(err, curtailment.ErrNoData) {
		t.Errorf("Overview() with no data = %v", err)
	}
	if _, err := svc.History(t.Context()); !errors.Is(err, curtailment.ErrNoData) {
		t.Errorf("History() with no data = %v", err)
	}

	svc = reports.New(&fakeStore{bounds: bounds})
	if _, err := svc.Plants(t.Context(), curtailment.Daily, day(10, 17), curtailment.PlantFilter{}, reports.Page{Limit: 20}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Plants() for a day not published = %v", err)
	}
	var invalid *domain.ValidationError
	if _, err := svc.Recent(t.Context(), reports.MaxRecentDays+1); !errors.As(err, &invalid) {
		t.Errorf("Recent(%d) = %v", reports.MaxRecentDays+1, err)
	}
}

func TestRecentFillsEveryDayOfEverySourceAndSubsystem(t *testing.T) {
	store := &fakeStore{bounds: bounds, days: map[curtailment.Source][]curtailment.SubsystemDay{
		curtailment.Wind: {
			{Day: day(10, 14), Subsystem: curtailment.Northeast, Energy: curtailment.Energy{CurtailedMWh: 150_000}},
			{Day: day(10, 16), Subsystem: curtailment.Northeast, Energy: curtailment.Energy{CurtailedMWh: 7_400}},
		},
	}}

	r, err := reports.New(store).Recent(t.Context(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if !r.From.Equal(day(10, 14)) || !r.To.Equal(day(10, 16)) || len(r.Series) != 8 {
		t.Fatalf("Recent() = %+v", r)
	}
	for _, s := range r.Series {
		want := []float64{0, 0, 0}
		if s.Source == curtailment.Wind && s.Subsystem == curtailment.Northeast {
			want = []float64{150_000, 0, 7_400}
		}
		if !slices.Equal(s.CurtailedMWh, want) {
			t.Errorf("%s %s = %v, want %v", s.Source, s.Subsystem, s.CurtailedMWh, want)
		}
	}
}

type fakeStore struct {
	bounds    curtailment.Bounds
	boundsErr error
	asked     curtailment.Period
	days      map[curtailment.Source][]curtailment.SubsystemDay
}

func (s *fakeStore) Bounds(context.Context) (curtailment.Bounds, error) {
	return s.bounds, s.boundsErr
}

func (s *fakeStore) Overview(_ context.Context, p curtailment.Period, _ time.Time) (curtailment.Overview, error) {
	s.asked = p
	return curtailment.Overview{}, nil
}

func (s *fakeStore) RankPlants(_ context.Context, p curtailment.Period, _ curtailment.PlantFilter, _ reports.Page) (curtailment.PlantRanking, error) {
	s.asked = p
	return curtailment.PlantRanking{}, nil
}

func (s *fakeStore) Months(context.Context, time.Time) ([]curtailment.MonthTotal, error) {
	return nil, nil
}

func (s *fakeStore) SubsystemDays(_ context.Context, source curtailment.Source, _, _ time.Time) ([]curtailment.SubsystemDay, error) {
	return s.days[source], nil
}
