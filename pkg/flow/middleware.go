package flow

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/google/uuid"
)

// LoggingMiddleware logs basic request info using the provided Logger.
func LoggingMiddleware(logger Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            logger.Printf("request start: %s %s", r.Method, r.URL.Path)
            next.ServeHTTP(w, r)
            logger.Printf("request complete: %s %s in %s", r.Method, r.URL.Path, time.Since(start))
        })
    }
}

// RequestIDMiddleware sets a request id header for tracing.
func RequestIDMiddleware(headerName string) Middleware {
    if headerName == "" {
        headerName = "X-Request-ID"
    }
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            id := r.Header.Get(headerName)
            if id == "" {
                id = uuid.New().String()
                r.Header.Set(headerName, id)
            }
            w.Header().Set(headerName, id)
            next.ServeHTTP(w, r)
        })
    }
}

// TimeoutMiddleware sets a per-request timeout; when the timeout elapses
// the request context will be cancelled. The handler should respect ctx.Done().
func TimeoutMiddleware(d time.Duration) Middleware {
    if d <= 0 {
        d = 0
    }
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if d <= 0 {
                next.ServeHTTP(w, r)
                return
            }
            ctx, cancel := context.WithTimeout(r.Context(), d)
            defer cancel()
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// MetricsMiddleware records simple timing metrics and sets an X-Response-Time header.
func MetricsMiddleware() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            next.ServeHTTP(w, r)
            elapsed := time.Since(start)
            w.Header().Set("X-Response-Time", fmt.Sprintf("%dms", elapsed.Milliseconds()))
        })
    }
}
