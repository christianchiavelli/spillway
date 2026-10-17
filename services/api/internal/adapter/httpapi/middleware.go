package httpapi

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/httpapi/oas"
	"github.com/christianchiavelli/spillway/services/api/internal/platform/logging"
)

// middleware wraps a handler. Plain net/http: nothing here needs a framework.
type middleware func(http.Handler) http.Handler

// chain applies mws so the first one is the outermost: it sees the request
// first and the response last.
func chain(h http.Handler, mws ...middleware) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}
	return h
}

// recoverPanics turns a panic in a handler into a logged 500 problem, instead
// of the dropped connection net/http gives on its own.
func recoverPanics(logger *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer answerPanic(r.Context(), logger, w)
			next.ServeHTTP(w, r)
		})
	}
}

// answerPanic must be what is deferred, since recover only stops a panic when
// the deferred function calls it itself.
func answerPanic(ctx context.Context, logger *slog.Logger, w http.ResponseWriter) {
	v := recover()
	if v == nil {
		return
	}
	if err, ok := v.(error); ok && errors.Is(err, http.ErrAbortHandler) {
		panic(v) // the way to abort a response on purpose; let net/http do it
	}
	logger.ErrorContext(ctx, "panic serving a request", slog.Any("panic", v), slog.String("stack", string(debug.Stack())))
	writeProblem(w, problem(http.StatusInternalServerError, oas.ProblemCodeInternal, ""))
}

// requestID gives every request an id, kept from X-Request-Id when the web
// server sent one, so its log lines and ours can be read together.
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" || len(id) > 64 {
			id = rand.Text()
		}
		w.Header().Set("X-Request-Id", id)
		ctx := logging.WithAttrs(r.Context(), slog.String("request_id", id))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// accessLog writes one line per request, after it is answered.
func accessLog(logger *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			level := slog.LevelInfo
			if rec.status >= 500 {
				level = slog.LevelError
			}
			logger.Log(r.Context(), level, "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int64("bytes", rec.bytes),
				slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000))
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int64
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status, r.wroteHeader = status, true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	n, err := r.ResponseWriter.Write(b)
	r.bytes += int64(n)
	return n, err
}

// Unwrap lets http.ResponseController reach the real writer.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// apiHeaders are sent with every answer. The API serves data, never a page,
// so nothing it returns may be framed, sniffed or kept in a shared cache.
func apiHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Cache-Control", "no-store")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// crossOrigin refuses requests from browsers on origins not in allowed, which
// is empty by default: only the web server, which sends no Origin, may call.
// An allowed origin gets the CORS headers a browser needs, and no more.
func crossOrigin(allowed []string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Add("Vary", "Origin")
			if !slices.Contains(allowed, origin) {
				writeProblem(w, problem(http.StatusForbidden, oas.ProblemCodeForbiddenOrigin, "Browsers may not call this API from that origin."))
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// limitBody caps how much of a body the decoder reads.
func limitBody(max int64) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, max)
			next.ServeHTTP(w, r)
		})
	}
}

// deadline gives every request a deadline its queries inherit, so a slow
// database cannot pile up requests forever.
func deadline(d time.Duration) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
