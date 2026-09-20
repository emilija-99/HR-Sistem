package middleware

import (
	"net/http"
	"os"
	"strings"
)

// allowedOrigins returns the list of allowed CORS origins.
// Configure via CORS_ORIGINS env.
func allowedOrigins() []string {
	if env := os.Getenv("CORS_ORIGINS"); env != "" {
		var origins []string
		for _, o := range strings.Split(env, ",") {
			if o = strings.TrimSpace(o); o != "" {
				origins = append(origins, o)
			}
		}
		return origins
	}
	return []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}
}

// originAllowed checks if the request Origin is in the allowed list.
func originAllowed(origin string) bool {
	for _, o := range allowedOrigins() {
		if o == origin {
			return true
		}
	}
	return false
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != "" && originAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
