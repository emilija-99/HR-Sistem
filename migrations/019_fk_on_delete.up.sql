-- Deleting an employee (directly, or through the user it belongs to) must not
-- fail because of dependent rows.
--
--   ON DELETE CASCADE  -> the row belongs to the employee, remove it with them
--   ON DELETE SET NULL -> the row is kept, only the reference is cleared
--
-- employee-owned data
ALTER TABLE leave_balance
    DROP CONSTRAINT IF EXISTS leave_balance_employee_id_fkey,
    ADD CONSTRAINT leave_balance_employee_id_fkey
        FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

ALTER TABLE employee_leave_policy
    DROP CONSTRAINT IF EXISTS employee_leave_policy_employee_id_fkey,
    ADD CONSTRAINT employee_leave_policy_employee_id_fkey
        FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

ALTER TABLE attendance
    DROP CONSTRAINT IF EXISTS attendance_employee_id_fkey,
    ADD CONSTRAINT attendance_employee_id_fkey
        FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

ALTER TABLE absence_requests
    DROP CONSTRAINT IF EXISTS absence_requests_employee_id_fkey,
    ADD CONSTRAINT absence_requests_employee_id_fkey
        FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE;

-- references to the employee, but not owned by them
ALTER TABLE absence_requests
    DROP CONSTRAINT IF EXISTS absence_requests_approved_by_fkey,
    ADD CONSTRAINT absence_requests_approved_by_fkey
        FOREIGN KEY (approved_by) REFERENCES employees(id) ON DELETE SET NULL;

ALTER TABLE employees
    DROP CONSTRAINT IF EXISTS employees_supervisor_id_fkey,
    ADD CONSTRAINT employees_supervisor_id_fkey
        FOREIGN KEY (supervisor_id) REFERENCES employees(id) ON DELETE SET NULL;

-- a deleted account must not block deleting the account it created
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_created_by_fkey,
    ADD CONSTRAINT users_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_updated_by_fkey,
    ADD CONSTRAINT users_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL;
