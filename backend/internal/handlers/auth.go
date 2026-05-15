package handlers

import (
	"net/http"

	"fixy-backend/internal/middleware"
)

func requireRole(w http.ResponseWriter, r *http.Request, allowed ...middleware.Role) (middleware.Role, bool) {
	role := middleware.RoleFromContext(r.Context())
	if role == middleware.RoleNone {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return role, false
	}
	if !middleware.RoleAllowed(role, allowed...) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return role, false
	}
	return role, true
}
