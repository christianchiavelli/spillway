// Package config reads the service's settings from the environment, checks
// them all at once and fails before anything starts when one is wrong.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config is every setting, with its environment variable and default.
type Config struct {
	APIAddr          string        // API_ADDR, ":8080"
	AdminAddr        string        // ADMIN_ADDR, ":9090": health and metrics, never published
	DatabaseURL      string        // DATABASE_URL, required
	DatabaseMaxConns int32         // DATABASE_MAX_CONNS, 10
	LogFormat        string        // LOG_FORMAT, "json"
	LogLevel         slog.Level    // LOG_LEVEL, "info"
	AllowedOrigins   []string      // CORS_ALLOWED_ORIGINS, comma separated, none
	RequestTimeout   time.Duration // REQUEST_TIMEOUT, 10s
	ShutdownDrain    time.Duration // SHUTDOWN_DRAIN, 0s: how long to answer "not ready" before stopping
	ShutdownTimeout  time.Duration // SHUTDOWN_TIMEOUT, 20s: how long requests in flight get to finish
	HashConcurrency  int           // PASSWORD_HASH_CONCURRENCY, the number of CPUs
	OTLPEndpoint     string        // OTEL_EXPORTER_OTLP_ENDPOINT, none: traces are exported only when set
}

// Load reads the environment through getenv, which tests replace with a map.
func Load(getenv func(string) string) (Config, error) {
	r := reader{getenv: getenv}
	c := Config{
		APIAddr:          r.str("API_ADDR", ":8080"),
		AdminAddr:        r.str("ADMIN_ADDR", ":9090"),
		DatabaseURL:      r.required("DATABASE_URL"),
		DatabaseMaxConns: r.int32("DATABASE_MAX_CONNS", 10, 1, 1000),
		LogFormat:        r.oneOf("LOG_FORMAT", "json", "json", "text"),
		LogLevel:         r.level("LOG_LEVEL", slog.LevelInfo),
		AllowedOrigins:   r.list("CORS_ALLOWED_ORIGINS"),
		RequestTimeout:   r.duration("REQUEST_TIMEOUT", 10*time.Second),
		ShutdownDrain:    r.duration("SHUTDOWN_DRAIN", 0),
		ShutdownTimeout:  r.duration("SHUTDOWN_TIMEOUT", 20*time.Second),
		HashConcurrency:  r.integer("PASSWORD_HASH_CONCURRENCY", runtime.NumCPU(), 1, 1024),
		OTLPEndpoint:     r.str("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
	}
	for _, origin := range c.AllowedOrigins {
		if origin == "*" || !strings.Contains(origin, "://") || strings.HasSuffix(origin, "/") {
			r.errs = append(r.errs, fmt.Errorf("CORS_ALLOWED_ORIGINS: %q is not an origin like https://example.com", origin))
		}
	}
	return c, errors.Join(r.errs...)
}

// reader collects every problem instead of stopping at the first, so one
// failed start lists everything to fix.
type reader struct {
	getenv func(string) string
	errs   []error
}

func (r *reader) str(key, fallback string) string {
	if v := strings.TrimSpace(r.getenv(key)); v != "" {
		return v
	}
	return fallback
}

func (r *reader) required(key string) string {
	v := r.str(key, "")
	if v == "" {
		r.errs = append(r.errs, fmt.Errorf("%s: required", key))
	}
	return v
}

func (r *reader) integer(key string, fallback, lo, hi int) int {
	v := r.str(key, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < lo || n > hi {
		r.errs = append(r.errs, fmt.Errorf("%s: %q is not a whole number from %d to %d", key, v, lo, hi))
		return fallback
	}
	return n
}

func (r *reader) int32(key string, fallback, lo, hi int32) int32 {
	v := r.str(key, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil || n < int64(lo) || n > int64(hi) {
		r.errs = append(r.errs, fmt.Errorf("%s: %q is not a whole number from %d to %d", key, v, lo, hi))
		return fallback
	}
	return int32(n)
}

func (r *reader) duration(key string, fallback time.Duration) time.Duration {
	v := r.str(key, "")
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d < 0 {
		r.errs = append(r.errs, fmt.Errorf("%s: %q is not a duration like 15s", key, v))
		return fallback
	}
	return d
}

func (r *reader) oneOf(key, fallback string, allowed ...string) string {
	v := strings.ToLower(r.str(key, fallback))
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	r.errs = append(r.errs, fmt.Errorf("%s: %q is not one of %s", key, v, strings.Join(allowed, ", ")))
	return fallback
}

func (r *reader) level(key string, fallback slog.Level) slog.Level {
	v := r.str(key, "")
	if v == "" {
		return fallback
	}
	var l slog.Level
	if err := l.UnmarshalText([]byte(v)); err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s: %q is not debug, info, warn or error", key, v))
		return fallback
	}
	return l
}

func (r *reader) list(key string) []string {
	var out []string
	for item := range strings.SplitSeq(r.str(key, ""), ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}
