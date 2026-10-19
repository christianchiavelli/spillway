package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/postgres/sqlcdb"
	"github.com/christianchiavelli/spillway/services/api/internal/app/ingest"
	"github.com/christianchiavelli/spillway/services/api/internal/app/reports"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

// Curtailment keeps the sums of the ONS's files. It implements ingest.Store,
// reports.Store and alerting.Calendar.
type Curtailment struct {
	pool *pgxpool.Pool
	q    *sqlcdb.Queries
	now  func() time.Time
}

// NewCurtailment returns Curtailment on pool.
func NewCurtailment(pool *pgxpool.Pool, now func() time.Time) *Curtailment {
	return &Curtailment{pool: pool, q: sqlcdb.New(pool), now: now}
}

// Versions returns the ETag of every file loaded, by key.
func (s *Curtailment) Versions(ctx context.Context) (map[string]string, error) {
	rows, err := s.q.FileVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("select file versions: %w", err)
	}
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		out[r.Key] = r.Etag
	}
	return out, nil
}

// Replace swaps a source's month for m in one transaction. The plants go
// first, since the days reference them; the rest is deleted and copied in.
func (s *Curtailment) Replace(ctx context.Context, f ingest.File, m curtailment.Month) error {
	source := string(m.Source)
	next := m.Start.AddDate(0, 1, 0)
	return pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)

		plants := make([]sqlcdb.UpsertPlantParams, len(m.Plants))
		for i, p := range m.Plants {
			plants[i] = sqlcdb.UpsertPlantParams{ID: p.ID, Name: p.Name, State: p.State, Subsystem: string(p.Subsystem), Source: string(p.Source), LastSeen: p.LastSeen}
		}
		var upsertErr error
		q.UpsertPlant(ctx, plants).Exec(func(_ int, err error) {
			if err != nil && upsertErr == nil {
				upsertErr = err
			}
		})
		if upsertErr != nil {
			return fmt.Errorf("upsert plants: %w", upsertErr)
		}

		days := sqlcdb.DeletePlantDaysParams{Source: source, FirstDay: m.Start, NextMonth: next}
		if err := q.DeletePlantDays(ctx, days); err != nil {
			return fmt.Errorf("delete plant days: %w", err)
		}
		if err := q.DeleteSubsystemDays(ctx, sqlcdb.DeleteSubsystemDaysParams(days)); err != nil {
			return fmt.Errorf("delete subsystem days: %w", err)
		}
		if err := q.DeleteReasonDays(ctx, sqlcdb.DeleteReasonDaysParams(days)); err != nil {
			return fmt.Errorf("delete reason days: %w", err)
		}
		halves := sqlcdb.DeleteHalfHoursParams{Source: source, FromInstant: curtailment.StartOf(m.Start), UntilInstant: curtailment.StartOf(next)}
		if err := q.DeleteHalfHours(ctx, halves); err != nil {
			return fmt.Errorf("delete half hours: %w", err)
		}

		plantDays := make([]sqlcdb.CopyPlantDaysParams, len(m.PlantDays))
		for i, d := range m.PlantDays {
			plantDays[i] = sqlcdb.CopyPlantDaysParams{
				Day: d.Day, PlantID: d.Plant, Source: source, Subsystem: string(d.Subsystem),
				CurtailedMwh: d.CurtailedMWh, GeneratedMwh: d.GeneratedMWh, LimitedMinutes: int32(d.LimitedMinutes), //nolint:gosec // at most 30 a half hour, 1,488 a day
			}
		}
		if _, err := q.CopyPlantDays(ctx, plantDays); err != nil {
			return fmt.Errorf("copy plant days: %w", err)
		}
		subsystemDays := make([]sqlcdb.CopySubsystemDaysParams, len(m.SubsystemDays))
		for i, d := range m.SubsystemDays {
			subsystemDays[i] = sqlcdb.CopySubsystemDaysParams{Source: source, Subsystem: string(d.Subsystem), Day: d.Day, CurtailedMwh: d.CurtailedMWh, GeneratedMwh: d.GeneratedMWh}
		}
		if _, err := q.CopySubsystemDays(ctx, subsystemDays); err != nil {
			return fmt.Errorf("copy subsystem days: %w", err)
		}
		reasonDays := make([]sqlcdb.CopyReasonDaysParams, len(m.ReasonDays))
		for i, d := range m.ReasonDays {
			reasonDays[i] = sqlcdb.CopyReasonDaysParams{Day: d.Day, Source: source, Reason: string(d.Reason), CurtailedMwh: d.CurtailedMWh}
		}
		if _, err := q.CopyReasonDays(ctx, reasonDays); err != nil {
			return fmt.Errorf("copy reason days: %w", err)
		}
		halfHours := make([]sqlcdb.CopyHalfHoursParams, len(m.HalfHours))
		for i, h := range m.HalfHours {
			halfHours[i] = sqlcdb.CopyHalfHoursParams{StartsAt: h.Start, Source: source, CurtailedMwh: h.CurtailedMWh}
		}
		if _, err := q.CopyHalfHours(ctx, halfHours); err != nil {
			return fmt.Errorf("copy half hours: %w", err)
		}

		err := q.RecordFile(ctx, sqlcdb.RecordFileParams{
			Key: f.Key, Source: source, Month: m.Start, Etag: f.ETag,
			Records:     int32(m.Records), //nolint:gosec // a month holds about 250,000
			LastInstant: pgtype.Timestamptz{Time: m.Latest, Valid: !m.Latest.IsZero()},
			LoadedAt:    s.now(),
		})
		if err != nil {
			return fmt.Errorf("record file: %w", err)
		}
		return nil
	})
}

// Bounds says how far the data goes, or returns curtailment.ErrNoData.
func (s *Curtailment) Bounds(ctx context.Context) (curtailment.Bounds, error) {
	b, err := s.q.Bounds(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return curtailment.Bounds{}, curtailment.ErrNoData
	case err != nil:
		return curtailment.Bounds{}, fmt.Errorf("select bounds: %w", err)
	}
	return curtailment.Bounds{First: b.FirstDay, Latest: b.Latest.UTC()}, nil
}

// Overview reads a period from one snapshot of the data, so a file replaced
// between two of its queries cannot make the totals and the chart disagree.
func (s *Curtailment) Overview(ctx context.Context, p curtailment.Period, latest time.Time) (curtailment.Overview, error) {
	var o curtailment.Overview
	err := pgx.BeginTxFunc(ctx, s.pool, snapshot, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		span := sqlcdb.SourceTotalsParams{FirstDay: p.Start, LastDay: p.Through}

		totals, err := q.SourceTotals(ctx, span)
		if err != nil {
			return fmt.Errorf("select source totals: %w", err)
		}
		for _, t := range totals {
			e := curtailment.Energy{CurtailedMWh: t.CurtailedMwh, GeneratedMWh: t.GeneratedMwh}
			switch curtailment.Source(t.Source) {
			case curtailment.Wind:
				o.Wind = e
			case curtailment.Solar:
				o.Solar = e
			}
		}

		if p.HasPrevious() {
			o.PreviousCurtailedMWh, err = q.CurtailedBetween(ctx, sqlcdb.CurtailedBetweenParams{FirstDay: p.PreviousStart, LastDay: p.PreviousEnd})
			if err != nil {
				return fmt.Errorf("select the previous span: %w", err)
			}
		}

		counts, err := q.PlantCounts(ctx, sqlcdb.PlantCountsParams{FirstDay: p.Start, LastDay: p.Through})
		if err != nil {
			return fmt.Errorf("count plants: %w", err)
		}
		o.PlantsLimited, o.PlantsReporting = int(counts.Limited), int(counts.Reporting)

		if o.Series, err = s.series(ctx, q, p, latest); err != nil {
			return err
		}

		subsystems, err := q.SubsystemTotals(ctx, sqlcdb.SubsystemTotalsParams(span))
		if err != nil {
			return fmt.Errorf("select subsystem totals: %w", err)
		}
		bySubsystem := map[curtailment.Subsystem]int{}
		for i, sub := range curtailment.Subsystems() {
			o.BySubsystem = append(o.BySubsystem, curtailment.SubsystemSplit{Subsystem: sub})
			bySubsystem[sub] = i
		}
		for _, r := range subsystems {
			if i, ok := bySubsystem[curtailment.Subsystem(r.Subsystem)]; ok {
				o.BySubsystem[i].Curtailed.Add(curtailment.Source(r.Source), r.CurtailedMwh)
			}
		}

		reasons, err := q.ReasonTotals(ctx, sqlcdb.ReasonTotalsParams(span))
		if err != nil {
			return fmt.Errorf("select reason totals: %w", err)
		}
		byReason := map[curtailment.Reason]int{}
		for _, r := range reasons {
			reason := curtailment.Reason(r.Reason)
			i, ok := byReason[reason]
			if !ok {
				i = len(o.ByReason)
				byReason[reason] = i
				o.ByReason = append(o.ByReason, curtailment.ReasonSplit{Reason: reason})
			}
			o.ByReason[i].Curtailed.Add(curtailment.Source(r.Source), r.CurtailedMwh)
		}
		return nil
	})
	return o, err
}

// series is a day's half hours, or a longer period's days.
func (s *Curtailment) series(ctx context.Context, q *sqlcdb.Queries, p curtailment.Period, latest time.Time) ([]curtailment.Slot, error) {
	var out []curtailment.Slot
	at := map[int64]int{}
	add := func(start time.Time, source string, mwh float64) {
		i, ok := at[start.Unix()]
		if !ok {
			i = len(out)
			at[start.Unix()] = i
			out = append(out, curtailment.Slot{Start: start.UTC()})
		}
		out[i].Curtailed.Add(curtailment.Source(source), mwh)
	}

	if p.Kind == curtailment.Daily {
		rows, err := q.HalfHourSeries(ctx, sqlcdb.HalfHourSeriesParams{
			FromInstant: curtailment.StartOf(p.Start), UntilInstant: curtailment.StartOf(p.End.AddDate(0, 0, 1)), Latest: latest,
		})
		if err != nil {
			return nil, fmt.Errorf("select half hours: %w", err)
		}
		for _, r := range rows {
			add(r.StartsAt, r.Source, r.CurtailedMwh)
		}
		return out, nil
	}
	rows, err := q.DailySeries(ctx, sqlcdb.DailySeriesParams{FirstDay: p.Start, LastDay: p.Through})
	if err != nil {
		return nil, fmt.Errorf("select days: %w", err)
	}
	for _, r := range rows {
		add(curtailment.StartOf(r.Day), r.Source, r.CurtailedMwh)
	}
	return out, nil
}

// RankPlants returns a page of the plants that lost energy in p, most first,
// from one snapshot of the data.
func (s *Curtailment) RankPlants(ctx context.Context, p curtailment.Period, f curtailment.PlantFilter, page reports.Page) (curtailment.PlantRanking, error) {
	source := pgtype.Text{String: string(f.Source), Valid: f.Source != ""}
	subsystem := pgtype.Text{String: string(f.Subsystem), Valid: f.Subsystem != ""}
	var r curtailment.PlantRanking
	err := pgx.BeginTxFunc(ctx, s.pool, snapshot, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		counts, err := q.PlantCounts(ctx, sqlcdb.PlantCountsParams{FirstDay: p.Start, LastDay: p.Through, Source: source, Subsystem: subsystem})
		if err != nil {
			return fmt.Errorf("count plants: %w", err)
		}
		r.Reporting, r.Limited = int(counts.Reporting), int(counts.Limited)

		rows, err := q.RankPlants(ctx, sqlcdb.RankPlantsParams{
			FirstDay: p.Start, LastDay: p.Through, Source: source, Subsystem: subsystem,
			RowLimit: int32(page.Limit), RowOffset: int32(page.Offset), //nolint:gosec // the contract caps both
		})
		if err != nil {
			return fmt.Errorf("rank plants: %w", err)
		}
		r.Items = make([]curtailment.PlantTotal, len(rows))
		for i, row := range rows {
			r.Items[i] = curtailment.PlantTotal{
				Plant: curtailment.Plant{
					ID: row.ID, Name: row.Name, State: row.State,
					Subsystem: curtailment.Subsystem(row.Subsystem), Source: curtailment.Source(row.Source),
				},
				Energy:         curtailment.Energy{CurtailedMWh: row.CurtailedMwh, GeneratedMWh: row.GeneratedMwh},
				LimitedMinutes: int(row.LimitedMinutes),
			}
		}
		return nil
	})
	return r, err
}

// Months returns every month up to through, oldest first.
func (s *Curtailment) Months(ctx context.Context, through time.Time) ([]curtailment.MonthTotal, error) {
	rows, err := s.q.MonthlyTotals(ctx, through)
	if err != nil {
		return nil, fmt.Errorf("select months: %w", err)
	}
	var out []curtailment.MonthTotal
	for _, r := range rows {
		if len(out) == 0 || !out[len(out)-1].Start.Equal(r.Month) {
			out = append(out, curtailment.MonthTotal{Start: r.Month})
		}
		e := &curtailment.Energy{CurtailedMWh: r.CurtailedMwh, GeneratedMWh: r.GeneratedMwh}
		switch curtailment.Source(r.Source) {
		case curtailment.Wind:
			out[len(out)-1].Wind = e
		case curtailment.Solar:
			out[len(out)-1].Solar = e
		}
	}
	return out, nil
}

// SubsystemDays returns a source's days from from to through, by subsystem.
func (s *Curtailment) SubsystemDays(ctx context.Context, source curtailment.Source, from, through time.Time) ([]curtailment.SubsystemDay, error) {
	rows, err := s.q.SubsystemDaysBetween(ctx, sqlcdb.SubsystemDaysBetweenParams{Source: string(source), FirstDay: from, LastDay: through})
	if err != nil {
		return nil, fmt.Errorf("select subsystem days: %w", err)
	}
	out := make([]curtailment.SubsystemDay, len(rows))
	for i, r := range rows {
		out[i] = curtailment.SubsystemDay{
			Day: r.Day, Subsystem: curtailment.Subsystem(r.Subsystem),
			Energy: curtailment.Energy{CurtailedMWh: r.CurtailedMwh, GeneratedMWh: r.GeneratedMwh},
		}
	}
	return out, nil
}

// snapshot runs a read's queries against one view of the data.
var snapshot = pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}
