package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateTokenClaimsAndTTL(t *testing.T) {
	tokenString, err := GenerateToken(7, "test@hr-sistem.com", "EMPLOYEE")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(tk *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("token should be valid, err=%v", err)
	}

	if claims.UserID != 7 || claims.Email != "test@hr-sistem.com" || claims.Role != "EMPLOYEE" {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	if accessTokenTTL != 20*time.Minute {
		t.Fatalf("accessTokenTTL = %s, want 20m", accessTokenTTL)
	}

	// Dužina sesije je ograničena: posle `SessionTTL` od prijave obavezna je
	// ponovna prijava (refresh_tokens.expires_at se postavlja iz ovog trajanja).
	if SessionTTL != 8*time.Hour {
		t.Fatalf("SessionTTL = %s, want 8h", SessionTTL)
	}

	ttl := claims.ExpiresAt.Sub(claims.IssuedAt.Time)
	if ttl != accessTokenTTL {
		t.Fatalf("token TTL = %s, want %s", ttl, accessTokenTTL)
	}
}
