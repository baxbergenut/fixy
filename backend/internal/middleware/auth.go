package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type Role string

const (
	RoleNone         Role = ""
	RoleAdmin        Role = "admin"
	RoleAccountant   Role = "accountant"
	RoleFleetManager Role = "fleet_manager"
)

type AuthTokens struct {
	Admin        string
	Accountant   string
	FleetManager string
}

type roleKey struct{}

func RoleFromContext(ctx context.Context) Role {
	role, ok := ctx.Value(roleKey{}).(Role)
	if !ok {
		return RoleNone
	}
	return role
}

func RoleAllowed(role Role, allowed ...Role) bool {
	for _, candidate := range allowed {
		if role == candidate {
			return true
		}
	}
	return false
}

func Auth(tokens AuthTokens) func(http.Handler) http.Handler {
	tokenRoles := map[string]Role{}

	if trimmed := strings.TrimSpace(tokens.Admin); trimmed != "" {
		tokenRoles[trimmed] = RoleAdmin
	}
	if trimmed := strings.TrimSpace(tokens.Accountant); trimmed != "" {
		tokenRoles[trimmed] = RoleAccountant
	}
	if trimmed := strings.TrimSpace(tokens.FleetManager); trimmed != "" {
		tokenRoles[trimmed] = RoleFleetManager
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			role := resolveRole(r.Header.Get("Authorization"), tokenRoles)
			if role == RoleNone {
				if isReadMethod(r.Method) {
					next.ServeHTTP(w, r)
					return
				}

				writeUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), roleKey{}, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func resolveRole(header string, tokenRoles map[string]Role) Role {
	token := strings.TrimSpace(header)
	if token == "" {
		return RoleNone
	}

	lowered := strings.ToLower(token)
	if strings.HasPrefix(lowered, "bearer ") {
		token = strings.TrimSpace(token[7:])
	}

	if token == "" {
		return RoleNone
	}

	role, ok := tokenRoles[token]
	if !ok {
		return RoleNone
	}

	return role
}

func isReadMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}
