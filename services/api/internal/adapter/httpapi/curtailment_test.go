package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/app/alerting"
	"github.com/christianchiavelli/spillway/services/api/internal/app/reports"
	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

var (
	latest = time.Date(2026, 10, 16, 23, 30, 0, 0, curtailment.Brasilia)
	friday = time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
)

func TestCurtailmentIsPublic(t *testing.T) {
	var asked struct {
		kind   curtailment.Kind
		anchor time.Time
	}
	r := &fakeReports{overview: func(_ context.Context, kind curtailment.Kind, anchor time.Time) (curtailment.Overview, error) {
		asked.kind, asked.anchor = kind, anchor
		return curtailment.Overview{
			Period: curtailment.Period{Kind: kind, Start: friday, End: friday, Through: friday, Earlier: friday.AddDate(0, 0, -1)},
			Latest: latest,
			Wind:   curtailment.Energy{CurtailedMWh: 7400, GeneratedMWh: 200000},
			Series: []curtailment.Slot{{Start: curtailment.StartOf(friday), Curtailed: curtailment.Split{Wind: 12.5}}},
			BySubsystem: []curtailment.SubsystemSplit{
				{Subsystem: curtailment.Northeast, Curtailed: curtailment.Split{Wind: 7400}},
			},
			ByReason: []curtailment.ReasonSplit{{Reason: curtailment.ReasonReliability, Curtailed: curtailment.Split{Wind: 7400}}},
		}, nil
	}}
	srv := newReportServer(t, r)

	res, body := do(t, srv, "GET", "/v1/curtailment", "")

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %v", res.StatusCode, body)
	}
	if asked.kind != curtailment.Daily || !asked.anchor.IsZero() {
		t.Errorf("asked for %q at %v, want the latest day", asked.kind, asked.anchor)
	}
	period := body["period"].(map[string]any)
	if period["kind"] != "day" || period["start"] != "2026-10-16" || period["earlier"] != "2026-10-15" {
		t.Errorf("period = %v", period)
	}
	for _, absent := range []string{"later", "previousStart", "previousEnd"} {
		if _, ok := period[absent]; ok {
			t.Errorf("period has %s: %v", absent, period)
		}
	}
	if _, ok := body["previousCurtailedMwh"]; ok {
		t.Error("a period with nothing to compare with has previousCurtailedMwh")
	}
	if body["publishedThrough"] != "2026-10-17T02:30:00Z" {
		t.Errorf("publishedThrough = %v, want the instant in UTC", body["publishedThrough"])
	}
	slot := body["series"].([]any)[0].(map[string]any)
	if slot["start"] != "2026-10-16T03:00:00Z" || slot["windMwh"] != 12.5 || slot["solarMwh"] != 0.0 {
		t.Errorf("series[0] = %v", slot)
	}
}

func TestCurtailmentReadsThePeriodAskedFor(t *testing.T) {
	var asked curtailment.Kind
	var anchor time.Time
	srv := newReportServer(t, &fakeReports{overview: func(_ context.Context, kind curtailment.Kind, a time.Time) (curtailment.Overview, error) {
		asked, anchor = kind, a
		return curtailment.Overview{Period: curtailment.Period{Kind: kind, Start: friday, End: friday, Through: friday}}, nil
	}})

	res, body := do(t, srv, "GET", "/v1/curtailment?period=week&date=2026-10-14", "")
	if res.StatusCode != http.StatusOK || asked != curtailment.Weekly || !anchor.Equal(time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("got %d %v, asked for %q at %v", res.StatusCode, body, asked, anchor)
	}

	for _, query := range []string{"period=year", "date=14-10-2026", "date=2026-02-30"} {
		res, body := do(t, srv, "GET", "/v1/curtailment?"+query, "")
		if res.StatusCode != http.StatusBadRequest || body["code"] != "malformed_request" {
			t.Errorf("%s: %d %v", query, res.StatusCode, body)
		}
	}
}

func TestCurtailmentErrors(t *testing.T) {
	for _, tt := range []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"no data loaded yet", wrap(curtailment.ErrNoData), http.StatusServiceUnavailable, "data_unavailable"},
		{"a day the data does not reach", wrap(domain.ErrNotFound), http.StatusNotFound, "not_found"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := newReportServer(t, &fakeReports{overview: func(context.Context, curtailment.Kind, time.Time) (curtailment.Overview, error) {
				return curtailment.Overview{}, tt.err
			}})
			res, body := do(t, srv, "GET", "/v1/curtailment?date=2020-01-01", "")
			if res.StatusCode != tt.wantStatus || body["code"] != tt.wantCode {
				t.Errorf("got %d %v", res.StatusCode, body)
			}
			if tt.wantStatus == http.StatusServiceUnavailable && res.Header.Get("Retry-After") != "60" {
				t.Errorf("Retry-After = %q", res.Header.Get("Retry-After"))
			}
		})
	}
}

func TestListPlants(t *testing.T) {
	var gotFilter curtailment.PlantFilter
	var gotPage reports.Page
	srv := newReportServer(t, &fakeReports{plants: func(_ context.Context, _ curtailment.Kind, _ time.Time, f curtailment.PlantFilter, page reports.Page) (curtailment.PlantRanking, error) {
		gotFilter, gotPage = f, page
		return curtailment.PlantRanking{
			Period:    curtailment.Period{Kind: curtailment.Weekly, Start: friday, End: friday, Through: friday},
			Latest:    latest,
			Reporting: 236,
			Limited:   48,
			Items: []curtailment.PlantTotal{{
				Plant:          curtailment.Plant{ID: "CJU_RNASSU", Name: "Conj. Assu Sol", State: "RN", Subsystem: curtailment.Northeast, Source: curtailment.Solar},
				Energy:         curtailment.Energy{CurtailedMWh: 2729, GeneratedMWh: 4100},
				LimitedMinutes: 540,
			}},
		}, nil
	}})

	res, body := do(t, srv, "GET", "/v1/plants?period=week", "")
	if res.StatusCode != http.StatusOK || body["total"] != 48.0 || body["reporting"] != 236.0 {
		t.Fatalf("got %d %v", res.StatusCode, body)
	}
	if gotFilter != (curtailment.PlantFilter{}) || gotPage != (reports.Page{Limit: 20, Offset: 0}) {
		t.Errorf("filter %+v, page %+v: want everything, the first 20", gotFilter, gotPage)
	}
	item := body["items"].([]any)[0].(map[string]any)
	if item["name"] != "Conj. Assu Sol" || item["source"] != "solar" || item["limitedMinutes"] != 540.0 {
		t.Errorf("items[0] = %v", item)
	}

	do(t, srv, "GET", "/v1/plants?source=wind&subsystem=S&limit=50&offset=100", "")
	if gotFilter != (curtailment.PlantFilter{Source: curtailment.Wind, Subsystem: curtailment.South}) || gotPage != (reports.Page{Limit: 50, Offset: 100}) {
		t.Errorf("filter %+v, page %+v", gotFilter, gotPage)
	}
	for _, query := range []string{"limit=0", "limit=51", "offset=-1", "source=coal", "subsystem=ne"} {
		if res, body := do(t, srv, "GET", "/v1/plants?"+query, ""); res.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: %d %v", query, res.StatusCode, body)
		}
	}
}

func TestListMonthlyCurtailmentLeavesOutWhatWasNotPublished(t *testing.T) {
	srv := newReportServer(t, &fakeReports{history: func(context.Context) (curtailment.History, error) {
		return curtailment.History{Latest: latest, Months: []curtailment.MonthTotal{
			{Start: time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC), Wind: &curtailment.Energy{CurtailedMWh: 9651.6, GeneratedMWh: 3e6}},
		}}, nil
	}})

	res, body := do(t, srv, "GET", "/v1/curtailment/months", "")
	month := body["months"].([]any)[0].(map[string]any)
	if res.StatusCode != http.StatusOK || month["month"] != "2021-10-01" || month["wind"] == nil {
		t.Errorf("got %d %v", res.StatusCode, body)
	}
	if _, ok := month["solar"]; ok {
		t.Errorf("a month before solar was published has solar: %v", month)
	}
}

func TestListDailyCurtailment(t *testing.T) {
	var asked int
	srv := newReportServer(t, &fakeReports{recent: func(_ context.Context, days int) (curtailment.Recent, error) {
		asked = days
		return curtailment.Recent{From: friday.AddDate(0, 0, 1-days), To: friday, Series: []curtailment.SubsystemSeries{
			{Source: curtailment.Wind, Subsystem: curtailment.Northeast, CurtailedMWh: make([]float64, days)},
		}}, nil
	}})

	res, body := do(t, srv, "GET", "/v1/curtailment/days", "")
	if res.StatusCode != http.StatusOK || asked != 30 || body["from"] != "2026-09-17" || body["to"] != "2026-10-16" {
		t.Errorf("got %d %v, asked for %d days", res.StatusCode, body, asked)
	}
	if res, _ := do(t, srv, "GET", "/v1/curtailment/days?days=91", ""); res.StatusCode != http.StatusBadRequest {
		t.Errorf("91 days: %d", res.StatusCode)
	}
}

func TestAlertsCarryTheirActivity(t *testing.T) {
	fired := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	quiet := anAlert
	quiet.ID, quiet.Name = uuid.NewV7(), "Wind in the South"
	srv := newServer(t, &fakeAlerts{list: func(context.Context, uuid.UUID) ([]alerting.Watched, error) {
		return []alerting.Watched{
			{Alert: anAlert, Activity: alert.Activity{LastFiredOn: fired, DaysFired: 7}},
			{Alert: quiet},
		}, nil
	}}, nil)

	res, body := do(t, srv, "GET", "/v1/alerts", "", "Authorization", "Bearer "+token)
	items, _ := body["items"].([]any)
	if res.StatusCode != http.StatusOK || len(items) != 2 {
		t.Fatalf("got %d %v", res.StatusCode, body)
	}
	first, second := items[0].(map[string]any), items[1].(map[string]any)
	if first["lastFiredOn"] != "2026-10-11" || first["daysFired"] != 7.0 || first["name"] != anAlert.Name {
		t.Errorf("items[0] = %v", first)
	}
	if _, ok := second["lastFiredOn"]; ok || second["daysFired"] != 0.0 {
		t.Errorf("an alert that never fired: %v", second)
	}
}

func TestListFirings(t *testing.T) {
	var limit, offset int
	srv := newServer(t, &fakeAlerts{firings: func(_ context.Context, owner uuid.UUID, l, o int) (alerting.FiringPage, error) {
		if owner != ana.AccountID {
			return alerting.FiringPage{}, errors.New("someone else's firings")
		}
		limit, offset = l, o
		return alerting.FiringPage{From: friday.AddDate(0, 0, -29), Through: friday, Total: 35, Items: []alert.Firing{
			{AlertID: anAlert.ID, AlertName: anAlert.Name, Day: fired(), CurtailedMWh: 169327.4, ThresholdMWh: 150000},
		}}, nil
	}}, nil)

	res, body := do(t, srv, "GET", "/v1/alerts/firings?offset=10", "", "Authorization", "Bearer "+token)
	if res.StatusCode != http.StatusOK || body["total"] != 35.0 || body["from"] != "2026-09-17" || limit != 10 || offset != 10 {
		t.Fatalf("got %d %v, limit %d offset %d", res.StatusCode, body, limit, offset)
	}
	item := body["items"].([]any)[0].(map[string]any)
	if item["day"] != "2026-10-11" || item["alertName"] != anAlert.Name || item["thresholdMwh"] != 150000.0 {
		t.Errorf("items[0] = %v", item)
	}

	if res, body := do(t, srv, "GET", "/v1/alerts/firings", ""); res.StatusCode != http.StatusUnauthorized {
		t.Errorf("without a session: %d %v", res.StatusCode, body)
	}
}

func fired() time.Time { return time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC) }

func newReportServer(t *testing.T, r *fakeReports) *httptest.Server {
	t.Helper()
	uc := useCases(&fakeAlerts{}, &fakeIdentity{})
	uc.Reports = r
	return serve(t, uc)
}

type fakeReports struct {
	overview func(context.Context, curtailment.Kind, time.Time) (curtailment.Overview, error)
	plants   func(context.Context, curtailment.Kind, time.Time, curtailment.PlantFilter, reports.Page) (curtailment.PlantRanking, error)
	history  func(context.Context) (curtailment.History, error)
	recent   func(context.Context, int) (curtailment.Recent, error)
}

func (f *fakeReports) Overview(ctx context.Context, kind curtailment.Kind, anchor time.Time) (curtailment.Overview, error) {
	return f.overview(ctx, kind, anchor)
}

func (f *fakeReports) Plants(ctx context.Context, kind curtailment.Kind, anchor time.Time, filter curtailment.PlantFilter, page reports.Page) (curtailment.PlantRanking, error) {
	return f.plants(ctx, kind, anchor, filter, page)
}

func (f *fakeReports) History(ctx context.Context) (curtailment.History, error) {
	return f.history(ctx)
}

func (f *fakeReports) Recent(ctx context.Context, days int) (curtailment.Recent, error) {
	return f.recent(ctx, days)
}
