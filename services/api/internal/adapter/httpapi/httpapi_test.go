package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/httpapi"
	"github.com/christianchiavelli/spillway/services/api/internal/app/alerting"
	"github.com/christianchiavelli/spillway/services/api/internal/app/identity"
	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/account"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/curtailment"
	"github.com/christianchiavelli/spillway/services/api/internal/platform/logging"
)

const token = "a-valid-session-token"

var (
	ana      = identity.Principal{AccountID: uuid.NewV7(), Email: "ana@example.com", ExpiresAt: time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)}
	anAlert  = alert.Alert{ID: uuid.NewV7(), AccountID: ana.AccountID, Name: "Wind in the Northeast", Source: curtailment.Wind, Subsystem: curtailment.Northeast, ThresholdMWh: 500, CreatedAt: time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)}
	validNew = `{"name":"Wind in the Northeast","source":"wind","subsystem":"NE","thresholdMwh":500}`
)

func TestCreateAlert(t *testing.T) {
	alerts := &fakeAlerts{create: func(_ context.Context, owner uuid.UUID, d alert.Draft) (alert.Alert, error) {
		if owner != ana.AccountID {
			t.Errorf("created for %v, want the signed-in account", owner)
		}
		if d != (alert.Draft{Name: "Wind in the Northeast", Source: "wind", Subsystem: "NE", ThresholdMWh: 500}) {
			t.Errorf("draft = %+v", d)
		}
		return anAlert, nil
	}}
	srv := newServer(t, alerts, nil)

	res, body := do(t, srv, "POST", "/v1/alerts", validNew, "Authorization", "Bearer "+token)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, body = %v", res.StatusCode, body)
	}
	if body["id"] != anAlert.ID.String() || body["subsystem"] != "NE" || body["thresholdMwh"] != 500.0 || body["createdAt"] != "2026-10-17T12:00:00Z" {
		t.Errorf("body = %v", body)
	}
	if got := res.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q", got)
	}
}

func TestEveryOperationButSigningInNeedsASession(t *testing.T) {
	srv := newServer(t, &fakeAlerts{}, nil)
	for _, tt := range []struct{ method, path, body string }{
		{"GET", "/v1/alerts", ""},
		{"POST", "/v1/alerts", validNew},
		{"DELETE", "/v1/alerts/" + anAlert.ID.String(), ""},
		{"GET", "/v1/sessions/current", ""},
		{"DELETE", "/v1/sessions/current", ""},
	} {
		for _, c := range []struct{ auth, challenge string }{
			{"", "Bearer"},
			{"Basic YW5hOnB3", "Bearer"},
			{"Bearer not-a-session", `Bearer error="invalid_token"`},
		} {
			res, body := do(t, srv, tt.method, tt.path, tt.body, "Authorization", c.auth)
			if res.StatusCode != http.StatusUnauthorized || body["code"] != "unauthenticated" {
				t.Errorf("%s %s with %q = %d %v", tt.method, tt.path, c.auth, res.StatusCode, body)
			}
			if got := res.Header.Get("WWW-Authenticate"); got != c.challenge {
				t.Errorf("%s %s with %q: WWW-Authenticate = %q, want %q", tt.method, tt.path, c.auth, got, c.challenge)
			}
			if ct := res.Header.Get("Content-Type"); ct != "application/problem+json" {
				t.Errorf("%s %s Content-Type = %q", tt.method, tt.path, ct)
			}
		}
	}
}

func TestTheEdgeChecksTheBodyAgainstTheContract(t *testing.T) {
	srv := newServer(t, &fakeAlerts{create: func(context.Context, uuid.UUID, alert.Draft) (alert.Alert, error) {
		t.Error("a body that breaks the contract reached the use case")
		return alert.Alert{}, nil
	}}, nil)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
		wantErrors []string
	}{
		{
			name:       "every field wrong",
			body:       `{"name":"","source":"coal","subsystem":"ne","thresholdMwh":0}`,
			wantStatus: 422, wantCode: "validation_failed",
			wantErrors: []string{"/name required", "/source invalid", "/subsystem invalid", "/thresholdMwh out_of_range"},
		},
		{
			name:       "fields missing",
			body:       `{"source":"wind"}`,
			wantStatus: 422, wantCode: "validation_failed",
			wantErrors: []string{"/name required", "/subsystem required", "/thresholdMwh required"},
		},
		{
			name:       "too long and too large",
			body:       `{"name":"` + strings.Repeat("a", 81) + `","source":"wind","subsystem":"NE","thresholdMwh":1000001}`,
			wantStatus: 422, wantCode: "validation_failed",
			wantErrors: []string{"/name too_long", "/thresholdMwh out_of_range"},
		},
		{name: "a field the contract does not have", body: `{"name":"x","source":"wind","subsystem":"NE","thresholdMwh":1,"owner":"me"}`, wantStatus: 400, wantCode: "malformed_request"},
		{name: "a number as text", body: `{"name":"x","source":"wind","subsystem":"NE","thresholdMwh":"500"}`, wantStatus: 400, wantCode: "malformed_request"},
		{name: "not JSON", body: `name=x`, wantStatus: 400, wantCode: "malformed_request"},
		{name: "no body", body: ``, wantStatus: 400, wantCode: "malformed_request"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, body := do(t, srv, "POST", "/v1/alerts", tt.body, "Authorization", "Bearer "+token)
			if res.StatusCode != tt.wantStatus || body["code"] != tt.wantCode {
				t.Fatalf("got %d %v, want %d %s", res.StatusCode, body, tt.wantStatus, tt.wantCode)
			}
			if got := fieldErrors(body); !slices.Equal(got, tt.wantErrors) {
				t.Errorf("errors = %q, want %q", got, tt.wantErrors)
			}
		})
	}
}

func TestTheEdgeRefusesWhatItCannotRead(t *testing.T) {
	srv := newServer(t, &fakeAlerts{}, nil)

	res, body := do(t, srv, "POST", "/v1/alerts", validNew, "Authorization", "Bearer "+token, "Content-Type", "text/plain")
	if res.StatusCode != http.StatusUnsupportedMediaType || body["code"] != "unsupported_media_type" {
		t.Errorf("text/plain: %d %v", res.StatusCode, body)
	}

	huge := `{"name":"` + strings.Repeat("a", 70<<10) + `"}`
	res, body = do(t, srv, "POST", "/v1/alerts", huge, "Authorization", "Bearer "+token)
	if res.StatusCode != http.StatusRequestEntityTooLarge || body["code"] != "request_too_large" {
		t.Errorf("70 KiB body: %d %v", res.StatusCode, body)
	}
}

func TestDomainErrorsBecomeProblems(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantErrors []string
	}{
		{
			name:       "a rule only the domain knows",
			err:        &domain.ValidationError{Violations: []domain.Violation{{Field: "name", Code: domain.CodeRequired}}},
			wantStatus: 422, wantCode: "validation_failed", wantErrors: []string{"/name required"},
		},
		{name: "a name in use", err: wrap(alert.ErrNameTaken), wantStatus: 409, wantCode: "conflict", wantErrors: []string{"/name taken"}},
		{name: "the limit", err: wrap(alert.ErrLimitReached), wantStatus: 409, wantCode: "alert_limit_reached"},
		{name: "a failure", err: wrap(errors.New("connection reset by the database at 10.0.0.3")), wantStatus: 500, wantCode: "internal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newServer(t, &fakeAlerts{create: func(context.Context, uuid.UUID, alert.Draft) (alert.Alert, error) {
				return alert.Alert{}, tt.err
			}}, nil)

			res, body := do(t, srv, "POST", "/v1/alerts", validNew, "Authorization", "Bearer "+token)

			if res.StatusCode != tt.wantStatus || body["code"] != tt.wantCode || body["status"] != float64(tt.wantStatus) {
				t.Fatalf("got %d %v, want %d %s", res.StatusCode, body, tt.wantStatus, tt.wantCode)
			}
			if got := fieldErrors(body); !slices.Equal(got, tt.wantErrors) {
				t.Errorf("errors = %q, want %q", got, tt.wantErrors)
			}
			if strings.Contains(fmt.Sprint(body), "10.0.0.3") {
				t.Error("the problem leaks the cause")
			}
		})
	}
}

func TestDeletingSomeoneElsesAlertSaysNotFound(t *testing.T) {
	srv := newServer(t, &fakeAlerts{delete: func(_ context.Context, owner, id uuid.UUID) error {
		return wrap(domain.ErrNotFound)
	}}, nil)

	res, body := do(t, srv, "DELETE", "/v1/alerts/"+uuid.NewV7().String(), "", "Authorization", "Bearer "+token)
	if res.StatusCode != http.StatusNotFound || body["code"] != "not_found" {
		t.Errorf("got %d %v", res.StatusCode, body)
	}

	res, body = do(t, srv, "DELETE", "/v1/alerts/not-a-uuid", "", "Authorization", "Bearer "+token)
	if res.StatusCode != http.StatusBadRequest || body["code"] != "malformed_request" {
		t.Errorf("a malformed id: %d %v", res.StatusCode, body)
	}
}

func TestSignIn(t *testing.T) {
	ident := &fakeIdentity{signIn: func(_ context.Context, email, password string) (identity.Grant, error) {
		switch email {
		case "ana@example.com":
			return identity.Grant{
				Token:   "new-token",
				Session: account.Session{ExpiresAt: ana.ExpiresAt},
				Account: account.Account{ID: ana.AccountID, Email: email},
			}, nil
		case "slow@example.com":
			return identity.Grant{}, &identity.ThrottledError{RetryAfter: 11500 * time.Millisecond}
		}
		return identity.Grant{}, identity.ErrInvalidCredentials
	}}
	srv := newServer(t, &fakeAlerts{}, ident)

	res, body := do(t, srv, "POST", "/v1/sessions", `{"email":"ana@example.com","password":"anything"}`)
	if res.StatusCode != http.StatusCreated || body["token"] != "new-token" || body["expiresAt"] != "2026-10-24T12:00:00Z" {
		t.Errorf("right credentials: %d %v", res.StatusCode, body)
	}

	res, body = do(t, srv, "POST", "/v1/sessions", `{"email":"bia@example.com","password":"anything"}`)
	if res.StatusCode != http.StatusBadRequest || body["code"] != "invalid_credentials" || res.Header.Get("WWW-Authenticate") != "" {
		t.Errorf("wrong credentials: %d %v", res.StatusCode, body)
	}

	res, body = do(t, srv, "POST", "/v1/sessions", `{"email":"slow@example.com","password":"anything"}`)
	if res.StatusCode != http.StatusTooManyRequests || body["code"] != "rate_limited" || res.Header.Get("Retry-After") != "12" {
		t.Errorf("throttled: %d %v Retry-After=%q", res.StatusCode, body, res.Header.Get("Retry-After"))
	}
}

func TestCurrentSession(t *testing.T) {
	ident := &fakeIdentity{}
	srv := newServer(t, &fakeAlerts{}, ident)

	res, body := do(t, srv, "GET", "/v1/sessions/current", "", "Authorization", "Bearer "+token)
	account, _ := body["account"].(map[string]any)
	if res.StatusCode != http.StatusOK || account["email"] != "ana@example.com" || body["expiresAt"] != "2026-10-24T12:00:00Z" {
		t.Errorf("GET: %d %v", res.StatusCode, body)
	}

	res, _ = do(t, srv, "DELETE", "/v1/sessions/current", "", "Authorization", "Bearer "+token)
	if res.StatusCode != http.StatusNoContent || ident.signedOut != 1 {
		t.Errorf("DELETE: %d, signed out %d times", res.StatusCode, ident.signedOut)
	}
}

func TestBrowsersAreTurnedAway(t *testing.T) {
	srv := newServer(t, &fakeAlerts{}, nil)

	res, body := do(t, srv, "GET", "/v1/alerts", "", "Authorization", "Bearer "+token, "Origin", "https://evil.example")
	if res.StatusCode != http.StatusForbidden || body["code"] != "forbidden_origin" {
		t.Errorf("from a browser: %d %v", res.StatusCode, body)
	}
	if res.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Error("a refused origin got CORS headers")
	}

	allowed := httptest.NewServer(must(httpapi.New(useCases(&fakeAlerts{}, &fakeIdentity{}), options(func(o *httpapi.Options) {
		o.AllowedOrigins = []string{"http://localhost:5173"}
	}))))
	t.Cleanup(allowed.Close)
	res, _ = do(t, allowed, "OPTIONS", "/v1/alerts", "", "Origin", "http://localhost:5173", "Access-Control-Request-Method", "POST")
	if res.StatusCode != http.StatusNoContent || res.Header.Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("preflight from an allowed origin: %d %v", res.StatusCode, res.Header)
	}
}

func TestUnknownRoutesAndMethods(t *testing.T) {
	srv := newServer(t, &fakeAlerts{}, nil)

	res, body := do(t, srv, "GET", "/v2/alerts", "")
	if res.StatusCode != http.StatusNotFound || body["code"] != "not_found" {
		t.Errorf("unknown path: %d %v", res.StatusCode, body)
	}
	res, body = do(t, srv, "PUT", "/v1/alerts", validNew, "Authorization", "Bearer "+token)
	if res.StatusCode != http.StatusMethodNotAllowed || body["code"] != "method_not_allowed" || res.Header.Get("Allow") == "" {
		t.Errorf("wrong method: %d %v Allow=%q", res.StatusCode, body, res.Header.Get("Allow"))
	}
}

func TestAPanicIsAnsweredAndLogged(t *testing.T) {
	var logs strings.Builder
	srv := httptest.NewServer(must(httpapi.New(useCases(&fakeAlerts{list: func(context.Context, uuid.UUID) ([]alerting.Watched, error) {
		panic("boom")
	}}, &fakeIdentity{}), options(func(o *httpapi.Options) {
		o.Logger = must(logging.New(&logs, "text", slog.LevelInfo))
	}))))
	t.Cleanup(srv.Close)

	res, body := do(t, srv, "GET", "/v1/alerts", "", "Authorization", "Bearer "+token, "X-Request-Id", "req-42")
	if res.StatusCode != http.StatusInternalServerError || body["code"] != "internal" {
		t.Errorf("got %d %v", res.StatusCode, body)
	}
	if !strings.Contains(logs.String(), "panic=boom") || !strings.Contains(logs.String(), "request_id=req-42") {
		t.Errorf("logs = %s", logs.String())
	}
	if res.Header.Get("X-Request-Id") != "req-42" {
		t.Errorf("X-Request-Id = %q", res.Header.Get("X-Request-Id"))
	}
}

func options(change ...func(*httpapi.Options)) httpapi.Options {
	o := httpapi.Options{
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		MaxBodyBytes:   64 << 10,
		RequestTimeout: 5 * time.Second,
	}
	for _, c := range change {
		c(&o)
	}
	return o
}

func newServer(t *testing.T, alerts *fakeAlerts, ident *fakeIdentity) *httptest.Server {
	t.Helper()
	if ident == nil {
		ident = &fakeIdentity{}
	}
	return serve(t, useCases(alerts, ident))
}

func serve(t *testing.T, uc httpapi.UseCases) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(must(httpapi.New(uc, options())))
	t.Cleanup(srv.Close)
	return srv
}

func useCases(alerts *fakeAlerts, ident *fakeIdentity) httpapi.UseCases {
	return httpapi.UseCases{Alerts: alerts, Identity: ident, Reports: &fakeReports{}}
}

func do(t *testing.T, srv *httptest.Server, method, path, body string, headers ...string) (*http.Response, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		if headers[i+1] != "" {
			req.Header.Set(headers[i], headers[i+1])
		}
	}
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("%s %s answered %d with something that is not JSON: %s", method, path, res.StatusCode, raw)
		}
	}
	return res, decoded
}

func fieldErrors(body map[string]any) []string {
	list, _ := body["errors"].([]any)
	var out []string
	for _, e := range list {
		m := e.(map[string]any)
		out = append(out, m["pointer"].(string)+" "+m["code"].(string))
	}
	return out
}

func wrap(err error) error {
	return errors.Join(errors.New("context the use case added"), err)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

type fakeAlerts struct {
	create  func(context.Context, uuid.UUID, alert.Draft) (alert.Alert, error)
	list    func(context.Context, uuid.UUID) ([]alerting.Watched, error)
	delete  func(context.Context, uuid.UUID, uuid.UUID) error
	firings func(context.Context, uuid.UUID, int, int) (alerting.FiringPage, error)
}

func (f *fakeAlerts) Create(ctx context.Context, owner uuid.UUID, d alert.Draft) (alert.Alert, error) {
	if f.create == nil {
		return anAlert, nil
	}
	return f.create(ctx, owner, d)
}

func (f *fakeAlerts) List(ctx context.Context, owner uuid.UUID) ([]alerting.Watched, error) {
	if f.list == nil {
		return []alerting.Watched{{Alert: anAlert}}, nil
	}
	return f.list(ctx, owner)
}

func (f *fakeAlerts) Firings(ctx context.Context, owner uuid.UUID, limit, offset int) (alerting.FiringPage, error) {
	if f.firings == nil {
		return alerting.FiringPage{}, nil
	}
	return f.firings(ctx, owner, limit, offset)
}

func (f *fakeAlerts) Delete(ctx context.Context, owner, id uuid.UUID) error {
	if f.delete == nil {
		return nil
	}
	return f.delete(ctx, owner, id)
}

type fakeIdentity struct {
	signIn    func(context.Context, string, string) (identity.Grant, error)
	signedOut int
}

func (f *fakeIdentity) SignUp(context.Context, string, string) (identity.Grant, error) {
	return identity.Grant{}, errors.New("not used")
}

func (f *fakeIdentity) SignIn(ctx context.Context, email, password string) (identity.Grant, error) {
	return f.signIn(ctx, email, password)
}

func (f *fakeIdentity) SignOut(_ context.Context, p identity.Principal) error {
	if p.AccountID != ana.AccountID {
		return errors.New("signed out someone else")
	}
	f.signedOut++
	return nil
}

func (f *fakeIdentity) Authenticate(_ context.Context, t string) (identity.Principal, error) {
	if t != token {
		return identity.Principal{}, domain.ErrUnauthenticated
	}
	return ana, nil
}
