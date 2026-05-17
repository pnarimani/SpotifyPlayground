package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/felixge/httpsnoop"
	"github.com/google/uuid"
)

type ContextKey string

const RequestIdKey ContextKey = "requestId"

func Wrap(mux http.Handler) http.Handler {
	return requestId(logging(panicRecovery(mux)))
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := r.Context().Value(RequestIdKey).(string)

		slog.Debug("request started", "id", id, "method", r.Method, "url", r.URL.Path, "remote_addr", getClientIP(r))

		m := httpsnoop.CaptureMetrics(next, w, r)

		var logLevel slog.Level
		switch {
		case m.Code >= 200 && m.Code < 300:
			logLevel = slog.LevelDebug
		case m.Code < 500:
			logLevel = slog.LevelInfo
		default:
			logLevel = slog.LevelError
		}

		slog.Log(r.Context(), logLevel, "request completed", "id", id, "method", r.Method, "code", m.Code, "duration", m.Duration, "response_length", m.Written)
	})
}

func requestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")

		if !isRequestIdValid(id) {
			id = uuid.New().String()
		}

		w.Header().Set("X-Request-Id", id)

		ctx := context.WithValue(r.Context(), RequestIdKey, id)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)

				id, _ := r.Context().Value(RequestIdKey).(string)
				slog.ErrorContext(
					r.Context(), "request panic",
					"id", id,
					"method", r.Method,
					"url", r.URL.Path,
					"remote_addr", getClientIP(r),
					"stack_trace", string(debug.Stack()),
					"panic_value", rec,
					"panic_type", fmt.Sprintf("%T", rec),
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func isRequestIdValid(id string) bool {
	// TODO: we should probably also check for allowed charset
	return id != "" && len(id) <= 128
}

func getClientIP(r *http.Request) string {
	// TODO: read from forwarded-for and real-ip
	return r.RemoteAddr
}
