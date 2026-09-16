package absence

import "testing"

func TestBusinessDays(t *testing.T) {
	holiday := map[string]bool{"2026-05-01": true} // Friday

	cases := []struct {
		name    string
		start   string
		end     string
		hol     map[string]bool
		want    float64
		wantErr bool
	}{
		{name: "mon to fri", start: "2026-06-01", end: "2026-06-05", want: 5},
		{name: "single workday", start: "2026-06-01", end: "2026-06-01", want: 1},
		{name: "fri to mon skips weekend", start: "2026-06-05", end: "2026-06-08", want: 2},
		{name: "full week incl weekend", start: "2026-06-01", end: "2026-06-07", want: 5},
		{name: "holiday skipped", start: "2026-05-01", end: "2026-05-04", hol: holiday, want: 1},
		{name: "weekend only", start: "2026-07-04", end: "2026-07-05", want: 0},
		{name: "reversed dates", start: "2026-06-05", end: "2026-06-01", wantErr: true},
		{name: "invalid start", start: "nope", end: "2026-06-01", wantErr: true},
		{name: "invalid end", start: "2026-06-01", end: "2026-13-40", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := businessDays(tc.start, tc.end, tc.hol)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("businessDays(%s..%s) = %v, want %v", tc.start, tc.end, got, tc.want)
			}
		})
	}
}

func TestYearOf(t *testing.T) {
	if got := yearOf("2026-09-12"); got != 2026 {
		t.Fatalf("yearOf = %d, want 2026", got)
	}
	// invalid input falls back to the current year (never panics)
	if got := yearOf("garbage"); got == 0 {
		t.Fatalf("yearOf fallback should not be zero")
	}
}
