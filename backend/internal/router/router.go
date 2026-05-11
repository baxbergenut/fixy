package router

import (
	"database/sql"
	"net/http"

	"fixy-backend/internal/handlers"
)

func New(database *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handlers.Health(database))
	mux.HandleFunc("/api/health", handlers.Health(database))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"fixy backend","status":"running"}`))
	})

	return mux
}
