// Command api is Spillway's API server. One binary does four jobs, so the
// image needs no shell:
//
//	api serve        serve the API, and health and metrics on the admin port
//	api sync         load the ONS's new and revised files, now and every
//	                 SYNC_INTERVAL, with health and metrics on the admin port
//	api migrate      apply the database migrations, then exit
//	api healthcheck  exit 0 when this host's server is ready, for HEALTHCHECK
//
// Settings come from the environment; see internal/platform/config.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/argon2id"
	"github.com/christianchiavelli/spillway/services/api/internal/adapter/httpapi"
	"github.com/christianchiavelli/spillway/services/api/internal/adapter/ons"
	"github.com/christianchiavelli/spillway/services/api/internal/adapter/postgres"
	"github.com/christianchiavelli/spillway/services/api/internal/adapter/ratelimit"
	"github.com/christianchiavelli/spillway/services/api/internal/app/alerting"
	"github.com/christianchiavelli/spillway/services/api/internal/app/identity"
	"github.com/christianchiavelli/spillway/services/api/internal/app/ingest"
	"github.com/christianchiavelli/spillway/services/api/internal/app/reports"
	"github.com/christianchiavelli/spillway/services/api/internal/platform/config"
	"github.com/christianchiavelli/spillway/services/api/internal/platform/health"
	"github.com/christianchiavelli/spillway/services/api/internal/platform/httpserver"
	"github.com/christianchiavelli/spillway/services/api/internal/platform/logging"
	"github.com/christianchiavelli/spillway/services/api/internal/platform/schedule"
	"github.com/christianchiavelli/spillway/services/api/internal/platform/telemetry"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "api: %v\n", err)
		os.Exit(1)
	}
}

// run is main without the process: tests call it with their own context,
// arguments, environment and output.
func run(ctx context.Context, args []string, getenv func(string) string, stdout io.Writer) error {
	command := "serve"
	if len(args) > 0 {
		command = args[0]
	}
	switch command {
	case "serve":
		return serve(ctx, getenv, stdout)
	case "sync":
		return syncData(ctx, getenv, stdout)
	case "migrate":
		return migrate(ctx, getenv, stdout)
	case "healthcheck":
		return healthcheck(ctx, getenv)
	default:
		return fmt.Errorf("unknown command %q: want serve, sync, migrate or healthcheck", command)
	}
}

func setup(getenv func(string) string, stdout io.Writer) (config.Config, *slog.Logger, error) {
	cfg, err := config.Load(getenv)
	if err != nil {
		return config.Config{}, nil, fmt.Errorf("read settings:\n%w", err)
	}
	logger, err := logging.New(stdout, cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		return config.Config{}, nil, err
	}
	slog.SetDefault(logger)
	return cfg, logger, nil
}

func serve(ctx context.Context, getenv func(string) string, stdout io.Writer) error {
	ctx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	cfg, logger, err := setup(getenv, stdout)
	if err != nil {
		return err
	}

	tel, err := telemetry.Setup(ctx, telemetry.Options{Service: "spillway-api", Version: version, ExportTraces: cfg.OTLPEndpoint != ""})
	if err != nil {
		return err
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := tel.Shutdown(flushCtx); err != nil {
			logger.Warn("telemetry did not flush", slog.Any("error", err))
		}
	}()

	pool, err := postgres.Connect(ctx, postgres.PoolConfig{URL: cfg.DatabaseURL, MaxConns: cfg.DatabaseMaxConns, Tracer: tel.QueryTracer()})
	if err != nil {
		return err
	}
	defer pool.Close() // runs after the servers stop, so no request loses its connection
	if err := tel.RecordPool(pool); err != nil {
		return err
	}

	// Five attempts per address, then one more every twelve seconds.
	limiter := ratelimit.New(12*time.Second, 5, time.Now)
	ident, err := identity.New(ctx, identity.Deps{
		Accounts: postgres.NewAccounts(pool),
		Sessions: postgres.NewSessions(pool),
		Hasher:   argon2id.New(argon2id.DefaultParams, cfg.HashConcurrency),
		Limiter:  limiter,
		Now:      time.Now,
	})
	if err != nil {
		return err
	}
	curtailed := postgres.NewCurtailment(pool, time.Now)
	alerts := alerting.New(postgres.NewAlerts(pool), curtailed, time.Now)

	api, err := httpapi.New(httpapi.UseCases{Alerts: alerts, Identity: ident, Reports: reports.New(curtailed)}, httpapi.Options{
		Logger:         logger,
		AllowedOrigins: cfg.AllowedOrigins,
		MaxBodyBytes:   64 << 10,
		RequestTimeout: cfg.RequestTimeout,
		TracerProvider: tel.TracerProvider,
		MeterProvider:  tel.MeterProvider,
	})
	if err != nil {
		return err
	}

	probe := health.New(pool.Ping)
	admin := http.NewServeMux()
	admin.HandleFunc("GET /livez", probe.Live)
	admin.HandleFunc("GET /readyz", probe.Ready)
	admin.Handle("GET /metrics", tel.Metrics)

	// The servers stop when serving is cancelled, which happens after the
	// drain, not at the signal: in the meantime they keep answering, and
	// /readyz says not to send more.
	serving, stopServing := context.WithCancel(context.WithoutCancel(ctx))
	defer stopServing()

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		<-gctx.Done()
		stopSignals() // a second Ctrl+C now kills the process at once
		probe.SetReady(false)
		logger.Info("shutting down", slog.Duration("drain", cfg.ShutdownDrain), slog.Duration("timeout", cfg.ShutdownTimeout))
		time.Sleep(cfg.ShutdownDrain)
		stopServing()
		return nil
	})
	g.Go(func() error {
		return httpserver.Run(serving, httpserver.New(cfg.APIAddr, api, logger), cfg.ShutdownTimeout)
	})
	g.Go(func() error {
		return httpserver.Run(serving, httpserver.New(cfg.AdminAddr, admin, logger), cfg.ShutdownTimeout)
	})
	g.Go(func() error {
		return schedule.Every(gctx, 15*time.Minute, func(ctx context.Context) {
			n, err := ident.PurgeExpiredSessions(ctx)
			if err != nil {
				logger.ErrorContext(ctx, "purging expired sessions failed", slog.Any("error", err))
				return
			}
			logger.DebugContext(ctx, "purged expired sessions", slog.Int64("count", n))
		})
	})
	g.Go(func() error {
		return schedule.Every(gctx, time.Minute, func(context.Context) { limiter.Sweep() })
	})

	probe.SetReady(true)
	logger.Info("serving", slog.String("api", cfg.APIAddr), slog.String("admin", cfg.AdminAddr), slog.String("version", version))
	if err := g.Wait(); err != nil {
		return err
	}
	logger.Info("stopped")
	return nil
}

// syncData keeps Spillway's copy of the ONS's data current. It runs apart from
// serve, so the API needs no way out to the internet, and it answers health
// and metrics on the admin port as serve does.
func syncData(ctx context.Context, getenv func(string) string, stdout io.Writer) error {
	ctx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	cfg, logger, err := setup(getenv, stdout)
	if err != nil {
		return err
	}
	tel, err := telemetry.Setup(ctx, telemetry.Options{Service: "spillway-sync", Version: version, ExportTraces: cfg.OTLPEndpoint != ""})
	if err != nil {
		return err
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := tel.Shutdown(flushCtx); err != nil {
			logger.Warn("telemetry did not flush", slog.Any("error", err))
		}
	}()

	pool, err := postgres.Connect(ctx, postgres.PoolConfig{URL: cfg.DatabaseURL, MaxConns: 4, Tracer: tel.QueryTracer()})
	if err != nil {
		return err
	}
	defer pool.Close()

	// A file is a few megabytes; two minutes is a slow line, not a hung one.
	client := &http.Client{Timeout: 2 * time.Minute}
	sync := ingest.New(ons.New(cfg.ONSURL, client), postgres.NewCurtailment(pool, time.Now), logger)
	run := func(ctx context.Context) error {
		start := time.Now()
		report, err := sync.Run(ctx)
		attrs := []any{slog.Int("files", report.Files), slog.Int("loaded", report.Loaded), slog.Int("failed", report.Failed), slog.Duration("took", time.Since(start))}
		if err != nil {
			logger.ErrorContext(ctx, "sync failed", append(attrs, slog.Any("error", err))...)
			return err
		}
		logger.InfoContext(ctx, "sync done", attrs...)
		return nil
	}
	if cfg.SyncInterval == 0 {
		return run(ctx)
	}

	probe := health.New(pool.Ping)
	admin := http.NewServeMux()
	admin.HandleFunc("GET /livez", probe.Live)
	admin.HandleFunc("GET /readyz", probe.Ready)
	admin.Handle("GET /metrics", tel.Metrics)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return httpserver.Run(gctx, httpserver.New(cfg.AdminAddr, admin, logger), cfg.ShutdownTimeout)
	})
	g.Go(func() error {
		// A failed run is logged and tried again at the next tick.
		_ = run(gctx)
		return schedule.Every(gctx, cfg.SyncInterval, func(ctx context.Context) { _ = run(ctx) })
	})
	probe.SetReady(true)
	logger.Info("syncing", slog.String("from", cfg.ONSURL), slog.Duration("every", cfg.SyncInterval), slog.String("version", version))
	if err := g.Wait(); err != nil {
		return err
	}
	logger.Info("stopped")
	return nil
}

func migrate(ctx context.Context, getenv func(string) string, stdout io.Writer) error {
	cfg, logger, err := setup(getenv, stdout)
	if err != nil {
		return err
	}
	pool, err := postgres.Connect(ctx, postgres.PoolConfig{URL: cfg.DatabaseURL, MaxConns: 2})
	if err != nil {
		return err
	}
	defer pool.Close()
	return postgres.Migrate(ctx, pool, logger)
}

// healthcheck asks this host's admin port whether the server is ready. The
// image is distroless, with no curl to do it.
func healthcheck(ctx context.Context, getenv func(string) string) error {
	addr := getenv("ADMIN_ADDR")
	if addr == "" {
		addr = ":9090"
	}
	if addr[0] == ':' {
		addr = "127.0.0.1" + addr
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+"/readyz", nil)
	if err != nil {
		return fmt.Errorf("ask %s: %w", addr, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ask %s: %w", addr, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %s", addr, res.Status)
	}
	return nil
}
