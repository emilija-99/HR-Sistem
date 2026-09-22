package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("super-secret-key")

type contextKey string

const UserContextKey = contextKey("user")

// JWTAuth validates the Bearer token and rejects tokens issued to deactivated
// accounts. It stores the parsed claims in the request context.
func JWTAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Nedostaje token.", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			// jwt.Parse(..) - validates exp and nbf
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				// Only accept HMAC-signed tokens
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return secret, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Nevažeći token.", http.StatusUnauthorized)
				return
			}
			// validation using generic map - user_id, email and role.
			// user_id is float 64 because JSON do not have int
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Nevažeći podaci u tokenu.", http.StatusUnauthorized)
				return
			}

			uid, ok := claims["user_id"].(float64)
			if !ok {
				http.Error(w, "Nevažeći podaci u tokenu.", http.StatusUnauthorized)
				return
			}

			// Reject requests from accounts that have been deactivated, and refresh
			// the role from the database so permission changes take effect
			// immediately (the token's role can be stale).
			var active bool
			var role sql.NullString
			// A user may have several roles; pick the highest-privilege one so the
			// effective role is deterministic (highest privilege wins).
			err = db.QueryRow(`
				SELECT u.is_active, r.name
				FROM users u
				LEFT JOIN user_roles ur ON ur.user_id = u.id
				LEFT JOIN roles r ON r.id = ur.role_id
				WHERE u.id = $1
				ORDER BY CASE r.name
					WHEN 'PLATFORM_ADMIN'        THEN 1
					WHEN 'HR_ADMIN'              THEN 2
					WHEN 'MANAGER_PORTAL_ACCESS' THEN 3
					WHEN 'EMPLOYEE'              THEN 4
					ELSE 5
				END
				LIMIT 1`, uint(uid)).Scan(&active, &role)
			if err != nil {
				http.Error(w, "Nevažeći token.", http.StatusUnauthorized)
				return
			}
			if !active {
				http.Error(w, "Nalog je deaktiviran.", http.StatusForbidden)
				return
			}
			claims["role"] = role.String

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
