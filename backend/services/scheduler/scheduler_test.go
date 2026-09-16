package scheduler

import (
	"testing"
	"time"

	atypes "main/types/absence"
)

type fakeStore struct {
	rolloverYears  []int
	monthlyPeriods []string
	expirations    int
}

func (f *fakeStore) RolloverYear(year int) (*atypes.RolloverReport, error) {
	f.rolloverYears = append(f.rolloverYears, year)
	return &atypes.RolloverReport{Year: year}, nil
}

func (f *fakeStore) RunMonthlyAccrual(period string) (*atypes.RolloverReport, error) {
	f.monthlyPeriods = append(f.monthlyPeriods, period)
	return &atypes.RolloverReport{}, nil
}

func (f *fakeStore) RunExpiration() (*atypes.RolloverReport, error) {
	f.expirations++
	return &atypes.RolloverReport{}, nil
}

func TestRunOnceInvokesAllJobs(t *testing.T) {
	f := &fakeStore{}
	s := New(f, time.Hour)

	s.RunOnce(time.Date(2026, 9, 12, 3, 0, 0, 0, time.UTC))

	if len(f.rolloverYears) != 1 || f.rolloverYears[0] != 2026 {
		t.Fatalf("rollover years = %v, want [2026]", f.rolloverYears)
	}
	if len(f.monthlyPeriods) != 1 || f.monthlyPeriods[0] != "2026-09" {
		t.Fatalf("monthly periods = %v, want [2026-09]", f.monthlyPeriods)
	}
	if f.expirations != 1 {
		t.Fatalf("expirations = %d, want 1", f.expirations)
	}
}

func TestNewDefaultsInterval(t *testing.T) {
	if s := New(&fakeStore{}, 0); s.interval != time.Hour {
		t.Fatalf("zero interval = %s, want 1h", s.interval)
	}
	if s := New(&fakeStore{}, -time.Minute); s.interval != time.Hour {
		t.Fatalf("negative interval = %s, want 1h", s.interval)
	}
	if s := New(&fakeStore{}, 30*time.Second); s.interval != 30*time.Second {
		t.Fatalf("explicit interval = %s, want 30s", s.interval)
	}
}
