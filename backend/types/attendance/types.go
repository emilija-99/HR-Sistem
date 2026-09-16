package types

import "time"

type AttendanceStore interface {
	ClockIn(employeeID uint) (*Attendance, error)
	ClockOut(employeeID uint) (*Attendance, error)
	GetCurrentStatus(employeeID uint) (*Attendance, error)
	GetByEmployee(employeeID uint) ([]Attendance, error)
	GetAll() ([]Attendance, error)
}

type Attendance struct {
	ID         uint       `json:"id"`
	EmployeeID uint       `json:"employee_id"`
	ClockIn    time.Time  `json:"clock_in"`
	ClockOut   *time.Time `json:"clock_out,omitempty"`
	Status     string     `json:"status"` // WORKING / DONE
	CreatedAt  time.Time  `json:"created_at"`

	// joined fields (read-only)
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	PositionTitle  string `json:"position_title,omitempty"`
	PositionLevel  string `json:"position_level,omitempty"`
	DepartmentID   *uint  `json:"department_id,omitempty"`
	DepartmentName string `json:"department_name,omitempty"`

	// computed
	WorkedMinutes *int `json:"worked_minutes,omitempty"`
}
