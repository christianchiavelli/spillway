package alerting_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/app/alerting"
	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
)

var now = time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

func TestCreateStoresTheAlert(t *testing.T) {
	store := &memStore{}
	svc := alerting.New(store, noData{}, func() time.Time { return now })
	owner := uuid.NewV7()

	got, err := svc.Create(t.Context(), owner, draft("Wind in the Northeast"))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if got.AccountID != owner || got.Name != "Wind in the Northeast" || !got.CreatedAt.Equal(now) {
		t.Errorf("Create() = %+v", got)
	}
	if got.ID == (uuid.UUID{}) {
		t.Error("Create() gave the alert no id")
	}
	if list, _ := store.ListByAccount(t.Context(), owner); !slices.Equal(list, []alert.Alert{got}) {
		t.Errorf("stored %v, want %v", list, []alert.Alert{got})
	}
}

func TestCreateStoresNothingWhenARuleIsBroken(t *testing.T) {
	store := &memStore{}
	svc := alerting.New(store, noData{}, func() time.Time { return now })

	_, err := svc.Create(t.Context(), uuid.NewV7(), alert.Draft{Name: "", Source: "wind", Subsystem: "NE", ThresholdMWh: 10})

	var verr *domain.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("Create() error = %v, want a *domain.ValidationError", err)
	}
	if store.calls != 0 {
		t.Errorf("the store was called %d times", store.calls)
	}
}

func TestCreatePassesOnWhatTheStoreRefuses(t *testing.T) {
	svc := alerting.New(&memStore{}, noData{}, func() time.Time { return now })
	owner := uuid.NewV7()

	if _, err := svc.Create(t.Context(), owner, draft("Solar in the South")); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Create(t.Context(), owner, draft("SOLAR in the south"))
	if !errors.Is(err, alert.ErrNameTaken) {
		t.Errorf("a repeated name gave %v, want alert.ErrNameTaken", err)
	}

	for i := range alert.MaxPerAccount - 1 {
		if _, err := svc.Create(t.Context(), owner, draft("alert "+string(rune('a'+i)))); err != nil {
			t.Fatalf("alert %d: %v", i+2, err)
		}
	}
	_, err = svc.Create(t.Context(), owner, draft("one too many"))
	if !errors.Is(err, alert.ErrLimitReached) || !errors.Is(err, domain.ErrConflict) {
		t.Errorf("alert %d gave %v, want alert.ErrLimitReached", alert.MaxPerAccount+1, err)
	}
}

func TestDeleteReportsSomeoneElsesAlertAsMissing(t *testing.T) {
	svc := alerting.New(&memStore{}, noData{}, func() time.Time { return now })
	owner, stranger := uuid.NewV7(), uuid.NewV7()
	a, err := svc.Create(t.Context(), owner, draft("Wind in the North"))
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.Delete(t.Context(), stranger, a.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete() by a stranger = %v, want domain.ErrNotFound", err)
	}
	if err := svc.Delete(t.Context(), owner, a.ID); err != nil {
		t.Errorf("Delete() by the owner = %v", err)
	}
	if list, _ := svc.List(t.Context(), owner); len(list) != 0 {
		t.Errorf("List() after Delete() = %v", list)
	}
}

func TestActivityAndFiringsCoverTheRecentDays(t *testing.T) {
	store := &memStore{}
	calendar := bounds{curtailment.Bounds{
		First:  time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC),
		Latest: time.Date(2026, 10, 16, 23, 30, 0, 0, curtailment.Brasilia),
	}}
	svc := alerting.New(store, calendar, func() time.Time { return now })
	owner := uuid.NewV7()
	a, err := svc.Create(t.Context(), owner, draft("Solar in the South"))
	if err != nil {
		t.Fatal(err)
	}
	store.activity = map[uuid.UUID]alert.Activity{a.ID: {LastFiredOn: time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC), DaysFired: 3}}

	list, err := svc.List(t.Context(), owner)
	if err != nil || len(list) != 1 || list[0].DaysFired != 3 || list[0].ID != a.ID {
		t.Fatalf("List() = %+v, %v", list, err)
	}
	wantFrom, wantThrough := time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	if !store.from.Equal(wantFrom) || !store.through.Equal(wantThrough) {
		t.Errorf("activity read from %v to %v, want the 30 days to the 16th", store.from, store.through)
	}

	page, err := svc.Firings(t.Context(), owner, 10, 0)
	if err != nil || !page.From.Equal(wantFrom) || !page.Through.Equal(wantThrough) {
		t.Errorf("Firings() = %+v, %v", page, err)
	}
}

func TestBeforeAnyDataAlertsHaveNoActivity(t *testing.T) {
	store := &memStore{}
	svc := alerting.New(store, noData{}, func() time.Time { return now })
	owner := uuid.NewV7()
	if _, err := svc.Create(t.Context(), owner, draft("Wind in the North")); err != nil {
		t.Fatal(err)
	}

	list, err := svc.List(t.Context(), owner)
	if err != nil || len(list) != 1 || list[0].Activity != (alert.Activity{}) {
		t.Errorf("List() = %+v, %v", list, err)
	}
	page, err := svc.Firings(t.Context(), owner, 10, 0)
	if err != nil || page.Total != 0 || page.Items == nil || !page.From.IsZero() {
		t.Errorf("Firings() = %+v, %v; want an empty page", page, err)
	}
}

func draft(name string) alert.Draft {
	return alert.Draft{Name: name, Source: "solar", Subsystem: "S", ThresholdMWh: 120}
}

// memStore keeps alerts in memory with the same rules as the Postgres store,
// and answers activity from what a test puts in it.
type memStore struct {
	mu            sync.Mutex
	alerts        []alert.Alert
	calls         int
	activity      map[uuid.UUID]alert.Activity
	from, through time.Time
}

func (m *memStore) Activity(_ context.Context, _ uuid.UUID, from, through time.Time) (map[uuid.UUID]alert.Activity, error) {
	m.from, m.through = from, through
	return m.activity, nil
}

func (m *memStore) Firings(_ context.Context, _ uuid.UUID, from, through time.Time, _, _ int) ([]alert.Firing, int, error) {
	m.from, m.through = from, through
	return []alert.Firing{}, 0, nil
}

type noData struct{}

func (noData) Bounds(context.Context) (curtailment.Bounds, error) {
	return curtailment.Bounds{}, curtailment.ErrNoData
}

type bounds struct{ b curtailment.Bounds }

func (c bounds) Bounds(context.Context) (curtailment.Bounds, error) { return c.b, nil }

func (m *memStore) Add(_ context.Context, a alert.Alert, max int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	held := 0
	for _, b := range m.alerts {
		if b.AccountID != a.AccountID {
			continue
		}
		held++
		if strings.EqualFold(b.Name, a.Name) {
			return alert.ErrNameTaken
		}
	}
	if held >= max {
		return alert.ErrLimitReached
	}
	m.alerts = append(m.alerts, a)
	return nil
}

func (m *memStore) ListByAccount(_ context.Context, accountID uuid.UUID) ([]alert.Alert, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	var out []alert.Alert
	for _, a := range slices.Backward(m.alerts) {
		if a.AccountID == accountID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (m *memStore) Delete(_ context.Context, accountID, alertID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	i := slices.IndexFunc(m.alerts, func(a alert.Alert) bool { return a.ID == alertID && a.AccountID == accountID })
	if i < 0 {
		return domain.ErrNotFound
	}
	m.alerts = slices.Delete(m.alerts, i, i+1)
	return nil
}
