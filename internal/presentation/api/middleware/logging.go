package middleware

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
// It also implements http.Hijacker for WebSocket support
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack implements http.Hijacker for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("hijack not supported")
	}
	return h.Hijack()
}

// Logging middleware logs HTTP requests
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := newResponseWriter(w)

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Log request details
		duration := time.Since(start)
		log.Printf(
			"%s %s %d %s",
			r.Method,
			r.RequestURI,
			wrapped.statusCode,
			duration,
		)
	})
}
