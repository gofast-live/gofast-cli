package rest

import (
	"log/slog"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		var authTokenPresent bool
		accessToken := extractAccessToken(r)
		if accessToken != "" {
			authTokenPresent = true
		}

		next.ServeHTTP(w, r)
		duration := time.Since(startTime)

		if r.URL.Path == "/ready" {
			return
		}

		slog.Info("HTTP call",
			slog.String("http_method", r.Method),
			slog.String("http_path", r.URL.Path),
			slog.String("duration", duration.String()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.Bool("auth_token_present", authTokenPresent),
		)
	})
}

