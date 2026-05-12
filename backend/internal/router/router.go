package router

import (
	"database/sql"
	"net/http"

	"fixy-backend/internal/handlers"
)

func New(database *sql.DB, groqToken string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handlers.Health(database))
	mux.HandleFunc("/api/health", handlers.Health(database))
	mux.Handle("/api/invoice/parse", handlers.NewInvoiceHandler(groqToken))
	mux.Handle("/api/maintenance", handlers.NewMaintenanceHandler(database))
	mux.Handle("/api/maintenance/", handlers.NewMaintenanceHandler(database))
	mux.Handle("/api/trucks", handlers.NewTrucksHandler(database))
	mux.Handle("/api/trucks/", handlers.NewTrucksHandler(database))
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
