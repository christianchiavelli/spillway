package httpapi

import (
	"context"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/httpapi/oas"
	"github.com/christianchiavelli/spillway/services/api/internal/app/reports"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

// Reports is what the edge asks of the reports use case.
type Reports interface {
	Overview(ctx context.Context, kind curtailment.Kind, anchor time.Time) (curtailment.Overview, error)
	Plants(ctx context.Context, kind curtailment.Kind, anchor time.Time, f curtailment.PlantFilter, page reports.Page) (curtailment.PlantRanking, error)
	History(ctx context.Context) (curtailment.History, error)
	Recent(ctx context.Context, days int) (curtailment.Recent, error)
}

func (h *handler) GetCurtailment(ctx context.Context, params oas.GetCurtailmentParams) (*oas.Curtailment, error) {
	o, err := h.reports.Overview(ctx, kindOf(params.Period), anchorOf(params.Date))
	if err != nil {
		return nil, err
	}
	out := &oas.Curtailment{
		Period:           toPeriod(o.Period),
		PublishedThrough: o.Latest.UTC(),
		Wind:             toSourceEnergy(o.Wind),
		Solar:            toSourceEnergy(o.Solar),
		PlantsLimited:    o.PlantsLimited,
		PlantsReporting:  o.PlantsReporting,
		Series:           make([]oas.Slot, len(o.Series)),
		BySubsystem:      make([]oas.SubsystemSplit, len(o.BySubsystem)),
		ByReason:         make([]oas.ReasonSplit, len(o.ByReason)),
	}
	if o.Period.HasPrevious() {
		out.PreviousCurtailedMwh = oas.NewOptEnergy(oas.Energy(o.PreviousCurtailedMWh))
	}
	for i, s := range o.Series {
		out.Series[i] = oas.Slot{Start: s.Start.UTC(), WindMwh: oas.Energy(s.Curtailed.Wind), SolarMwh: oas.Energy(s.Curtailed.Solar)}
	}
	for i, s := range o.BySubsystem {
		out.BySubsystem[i] = oas.SubsystemSplit{Subsystem: oas.Subsystem(s.Subsystem), WindMwh: oas.Energy(s.Curtailed.Wind), SolarMwh: oas.Energy(s.Curtailed.Solar)}
	}
	for i, r := range o.ByReason {
		out.ByReason[i] = oas.ReasonSplit{Reason: oas.Reason(r.Reason), WindMwh: oas.Energy(r.Curtailed.Wind), SolarMwh: oas.Energy(r.Curtailed.Solar)}
	}
	return out, nil
}

func (h *handler) ListPlants(ctx context.Context, params oas.ListPlantsParams) (*oas.PlantRanking, error) {
	var f curtailment.PlantFilter
	if v, ok := params.Source.Get(); ok {
		f.Source = curtailment.Source(v)
	}
	if v, ok := params.Subsystem.Get(); ok {
		f.Subsystem = curtailment.Subsystem(v)
	}
	page := reports.Page{Limit: params.Limit.Or(20), Offset: params.Offset.Or(0)}
	r, err := h.reports.Plants(ctx, kindOf(params.Period), anchorOf(params.Date), f, page)
	if err != nil {
		return nil, err
	}
	out := &oas.PlantRanking{
		Period:           toPeriod(r.Period),
		PublishedThrough: r.Latest.UTC(),
		Reporting:        r.Reporting,
		Total:            r.Limited,
		Items:            make([]oas.PlantCurtailment, len(r.Items)),
	}
	for i, p := range r.Items {
		out.Items[i] = oas.PlantCurtailment{
			ID:             p.ID,
			Name:           p.Name,
			State:          p.State,
			Subsystem:      oas.Subsystem(p.Subsystem),
			Source:         oas.Source(p.Source),
			CurtailedMwh:   oas.Energy(p.CurtailedMWh),
			GeneratedMwh:   oas.Energy(p.GeneratedMWh),
			LimitedMinutes: int32(p.LimitedMinutes), //nolint:gosec // at most a month's minutes
		}
	}
	return out, nil
}

func (h *handler) ListMonthlyCurtailment(ctx context.Context) (*oas.MonthlyCurtailment, error) {
	history, err := h.reports.History(ctx)
	if err != nil {
		return nil, err
	}
	out := &oas.MonthlyCurtailment{PublishedThrough: history.Latest.UTC(), Months: make([]oas.MonthCurtailment, len(history.Months))}
	for i, m := range history.Months {
		out.Months[i] = oas.MonthCurtailment{Month: oas.Day(m.Start)}
		if m.Wind != nil {
			out.Months[i].Wind = oas.NewOptSourceEnergy(toSourceEnergy(*m.Wind))
		}
		if m.Solar != nil {
			out.Months[i].Solar = oas.NewOptSourceEnergy(toSourceEnergy(*m.Solar))
		}
	}
	return out, nil
}

func (h *handler) ListDailyCurtailment(ctx context.Context, params oas.ListDailyCurtailmentParams) (*oas.DailyCurtailment, error) {
	r, err := h.reports.Recent(ctx, params.Days.Or(30))
	if err != nil {
		return nil, err
	}
	out := &oas.DailyCurtailment{From: oas.Day(r.From), To: oas.Day(r.To), Series: make([]oas.SubsystemDays, len(r.Series))}
	for i, s := range r.Series {
		values := make([]oas.Energy, len(s.CurtailedMWh))
		for j, v := range s.CurtailedMWh {
			values[j] = oas.Energy(v)
		}
		out.Series[i] = oas.SubsystemDays{Source: oas.Source(s.Source), Subsystem: oas.Subsystem(s.Subsystem), CurtailedMwh: values}
	}
	return out, nil
}

func kindOf(p oas.OptPeriodKind) curtailment.Kind {
	return curtailment.Kind(p.Or(oas.PeriodKindDay))
}

// anchorOf is the day asked for, or zero for the latest published.
func anchorOf(d oas.OptDay) time.Time {
	if v, ok := d.Get(); ok {
		return time.Time(v)
	}
	return time.Time{}
}

func toPeriod(p curtailment.Period) oas.Period {
	return oas.Period{
		Kind:          oas.PeriodKind(p.Kind),
		Start:         oas.Day(p.Start),
		End:           oas.Day(p.End),
		Through:       oas.Day(p.Through),
		PreviousStart: optDay(p.PreviousStart),
		PreviousEnd:   optDay(p.PreviousEnd),
		Earlier:       optDay(p.Earlier),
		Later:         optDay(p.Later),
	}
}

func toSourceEnergy(e curtailment.Energy) oas.SourceEnergy {
	return oas.SourceEnergy{CurtailedMwh: oas.Energy(e.CurtailedMWh), GeneratedMwh: oas.Energy(e.GeneratedMWh)}
}

// optDay leaves the zero day out.
func optDay(t time.Time) oas.OptDay {
	if t.IsZero() {
		return oas.OptDay{}
	}
	return oas.NewOptDay(oas.Day(t))
}
