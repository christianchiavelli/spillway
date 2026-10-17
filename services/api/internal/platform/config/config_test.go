package config_test

import (
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/platform/config"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDefaults(t *testing.T) {
	c, err := config.Load(env(map[string]string{"DATABASE_URL": "postgres://db/spillway"}))
	if err != nil {
		t.Fatal(err)
	}
	if c.APIAddr != ":8080" || c.AdminAddr != ":9090" || c.DatabaseMaxConns != 10 || c.LogFormat != "json" ||
		c.LogLevel != slog.LevelInfo || c.AllowedOrigins != nil || c.RequestTimeout != 10*time.Second ||
		c.ShutdownDrain != 0 || c.ShutdownTimeout != 20*time.Second || c.HashConcurrency < 1 || c.OTLPEndpoint != "" {
		t.Errorf("defaults = %+v", c)
	}
}

func TestEverySettingIsRead(t *testing.T) {
	c, err := config.Load(env(map[string]string{
		"API_ADDR":                    "127.0.0.1:8085",
		"ADMIN_ADDR":                  "127.0.0.1:9095",
		"DATABASE_URL":                "postgres://db/spillway",
		"DATABASE_MAX_CONNS":          "25",
		"LOG_FORMAT":                  "TEXT",
		"LOG_LEVEL":                   "debug",
		"CORS_ALLOWED_ORIGINS":        " http://localhost:5173 , https://spillway.example ",
		"REQUEST_TIMEOUT":             "3s",
		"SHUTDOWN_DRAIN":              "5s",
		"SHUTDOWN_TIMEOUT":            "1m",
		"PASSWORD_HASH_CONCURRENCY":   "2",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://lgtm:4318",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.APIAddr != "127.0.0.1:8085" || c.DatabaseMaxConns != 25 || c.LogFormat != "text" || c.LogLevel != slog.LevelDebug ||
		!slices.Equal(c.AllowedOrigins, []string{"http://localhost:5173", "https://spillway.example"}) ||
		c.RequestTimeout != 3*time.Second || c.ShutdownDrain != 5*time.Second || c.ShutdownTimeout != time.Minute ||
		c.HashConcurrency != 2 || c.OTLPEndpoint != "http://lgtm:4318" {
		t.Errorf("config = %+v", c)
	}
}

func TestEveryProblemIsReportedAtOnce(t *testing.T) {
	_, err := config.Load(env(map[string]string{
		"DATABASE_MAX_CONNS":   "0",
		"LOG_FORMAT":           "xml",
		"LOG_LEVEL":            "loud",
		"REQUEST_TIMEOUT":      "ten seconds",
		"CORS_ALLOWED_ORIGINS": "*",
	}))
	if err == nil {
		t.Fatal("Load() accepted a broken environment")
	}
	for _, key := range []string{"DATABASE_URL", "DATABASE_MAX_CONNS", "LOG_FORMAT", "LOG_LEVEL", "REQUEST_TIMEOUT", "CORS_ALLOWED_ORIGINS"} {
		if !strings.Contains(err.Error(), key+":") {
			t.Errorf("the error does not mention %s:\n%v", key, err)
		}
	}
}
