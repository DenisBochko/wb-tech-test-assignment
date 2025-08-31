package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Logger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			latency := time.Since(start).Microseconds()
			log.Info("request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("latency", fmt.Sprintf("%d µs", latency)),
			)
		})
	}
}
