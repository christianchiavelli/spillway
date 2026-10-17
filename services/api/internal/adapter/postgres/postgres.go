// Package postgres stores accounts, sessions and alerts in PostgreSQL 18. The
// SQL lives in queries/ and migrations/; sqlc turns the queries into the typed
// Go of sqlcdb, and this package maps those rows to the domain. Times come back
// from pgx in the server's zone and leave here in UTC, as the domain makes them.
package postgres

// sqlc runs at a pinned version rather than as a module tool: its parsers
// would join the API's module graph for a step that only writes code.
//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate --file ../../../sqlc.yaml

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// PoolConfig sizes the connection pool.
type PoolConfig struct {
	URL      string
	MaxConns int32
	// Tracer, when set, sees every query; telemetry passes an OpenTelemetry one.
	Tracer pgx.QueryTracer
}

// Connect opens a pool and checks that the database answers.
func Connect(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	pc, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err) // the error never echoes the password
	}
	pc.MaxConns = cfg.MaxConns
	pc.MinConns = min(2, cfg.MaxConns)
	pc.MaxConnLifetime = time.Hour
	pc.MaxConnLifetimeJitter = 5 * time.Minute // so the pool does not reconnect all at once
	pc.MaxConnIdleTime = 15 * time.Minute
	pc.HealthCheckPeriod = time.Minute
	pc.ConnConfig.Tracer = cfg.Tracer

	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("reach the database: %w", err)
	}
	return pool, nil
}

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate applies every migration not yet applied. A Postgres advisory lock
// holds off a second migrator, so replicas started together take turns.
func Migrate(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("create migration lock: %w", err)
	}
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys,
		goose.WithSessionLocker(locker),
		goose.WithDisableGlobalRegistry(true),
	)
	if err != nil {
		return fmt.Errorf("prepare migrations: %w", err)
	}
	results, err := provider.Up(ctx)
	for _, r := range results {
		logger.InfoContext(ctx, "migration applied",
			slog.Int64("version", r.Source.Version),
			slog.String("file", r.Source.Path),
			slog.Duration("took", r.Duration))
	}
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	if len(results) == 0 {
		logger.InfoContext(ctx, "schema up to date")
	}
	return nil
}

// isUniqueViolation reports whether err is Postgres refusing a duplicate in
// the named constraint or unique index.
func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == constraint
}
