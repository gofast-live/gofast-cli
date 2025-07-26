package rest

import (
	"app/pkg"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"service-core/config"
)

func Run(h *Handler) {
	cfg := h.Cfg
	mux := http.NewServeMux()
	// Users
	mux.HandleFunc("/users", h.getAllUsers)

	// Cron jobs
	mux.HandleFunc("/tasks/delete-tokens", h.handleTasksDeleteTokens)
	// Health checks
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		_, err := h.Store.Ping(r.Context())
		if err != nil {
			slog.Error("Error pinging database", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("OK"))
		if err != nil {
			panic(err)
		}
	})

	handler := loggingMiddleware(mux)
	go func() {
		slog.Info("HTTP server listening on", "port", cfg.HTTPPort)
		server := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: handler, ReadHeaderTimeout: cfg.HTTPTimeout, WriteTimeout: cfg.HTTPTimeout}
		err := server.ListenAndServe()
		if err != nil {
			slog.Error("Error serving HTTP", "error", err)
			panic(err)
		}
	}()
}

func extractAccessToken(r *http.Request) string {
	token, err := r.Cookie("access_token")
	if err != nil {
		return r.Header.Get("Authorization")
	}
	return token.Value
}

func writeResponse(cfg *config.Config, w http.ResponseWriter, r *http.Request, data any, err error) {
	w.Header().Set("Access-Control-Allow-Origin", cfg.ClientURL)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	if err != nil {
		var unauthorizedError pkg.UnauthorizedError
		var internalError pkg.InternalError
		var badRequestError pkg.BadRequestError
		var notFoundError pkg.NotFoundError
		var validationErrors pkg.ValidationErrors
		switch {
		case errors.As(err, &unauthorizedError):
			slog.Error("Unauthorized", "error", err)
			returnURL := r.FormValue("return_url")
			if returnURL == "" {
				http.Error(w, unauthorizedError.Error(), http.StatusUnauthorized)
			} else {
				http.Redirect(w, r, returnURL+"/login?error=unauthorized", http.StatusSeeOther)
			}
			return
		case errors.As(err, &internalError):
			slog.Error("Internal error", "error", internalError)
			http.Error(w, internalError.Message, http.StatusInternalServerError)
			return
		case errors.As(err, &badRequestError):
			slog.Error("Bad request error", "error", badRequestError)
			http.Error(w, badRequestError.Message, http.StatusBadRequest)
			return
		case errors.As(err, &notFoundError):
			slog.Error("Not found error", "error", notFoundError)
			http.Error(w, notFoundError.Message, http.StatusNotFound)
			return
		case errors.As(err, &validationErrors):
			slog.Error("Validation error", "error", validationErrors)
			http.Error(w, validationErrors.Error(), http.StatusUnprocessableEntity)
			return
		default:
			slog.Error("Error", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if data == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("Error writing response", "error", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}
