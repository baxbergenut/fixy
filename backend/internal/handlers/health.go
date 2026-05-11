package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

func Health(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := healthResponse{
			Status:    "ok",
			Database:  "not_configured",
			CheckedAt: time.Now().UTC(),
		}

		if database != nil {
			if err := database.PingContext(r.Context()); err != nil {
				response.Status = "degraded"
				response.Database = "down"
				writeJSON(w, http.StatusServiceUnavailable, response)
				return
			}
			response.Database = "up"
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
