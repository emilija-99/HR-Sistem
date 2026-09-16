-- Idempotency markers for scheduled jobs, so a job runs at most once per
-- (job, period, subject) — e.g. monthly accrual per employee+absence type.
CREATE TABLE IF NOT EXISTS scheduler_runs (
    job     TEXT NOT NULL,
    period  TEXT NOT NULL,
    subject TEXT NOT NULL DEFAULT '',
    run_at  TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (job, period, subject)
);
