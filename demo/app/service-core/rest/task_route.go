package rest

import (
	"log/slog"
	"net/http"
)

func (h *Handler) handleTasksDeleteTokens(w http.ResponseWriter, r *http.Request) {
	slog.Info("Running Task: Delete Tokens")
	apiKey := r.Header.Get("X-Api-Key")
	if apiKey != h.Cfg.TaskToken {
		slog.Error("Invalid API key")
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}
	err := h.Store.DeleteTokens(r.Context())
	if err != nil {
		slog.Error("Error deleting tokens", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
