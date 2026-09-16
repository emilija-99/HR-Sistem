ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS supervisor_id BIGINT REFERENCES employees(id);

CREATE INDEX IF NOT EXISTS idx_employees_supervisor ON employees(supervisor_id);
