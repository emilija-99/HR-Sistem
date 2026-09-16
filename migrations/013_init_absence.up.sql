-- state machine for every absence
CREATE TYPE absence_status AS ENUM ('PENDING','APPROVED','REJECTED','CANCELLED','DRAFT');

-- how days are granted for a policy
CREATE TYPE GRANT_POLICY AS ENUM ('YEARLY_GRANT','MATERNITY_GRANT','MONTHLY_GRANT','MISC');

-- how leave balance entries are categorized
CREATE TYPE LEAVE_TYPE AS ENUM ('ACCRUAL','CARRY_OVER','CONSUMED','CANCELLED','MANUAL_ADJUST','EXPIRATION');

-- rules for each absence category
CREATE TABLE leave_policies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    grant_policy GRANT_POLICY NOT NULL,
    days_per_period NUMERIC(6,2),
    period_months INT,
    allow_carry_over BOOLEAN DEFAULT FALSE,
    carry_over_max_days NUMERIC(6,2),
    carry_over_expiry_month INT,
    carry_over_expiry_day INT,
    requires_balance BOOLEAN DEFAULT TRUE
);

INSERT INTO leave_policies
(name, grant_policy, days_per_period, period_months,
 allow_carry_over, carry_over_max_days,
 carry_over_expiry_month, carry_over_expiry_day, requires_balance)
VALUES
('Vacation standard', 'YEARLY_GRANT', 20, 12,
 TRUE, 10, 6, 30, TRUE);

INSERT INTO leave_policies
(name, grant_policy, requires_balance)
VALUES
('Sick unlimited', 'MISC', FALSE);

INSERT INTO leave_policies
(name, grant_policy, days_per_period, requires_balance)
VALUES
('Parental leave', 'MATERNITY_GRANT', 365, TRUE);

-- which policy applies to an employee (with history)
CREATE TABLE employee_leave_policy (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(id),
    absence_type_id INT NOT NULL REFERENCES absence_types(id),
    policy_id BIGINT NOT NULL REFERENCES leave_policies(id),
    valid_from DATE NOT NULL,
    valid_to DATE,
    UNIQUE(employee_id, absence_type_id, valid_from)
);

-- ledger of days: positive = granted, negative = consumed
CREATE TABLE leave_balance (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(id),
    absence_type_id INT NOT NULL REFERENCES absence_types(id),
    entry_type LEAVE_TYPE NOT NULL,
    days NUMERIC(6,2) NOT NULL,
    accrual_year INT NOT NULL,
    expires_at DATE,
    reference_id BIGINT,
    created_at TIMESTAMP DEFAULT now()
);

-- the actual absence requests
CREATE TABLE absence_requests (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(id),
    absence_type_id INT NOT NULL REFERENCES absence_types(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    total_days NUMERIC(5,2) NOT NULL,
    reason TEXT,
    status absence_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    created_by BIGINT NOT NULL REFERENCES employees(id),
    approved_at TIMESTAMP,
    approved_by BIGINT REFERENCES employees(id),
    CONSTRAINT chk_dates CHECK (start_date <= end_date)
);

CREATE INDEX idx_absence_requests_employee ON absence_requests(employee_id);
CREATE INDEX idx_absence_requests_status ON absence_requests(status);
