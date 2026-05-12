package middleware

import (
	"net/http"
	"strings"
)

var allowedOrigins = map[string]struct{}{
	"http://localhost:3000": {},
	"http://localhost:3001": {},
	"http://127.0.0.1:3000": {},
	"http://127.0.0.1:3001": {},
	"http://localhost:3002": {},
	"http://127.0.0.1:3002": {},
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" {
			if _, ok := allowedOrigins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
