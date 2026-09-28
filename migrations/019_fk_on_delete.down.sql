-- Revert to the default NO ACTION constraints (see the .up migration).
ALTER TABLE leave_balance
    DROP CONSTRAINT IF EXISTS leave_balance_employee_id_fkey,
    ADD CONSTRAINT leave_balance_employee_id_fkey
        FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE employee_leave_policy
    DROP CONSTRAINT IF EXISTS employee_leave_policy_employee_id_fkey,
    ADD CONSTRAINT employee_leave_policy_employee_id_fkey
        FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE attendance
    DROP CONSTRAINT IF EXISTS attendance_employee_id_fkey,
    ADD CONSTRAINT attendance_employee_id_fkey
        FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE absence_requests
    DROP CONSTRAINT IF EXISTS absence_requests_employee_id_fkey,
    ADD CONSTRAINT absence_requests_employee_id_fkey
        FOREIGN KEY (employee_id) REFERENCES employees(id);

ALTER TABLE absence_requests
    DROP CONSTRAINT IF EXISTS absence_requests_approved_by_fkey,
    ADD CONSTRAINT absence_requests_approved_by_fkey
        FOREIGN KEY (approved_by) REFERENCES employees(id);

ALTER TABLE employees
    DROP CONSTRAINT IF EXISTS employees_supervisor_id_fkey,
    ADD CONSTRAINT employees_supervisor_id_fkey
        FOREIGN KEY (supervisor_id) REFERENCES employees(id);

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_created_by_fkey,
    ADD CONSTRAINT users_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users(id);

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_updated_by_fkey,
    ADD CONSTRAINT users_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users(id);
