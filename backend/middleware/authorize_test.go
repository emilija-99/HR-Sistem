package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestRoleFromContext(t *testing.T) {
	claims := jwt.MapClaims{"user_id": float64(7), "role": "HR_ADMIN"}
	ctx := context.WithValue(context.Background(), UserContextKey, claims)
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)

	id, role, ok := RoleFromContext(r)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if id != 7 {
		t.Fatalf("user id = %d, want 7", id)
	}
	if role != "HR_ADMIN" {
		t.Fatalf("role = %q, want HR_ADMIN", role)
	}
}

func TestRoleFromContextMissingClaims(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	if _, _, ok := RoleFromContext(r); ok {
		t.Fatal("expected ok=false when no claims in context")
	}
}
