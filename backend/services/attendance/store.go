package attendance

import (
	"database/sql"
	"fmt"
	types "main/types/attendance"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

const attendanceSelect = `SELECT a.id, a.employee_id, a.clock_in, a.clock_out, a.status, a.created_at,
	e.first_name, e.last_name
	FROM attendance a
	LEFT JOIN employees e ON e.id = a.employee_id`

func scanAttendance(scanner interface {
	Scan(dest ...any) error
}) (*types.Attendance, error) {
	var a types.Attendance
	err := scanner.Scan(
		&a.ID, &a.EmployeeID, &a.ClockIn, &a.ClockOut, &a.Status, &a.CreatedAt,
		&a.FirstName, &a.LastName,
	)
	if err != nil {
		return nil, err
	}

	// compute worked minutes when clock_out exists
	if a.ClockOut != nil {
		minutes := int(a.ClockOut.Sub(a.ClockIn).Minutes())
		a.WorkedMinutes = &minutes
	}

	return &a, nil
}

func (s *Store) ClockIn(employeeID uint) (*types.Attendance, error) {
	// prevent double clock-in: only one open (WORKING) record per employee
	var existing types.Attendance
	err := s.db.QueryRow(
		attendanceSelect+` WHERE a.employee_id = $1 AND a.status = 'WORKING'`, employeeID,
	).Scan(
		&existing.ID, &existing.EmployeeID, &existing.ClockIn, &existing.ClockOut,
		&existing.Status, &existing.CreatedAt, &existing.FirstName, &existing.LastName,
	)
	if err == nil {
		return nil, fmt.Errorf("already clocked in (record %d)", existing.ID)
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	var id uint
	err = s.db.QueryRow(
		`INSERT INTO attendance (employee_id) VALUES ($1) RETURNING id`, employeeID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	a, err := scanAttendance(s.db.QueryRow(
		attendanceSelect+` WHERE a.id = $1`, id,
	))
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) ClockOut(employeeID uint) (*types.Attendance, error) {
	var id uint
	err := s.db.QueryRow(
		`UPDATE attendance SET clock_out = NOW(), status = 'DONE'
		 WHERE employee_id = $1 AND status = 'WORKING'
		 RETURNING id`, employeeID,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no active clock-in found")
	}
	if err != nil {
		return nil, err
	}

	a, err := scanAttendance(s.db.QueryRow(
		attendanceSelect+` WHERE a.id = $1`, id,
	))
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) GetCurrentStatus(employeeID uint) (*types.Attendance, error) {
	a, err := scanAttendance(s.db.QueryRow(
		attendanceSelect+` WHERE a.employee_id = $1 AND a.status = 'WORKING'
		 ORDER BY a.clock_in DESC LIMIT 1`, employeeID,
	))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) GetByEmployee(employeeID uint) ([]types.Attendance, error) {
	rows, err := s.db.Query(
		attendanceSelect+` WHERE a.employee_id = $1 ORDER BY a.clock_in DESC`, employeeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []types.Attendance
	for rows.Next() {
		a, err := scanAttendance(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *a)
	}
	if list == nil {
		list = []types.Attendance{}
	}
	return list, rows.Err()
}

func (s *Store) GetAll() ([]types.Attendance, error) {
	rows, err := s.db.Query(`
		SELECT a.id, a.employee_id, a.clock_in, a.clock_out, a.status, a.created_at,
		       e.first_name, e.last_name,
		       p.title, p.level, d.id, d.name
		FROM attendance a
		LEFT JOIN employees e ON e.id = a.employee_id
		LEFT JOIN positions p ON p.id = e.position_id
		LEFT JOIN departments d ON d.id = p.department_id
		ORDER BY a.clock_in DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []types.Attendance
	for rows.Next() {
		var a types.Attendance
		if err := rows.Scan(
			&a.ID, &a.EmployeeID, &a.ClockIn, &a.ClockOut, &a.Status, &a.CreatedAt,
			&a.FirstName, &a.LastName,
			&a.PositionTitle, &a.PositionLevel, &a.DepartmentID, &a.DepartmentName,
		); err != nil {
			return nil, err
		}
		if a.ClockOut != nil {
			minutes := int(a.ClockOut.Sub(a.ClockIn).Minutes())
			a.WorkedMinutes = &minutes
		}
		list = append(list, a)
	}
	if list == nil {
		list = []types.Attendance{}
	}
	return list, rows.Err()
}
