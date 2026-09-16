CREATE TABLE IF NOT EXISTS attendance (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(id),
    clock_in TIMESTAMP NOT NULL DEFAULT now(),
    clock_out TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'WORKING',
    created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_attendance_employee ON attendance(employee_id);
CREATE INDEX idx_attendance_status ON attendance(status);
