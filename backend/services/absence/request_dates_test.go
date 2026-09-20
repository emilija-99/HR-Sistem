package absence

import (
	"testing"
	"time"
)

// today is Wednesday 2026-09-23 so the cases below are easy to read:
// 2026-09-21 Mon, 09-22 Tue, 09-23 Wed, 09-26 Sat, 09-27 Sun.
var requestDateToday = time.Date(2026, 9, 23, 15, 30, 0, 0, time.UTC)

func TestValidateRequestDates(t *testing.T) {
	cases := []struct {
		name    string
		start   string
		end     string
		wantErr bool
	}{
		{name: "today is allowed", start: "2026-09-23", end: "2026-09-23"},
		{name: "a single day (start == end) is one day off", start: "2026-09-24", end: "2026-09-24"},
		{name: "future working period", start: "2026-09-24", end: "2026-09-25"},
		{name: "period spanning a weekend", start: "2026-09-24", end: "2026-09-28"},

		{name: "yesterday is rejected", start: "2026-09-22", end: "2026-09-24", wantErr: true},
		{name: "start in the past", start: "2026-02-11", end: "2026-09-12", wantErr: true},
		{name: "end before start", start: "2026-09-25", end: "2026-09-24", wantErr: true},
		{name: "start on Saturday", start: "2026-09-26", end: "2026-09-28", wantErr: true},
		{name: "start on Sunday", start: "2026-09-27", end: "2026-09-28", wantErr: true},
		{name: "end on Saturday", start: "2026-09-24", end: "2026-09-26", wantErr: true},
		{name: "end on Sunday", start: "2026-09-24", end: "2026-09-27", wantErr: true},
		{name: "malformed start", start: "nope", end: "2026-09-24", wantErr: true},
		{name: "malformed end", start: "2026-09-24", end: "2026-13-40", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRequestDates(tc.start, tc.end, requestDateToday)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %s..%s", tc.start, tc.end)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for %s..%s: %v", tc.start, tc.end, err)
			}
		})
	}
}

func TestValidateRequestDatesChecksPastBeforeWeekend(t *testing.T) {
	// A past weekend must report the past-date problem, not the weekend one:
	// 2026-09-19 is a Saturday before `today`.
	err := validateRequestDates("2026-09-19", "2026-09-21", requestDateToday)
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := err.Error(); got != "start_date cannot be in the past" {
		t.Fatalf("unexpected message: %s", got)
	}
}
