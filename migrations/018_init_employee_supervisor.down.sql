DROP INDEX IF EXISTS idx_employees_supervisor;

ALTER TABLE employees
    DROP COLUMN IF EXISTS supervisor_id;
