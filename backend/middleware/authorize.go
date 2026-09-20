package middleware

import (
	"database/sql"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// RoleFromContext reads the user_id and role claims stored by JWTAuth.
func RoleFromContext(r *http.Request) (userID uint, role string, ok bool) {
	claims, okClaims := r.Context().Value(UserContextKey).(jwt.MapClaims)
	if !okClaims {
		return 0, "", false
	}
	uid, ok := claims["user_id"].(float64)
	if !ok {
		return 0, "", false
	}
	role, _ = claims["role"].(string)
	return uint(uid), role, role != ""
}

// RequirePermission returns middleware that allows the request only if the
// authenticated role holds the given permission code in role_permissions.
func RequirePermission(db *sql.DB, code string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, role, ok := RoleFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var has bool
		err := db.QueryRow(`SELECT EXISTS(
			SELECT 1 FROM role_permissions rp
			JOIN roles r ON r.id = rp.role_id
			JOIN permissions p ON p.id = rp.permission_id
			WHERE r.name = $1 AND p.code = $2
		)`, role, code).Scan(&has)
		if err != nil || !has {
			http.Error(w, "Forbidden: missing permission "+code, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
