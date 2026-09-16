// Package scheduler periodically runs leave-related maintenance jobs:
//
//   - annual accrual + carry-over (reuses the idempotent RolloverYear logic)
//   - monthly accrual for MONTHLY_GRANT policies
//   - expiry of balances that passed their expiry date
//
// Every job is idempotent, so the scheduler is safe to run in every instance
// and to restart: repeating a tick never grants or expires days twice.
package scheduler

import (
	"context"
	"log"
	"time"

	atypes "main/types/absence"
)

// JobStore is the subset of the absence store the scheduler needs.
type JobStore interface {
	RolloverYear(year int) (*atypes.RolloverReport, error)
	RunMonthlyAccrual(period string) (*atypes.RolloverReport, error)
	RunExpiration() (*atypes.RolloverReport, error)
}

type Scheduler struct {
	store    JobStore
	interval time.Duration
}

// New creates a scheduler. A non-positive interval defaults to one hour.
func New(store JobStore, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = time.Hour
	}
	return &Scheduler{store: store, interval: interval}
}

// Start runs the jobs once immediately, then on every tick until ctx is done.
func (s *Scheduler) Start(ctx context.Context) {
	log.Printf("scheduler: started (interval %s)", s.interval)
	s.RunOnce(time.Now())

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler: stopped")
			return
		case t := <-ticker.C:
			s.RunOnce(t)
		}
	}
}

// RunOnce executes all jobs. Each job is idempotent, so calling it more often
// than necessary is harmless.
func (s *Scheduler) RunOnce(now time.Time) {
	year := now.Year()

	if rep, err := s.store.RolloverYear(year); err != nil {
		log.Printf("scheduler: annual rollover %d failed: %v", year, err)
	} else {
		log.Printf("scheduler: annual rollover %d accrued=%d carried=%d expired=%d skipped=%d",
			year, len(rep.Accrued), len(rep.Carried), len(rep.Expired), rep.SkippedCnt)
	}

	period := now.Format("2006-01")
	if rep, err := s.store.RunMonthlyAccrual(period); err != nil {
		log.Printf("scheduler: monthly accrual %s failed: %v", period, err)
	} else {
		log.Printf("scheduler: monthly accrual %s granted=%d skipped=%d",
			period, len(rep.Accrued), rep.SkippedCnt)
	}

	if rep, err := s.store.RunExpiration(); err != nil {
		log.Printf("scheduler: expiration failed: %v", err)
	} else {
		log.Printf("scheduler: expiration expired=%d skipped=%d", len(rep.Expired), rep.SkippedCnt)
	}
}
