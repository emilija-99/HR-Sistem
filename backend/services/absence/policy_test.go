package absence

import (
	types "main/types/absence"
	"testing"
)

func intPtr(i int) *int { return &i }

func TestExpiryForPolicy(t *testing.T) {
	year := 2026

	t.Run("without carry-over expires end of same year", func(t *testing.T) {
		p := &types.LeavePolicy{AllowCarryOver: false}
		if got := expiryForPolicy(p, year); got != "2026-12-31" {
			t.Fatalf("got %s, want 2026-12-31", got)
		}
	})

	t.Run("with carry-over extends to policy expiry next year", func(t *testing.T) {
		p := &types.LeavePolicy{
			AllowCarryOver:       true,
			CarryOverExpiryMonth: intPtr(6),
			CarryOverExpiryDay:   intPtr(30),
		}
		if got := expiryForPolicy(p, year); got != "2027-06-30" {
			t.Fatalf("got %s, want 2027-06-30", got)
		}
	})

	t.Run("carry-over flag without dates stays same year", func(t *testing.T) {
		p := &types.LeavePolicy{AllowCarryOver: true}
		if got := expiryForPolicy(p, year); got != "2026-12-31" {
			t.Fatalf("got %s, want 2026-12-31", got)
		}
	})
}

func TestIntVal(t *testing.T) {
	if got := intVal(nil, 30); got != 30 {
		t.Fatalf("nil fallback = %d, want 30", got)
	}
	if got := intVal(intPtr(15), 30); got != 15 {
		t.Fatalf("value = %d, want 15", got)
	}
}
