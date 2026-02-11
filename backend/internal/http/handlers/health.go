package handlers

import (
	"net/http"

	"sfa/backend/internal/store"
)

type HealthHandler struct {
	Store *store.Store
}

func NewHealthHandler(store *store.Store) HealthHandler {
	return HealthHandler{Store: store}
}

func (h HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "degraded",
			"error":  "database_unreachable",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = jsonEncoder(w, payload)
}
