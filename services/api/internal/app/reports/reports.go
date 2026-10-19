// Package reports is the use case of reading the curtailment Spillway keeps:
// a period's overview, the plants ranked by what they lost, every month since
// the ONS began, and the latest days by source and subsystem.
package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

// Store reads the sums the ingest use case keeps.
type Store interface {
	// Bounds says how far the data goes, or returns curtailment.ErrNoData.
	Bounds(ctx context.Context) (curtailment.Bounds, error)
	// Overview fills in everything of p's overview but the period and the
	// bounds, from one snapshot of the data.
	Overview(ctx context.Context, p curtailment.Period, latest time.Time) (curtailment.Overview, error)
	// RankPlants returns a page of the plants that lost energy in p, most
	// first, and counts those that reported and those that lost any.
	RankPlants(ctx context.Context, p curtailment.Period, f curtailment.PlantFilter, page Page) (curtailment.PlantRanking, error)
	// Months returns every month up to through, oldest first.
	Months(ctx context.Context, through time.Time) ([]curtailment.MonthTotal, error)
	// SubsystemDays returns the days from from to through, by subsystem.
	SubsystemDays(ctx context.Context, source curtailment.Source, from, through time.Time) ([]curtailment.SubsystemDay, error)
}

// Page is a slice of a long list.
type Page struct {
	Limit  int
	Offset int
}

// Service runs the use case.
type Service struct {
	store Store
}

// New returns a Service.
func New(store Store) *Service {
	return &Service{store: store}
}

// Overview returns the period of kind that holds anchor, a day; the zero
// anchor is the latest day published.
func (s *Service) Overview(ctx context.Context, kind curtailment.Kind, anchor time.Time) (curtailment.Overview, error) {
	b, p, err := s.period(ctx, kind, anchor)
	if err != nil {
		return curtailment.Overview{}, err
	}
	o, err := s.store.Overview(ctx, p, b.Latest)
	if err != nil {
		return curtailment.Overview{}, fmt.Errorf("read the overview: %w", err)
	}
	o.Period, o.Latest = p, b.Latest
	return o, nil
}

// Plants returns a page of the plants that lost energy in the period, most
// first.
func (s *Service) Plants(ctx context.Context, kind curtailment.Kind, anchor time.Time, f curtailment.PlantFilter, page Page) (curtailment.PlantRanking, error) {
	b, p, err := s.period(ctx, kind, anchor)
	if err != nil {
		return curtailment.PlantRanking{}, err
	}
	r, err := s.store.RankPlants(ctx, p, f, page)
	if err != nil {
		return curtailment.PlantRanking{}, fmt.Errorf("rank the plants: %w", err)
	}
	r.Period, r.Latest = p, b.Latest
	return r, nil
}

// History returns every month the ONS has published.
func (s *Service) History(ctx context.Context) (curtailment.History, error) {
	b, err := s.store.Bounds(ctx)
	if err != nil {
		return curtailment.History{}, fmt.Errorf("read the bounds: %w", err)
	}
	months, err := s.store.Months(ctx, b.LastDay())
	if err != nil {
		return curtailment.History{}, fmt.Errorf("read the months: %w", err)
	}
	return curtailment.History{Latest: b.Latest, Months: months}, nil
}

// MaxRecentDays is the most days Recent returns.
const MaxRecentDays = 90

// Recent returns the last days published, by source and subsystem: every
// pair, with a zero for each day one lost nothing.
func (s *Service) Recent(ctx context.Context, days int) (curtailment.Recent, error) {
	if days < 1 || days > MaxRecentDays {
		var vs domain.Violations
		vs.Add("days", domain.CodeOutOfRange)
		return curtailment.Recent{}, vs.Err()
	}
	b, err := s.store.Bounds(ctx)
	if err != nil {
		return curtailment.Recent{}, fmt.Errorf("read the bounds: %w", err)
	}
	to := b.LastDay()
	from := to.AddDate(0, 0, 1-days)
	r := curtailment.Recent{From: from, To: to}
	for _, source := range curtailment.Sources() {
		rows, err := s.store.SubsystemDays(ctx, source, from, to)
		if err != nil {
			return curtailment.Recent{}, fmt.Errorf("read the days: %w", err)
		}
		series := map[curtailment.Subsystem][]float64{}
		for _, sub := range curtailment.Subsystems() {
			series[sub] = make([]float64, days)
		}
		for _, row := range rows {
			values, ok := series[row.Subsystem]
			if i := curtailment.DaysBetween(from, row.Day); ok && i >= 0 && i < days {
				values[i] = row.CurtailedMWh
			}
		}
		for _, sub := range curtailment.Subsystems() {
			r.Series = append(r.Series, curtailment.SubsystemSeries{Source: source, Subsystem: sub, CurtailedMWh: series[sub]})
		}
	}
	return r, nil
}

func (s *Service) period(ctx context.Context, kind curtailment.Kind, anchor time.Time) (curtailment.Bounds, curtailment.Period, error) {
	b, err := s.store.Bounds(ctx)
	if err != nil {
		return curtailment.Bounds{}, curtailment.Period{}, fmt.Errorf("read the bounds: %w", err)
	}
	p, err := curtailment.NewPeriod(kind, anchor, b)
	if err != nil {
		return curtailment.Bounds{}, curtailment.Period{}, err
	}
	return b, p, nil
}
