package absence

import (
	"testing"

	types "main/types/absence"
)

// Regression: absence types without any configured policy (TRAINING, PERSONAL,
// DISABILITY) used to skip the balance check entirely, so an employee could
// request an unlimited number of days (e.g. 151) with no balance at all.
func TestBalanceRequiredFailsClosed(t *testing.T) {
	cases := []struct {
		name   string
		policy *types.LeavePolicy
		want   bool
	}{
		{
			name:   "no policy configured → days must be covered",
			policy: nil,
			want:   true,
		},
		{
			name:   "policy requires balance",
			policy: &types.LeavePolicy{Name: "Vacation standard", RequiresBalance: true},
			want:   true,
		},
		{
			name:   "policy explicitly unlimited (e.g. Sick unlimited)",
			policy: &types.LeavePolicy{Name: "Sick unlimited", RequiresBalance: false},
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := balanceRequired(tc.policy); got != tc.want {
				t.Fatalf("balanceRequired = %v, want %v", got, tc.want)
			}
		})
	}
}
