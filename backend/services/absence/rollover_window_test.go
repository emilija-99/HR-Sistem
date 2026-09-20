package absence

import (
	"testing"
	"time"
)

func TestRolloverWindow(t *testing.T) {
	cases := []struct {
		name        string
		now         time.Time
		wantYear    int
		wantAllowed bool
	}{
		{"dec 25 targets current year", time.Date(2026, 12, 25, 10, 0, 0, 0, time.UTC), 2026, true},
		{"dec 31 targets current year", time.Date(2026, 12, 31, 23, 59, 0, 0, time.UTC), 2026, true},
		{"dec 24 not allowed", time.Date(2026, 12, 24, 12, 0, 0, 0, time.UTC), 0, false},
		{"jan 1 targets previous year", time.Date(2027, 1, 1, 8, 0, 0, 0, time.UTC), 2026, true},
		{"jan 7 targets previous year", time.Date(2027, 1, 7, 20, 0, 0, 0, time.UTC), 2026, true},
		{"jan 8 not allowed", time.Date(2027, 1, 8, 0, 1, 0, 0, time.UTC), 0, false},
		{"november not allowed", time.Date(2026, 11, 30, 12, 0, 0, 0, time.UTC), 0, false},
		{"june not allowed", time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC), 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			year, allowed := rolloverWindow(tc.now)
			if allowed != tc.wantAllowed || year != tc.wantYear {
				t.Fatalf("rolloverWindow(%s) = (%d, %v), want (%d, %v)",
					tc.now.Format("2006-01-02"), year, allowed, tc.wantYear, tc.wantAllowed)
			}
		})
	}
}
