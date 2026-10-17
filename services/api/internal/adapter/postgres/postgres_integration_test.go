//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/postgres"
	"github.com/christianchiavelli/spillway/services/api/internal/domain"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/account"
	"github.com/christianchiavelli/spillway/services/api/internal/domain/alert"
)

// Image is the Postgres the compose file runs, so the tests read what production reads.
const Image = "postgres:18.6-alpine3.24"

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()
	url, stop, err := database(ctx)
	if err != nil {
		log.Printf("start postgres: %v", err)
		return 1
	}
	defer stop()

	// Twelve connections: the concurrency test needs more than one in flight.
	pool, err = postgres.Connect(ctx, postgres.PoolConfig{URL: url, MaxConns: 12})
	if err != nil {
		log.Printf("connect: %v", err)
		return 1
	}
	defer pool.Close()
	if err := postgres.Migrate(ctx, pool, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		log.Printf("migrate: %v", err)
		return 1
	}
	return m.Run()
}

// database is TEST_DATABASE_URL when set, an empty database of one's own, and
// otherwise a new Postgres in a container, gone when the tests end.
func database(ctx context.Context) (string, func(), error) {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url, func() {}, nil
	}
	ctr, err := tcpostgres.Run(ctx, Image,
		tcpostgres.WithDatabase("spillway"),
		tcpostgres.WithUsername("spillway"),
		tcpostgres.WithPassword("spillway"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return "", nil, fmt.Errorf("run container: %w", err)
	}
	stop := func() {
		// Without the cancel: the container goes even when the run was cut short.
		if err := ctr.Terminate(context.WithoutCancel(ctx)); err != nil {
			log.Printf("stop postgres: %v", err)
		}
	}
	url, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		stop()
		return "", nil, fmt.Errorf("read connection string: %w", err)
	}
	return url, stop, nil
}

func TestMigrateTwiceChangesNothing(t *testing.T) {
	if err := postgres.Migrate(t.Context(), pool, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("second Migrate() = %v", err)
	}
}

func TestAccounts(t *testing.T) {
	t.Parallel()
	accounts := postgres.NewAccounts(pool)
	a := newAccount(t)

	got, err := accounts.ByEmail(t.Context(), a.Email)
	if err != nil {
		t.Fatal(err)
	}
	if got != a {
		t.Errorf("ByEmail() = %+v, want %+v", got, a)
	}

	twin := a
	twin.ID = uuid.NewV7()
	if err := accounts.Create(t.Context(), twin); !errors.Is(err, account.ErrEmailTaken) {
		t.Errorf("Create() with a used address = %v, want account.ErrEmailTaken", err)
	}
	if _, err := accounts.ByEmail(t.Context(), "nobody@example.com"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ByEmail() of nobody = %v, want domain.ErrNotFound", err)
	}
}

func TestSessions(t *testing.T) {
	t.Parallel()
	sessions := postgres.NewSessions(pool)
	a := newAccount(t)
	start := time.Now().UTC().Truncate(time.Microsecond)
	sess := account.NewSession(account.HashToken(uuid.New().String()), a.ID, start)
	if err := sessions.Create(t.Context(), sess); err != nil {
		t.Fatal(err)
	}

	gotSess, gotAcc, err := sessions.Find(t.Context(), sess.TokenHash)
	if err != nil {
		t.Fatal(err)
	}
	if !gotSess.ExpiresAt.Equal(sess.ExpiresAt) || gotSess.AccountID != a.ID || gotAcc.Email != a.Email {
		t.Errorf("Find() = %+v, %+v", gotSess, gotAcc)
	}
	if gotAcc.PasswordHash != "" {
		t.Error("Find() returned the password hash")
	}

	later := sess.ExpiresAt.Add(time.Hour)
	if err := sessions.Extend(t.Context(), sess.TokenHash, later); err != nil {
		t.Fatal(err)
	}
	if gotSess, _, _ := sessions.Find(t.Context(), sess.TokenHash); !gotSess.ExpiresAt.Equal(later) {
		t.Errorf("ExpiresAt after Extend() = %v, want %v", gotSess.ExpiresAt, later)
	}

	if err := sessions.Delete(t.Context(), sess.TokenHash); err != nil {
		t.Fatal(err)
	}
	if err := sessions.Delete(t.Context(), sess.TokenHash); err != nil {
		t.Errorf("a second Delete() = %v, want nil", err)
	}
	if _, _, err := sessions.Find(t.Context(), sess.TokenHash); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Find() after Delete() = %v, want domain.ErrNotFound", err)
	}
}

func TestDeleteExpiredSessions(t *testing.T) {
	// Not parallel: it deletes every session that is over, other tests' too.
	sessions := postgres.NewSessions(pool)
	a := newAccount(t)
	start := time.Now().UTC().Add(-30 * 24 * time.Hour).Truncate(time.Microsecond)
	over := account.NewSession(account.HashToken("over "+a.ID.String()), a.ID, start)
	live := account.NewSession(account.HashToken("live "+a.ID.String()), a.ID, time.Now())
	for _, s := range []account.Session{over, live} {
		if err := sessions.Create(t.Context(), s); err != nil {
			t.Fatal(err)
		}
	}

	n, err := sessions.DeleteExpired(t.Context(), time.Now())
	if err != nil || n < 1 {
		t.Fatalf("DeleteExpired() = %d, %v", n, err)
	}
	if _, _, err := sessions.Find(t.Context(), over.TokenHash); !errors.Is(err, domain.ErrNotFound) {
		t.Error("the expired session survived")
	}
	if _, _, err := sessions.Find(t.Context(), live.TokenHash); err != nil {
		t.Errorf("the live session is gone: %v", err)
	}
}

func TestAlerts(t *testing.T) {
	t.Parallel()
	alerts := postgres.NewAlerts(pool)
	owner := newAccount(t)

	first := newAlert(t, owner.ID, "Wind in the Northeast", time.Now().Add(-time.Minute))
	second := newAlert(t, owner.ID, "Solar in the South", time.Now())
	for _, a := range []alert.Alert{first, second} {
		if err := alerts.Add(t.Context(), a, alert.MaxPerAccount); err != nil {
			t.Fatal(err)
		}
	}

	list, err := alerts.ListByAccount(t.Context(), owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0] != second || list[1] != first {
		t.Errorf("ListByAccount() = %+v, want newest first", list)
	}

	twin := newAlert(t, owner.ID, "WIND in the northeast", time.Now())
	if err := alerts.Add(t.Context(), twin, alert.MaxPerAccount); !errors.Is(err, alert.ErrNameTaken) {
		t.Errorf("Add() of a repeated name = %v, want alert.ErrNameTaken", err)
	}
	someoneElse := newAccount(t)
	if err := alerts.Add(t.Context(), newAlert(t, someoneElse.ID, first.Name, time.Now()), alert.MaxPerAccount); err != nil {
		t.Errorf("another account cannot reuse the name: %v", err)
	}

	if err := alerts.Delete(t.Context(), someoneElse.ID, first.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete() by another account = %v, want domain.ErrNotFound", err)
	}
	if err := alerts.Delete(t.Context(), owner.ID, first.ID); err != nil {
		t.Fatal(err)
	}
	if err := alerts.Delete(t.Context(), owner.ID, first.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("a second Delete() = %v, want domain.ErrNotFound", err)
	}
}

// Forty requests for one account at the same moment, as many as the pool
// lets run, must end with twenty alerts: the limit holds under a race.
func TestTheAlertLimitHoldsUnderConcurrentAdds(t *testing.T) {
	t.Parallel()
	alerts := postgres.NewAlerts(pool)
	owner := newAccount(t)

	drafts := make([]alert.Alert, 40)
	for i := range drafts {
		drafts[i] = newAlert(t, owner.ID, fmt.Sprintf("alert %02d", i), time.Now())
	}

	var (
		stored, refused atomic.Int32
		unexpected      = make(chan error, len(drafts))
		start           = make(chan struct{})
		wg              sync.WaitGroup
	)
	for _, a := range drafts {
		wg.Go(func() {
			<-start
			err := alerts.Add(t.Context(), a, alert.MaxPerAccount)
			switch {
			case err == nil:
				stored.Add(1)
			case errors.Is(err, alert.ErrLimitReached):
				refused.Add(1)
			default:
				unexpected <- err
			}
		})
	}
	close(start)
	wg.Wait()
	close(unexpected)

	for err := range unexpected {
		t.Errorf("unexpected error: %v", err)
	}
	if stored.Load() != alert.MaxPerAccount || refused.Load() != 40-alert.MaxPerAccount {
		t.Errorf("stored %d and refused %d, want %d and %d", stored.Load(), refused.Load(), alert.MaxPerAccount, 40-alert.MaxPerAccount)
	}
	if list, _ := alerts.ListByAccount(t.Context(), owner.ID); len(list) != alert.MaxPerAccount {
		t.Errorf("the account holds %d alerts", len(list))
	}
}

func TestTheDatabaseRefusesWhatTheDomainWould(t *testing.T) {
	t.Parallel()
	owner := newAccount(t)
	for name, stmt := range map[string]string{
		"padded name":     `insert into alerts values ($1, $2, ' padded ', 'wind', 'NE', 10, now())`,
		"unknown source":  `insert into alerts values ($1, $2, 'name', 'coal', 'NE', 10, now())`,
		"lower-case code": `insert into alerts values ($1, $2, 'name', 'wind', 'ne', 10, now())`,
		"zero threshold":  `insert into alerts values ($1, $2, 'name', 'wind', 'NE', 0, now())`,
	} {
		if _, err := pool.Exec(t.Context(), stmt, uuid.NewV7(), owner.ID); err == nil {
			t.Errorf("%s: the insert went through", name)
		}
	}
	if _, err := pool.Exec(t.Context(), `insert into accounts values ($1, 'Ana@Example.com', 'x', now())`, uuid.NewV7()); err == nil {
		t.Error("an address that is not normalised went through")
	}
}

func newAccount(t *testing.T) account.Account {
	t.Helper()
	id := uuid.NewV7()
	a := account.Account{
		ID:           id,
		Email:        "someone-" + id.String() + "@example.com",
		PasswordHash: "$argon2id$v=19$m=64,t=1,p=1$c2FsdA$aGFzaA",
		CreatedAt:    time.Now().UTC().Truncate(time.Microsecond),
	}
	if err := postgres.NewAccounts(pool).Create(t.Context(), a); err != nil {
		t.Fatalf("create account: %v", err)
	}
	return a
}

func newAlert(t *testing.T, owner uuid.UUID, name string, at time.Time) alert.Alert {
	t.Helper()
	a, err := alert.New(uuid.NewV7(), owner, alert.Draft{Name: name, Source: "wind", Subsystem: "NE", ThresholdMWh: 500}, at)
	if err != nil {
		t.Fatalf("new alert: %v", err)
	}
	return a
}
