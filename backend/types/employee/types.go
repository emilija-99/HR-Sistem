package types

import "time"

type EmployeeStore interface {
	Create(employee Employee) (*Employee, error)
	CreateWithUser(email, passwordHash, roleName string, createdBy *uint, employee Employee) (*Employee, error)
	GetByID(id int64) (*Employee, error)
	GetByUserID(userID uint) (*Employee, error)
	GetAll() ([]Employee, error)
	Update(id int64, p UpdateEmployeePayload) (*Employee, error)

	// reference data for dropdowns
	GetCountries() ([]Country, error)
	GetPositions() ([]Position, error)
	GetDepartments() ([]Department, error)
}

type Employee struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	PhoneNumber  *string   `json:"phone_number,omitempty"`
	PrivateEmail *string   `json:"private_email,omitempty"`
	Street       *string   `json:"street,omitempty"`
	Country      uint      `json:"country"`
	City         *string   `json:"city,omitempty"`
	DateOfBirth  *string   `json:"date_of_birth,omitempty"`
	HireDate     *string   `json:"hire_date,omitempty"`
	PositionID   *uint     `json:"position_id,omitempty"`
	SupervisorID *uint     `json:"supervisor_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	// Joined fields (read-only from positions + departments)
	PositionTitle  *string `json:"position_title,omitempty"`
	PositionLevel  *string `json:"position_level,omitempty"`
	DepartmentID   *uint   `json:"department_id,omitempty"`
	DepartmentName *string `json:"department_name,omitempty"`
	SupervisorName *string `json:"supervisor_name,omitempty"`
}

type CreateEmployeePayload struct {
	FirstName    string  `json:"first_name"    validate:"required,min=1,max=50"`
	LastName     string  `json:"last_name"     validate:"required,min=1,max=50"`
	PhoneNumber  *string `json:"phone_number"`
	PrivateEmail *string `json:"private_email"`
	Street       *string `json:"street"`
	Country      uint    `json:"country"       validate:"required"`
	City         *string `json:"city"`
	DateOfBirth  *string `json:"date_of_birth"`
	HireDate     *string `json:"hire_date"`
	PositionID   *uint   `json:"position_id"`
}

// CreateEmployeeByHRPayload creates both the user account and the employee
// profile in one step (HR / platform admin only).
type CreateEmployeeByHRPayload struct {
	Email        string  `json:"email"         validate:"required,email"`
	Password     string  `json:"password"      validate:"required,min=8,max=25,strongpwd"`
	FirstName    string  `json:"first_name"    validate:"required,min=1,max=50"`
	LastName     string  `json:"last_name"     validate:"required,min=1,max=50"`
	PhoneNumber  *string `json:"phone_number"`
	PrivateEmail *string `json:"private_email"`
	Street       *string `json:"street"`
	Country      uint    `json:"country"       validate:"required"`
	City         *string `json:"city"`
	DateOfBirth  *string `json:"date_of_birth"`
	HireDate     *string `json:"hire_date"`
	PositionID   *uint   `json:"position_id"`
	SupervisorID *uint   `json:"supervisor_id"`
}

type UpdateEmployeePayload struct {
	FirstName    *string `json:"first_name"`
	LastName     *string `json:"last_name"`
	PhoneNumber  *string `json:"phone_number"`
	PrivateEmail *string `json:"private_email"`
	Street       *string `json:"street"`
	Country      *uint   `json:"country"`
	City         *string `json:"city"`
	DateOfBirth  *string `json:"date_of_birth"`
	HireDate     *string `json:"hire_date"`
	PositionID   *uint   `json:"position_id"`
	SupervisorID *uint   `json:"supervisor_id"`
}

// ── Reference data ────────────────────────────────────────────

type Country struct {
	CountryID   uint   `json:"country_id"`
	CountryName string `json:"country_name"`
	Iso         string `json:"iso"`
}

type Position struct {
	ID           uint    `json:"id"`
	DepartmentID uint    `json:"department_id"`
	Title        string  `json:"title"`
	Level        string  `json:"level"`
	Description  *string `json:"description,omitempty"`

	// joined
	DepartmentName string `json:"department_name,omitempty"`
}

type Department struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
}

// JoinsQuery is shared by all SELECT queries that need position + department data
const JoinsQuery = `SELECT e.id, e.user_id, e.first_name, e.last_name,
	e.phone_number, e.private_email, e.street, e.country, e.city,
	e.date_of_birth, e.hire_date, e.position_id, e.created_at,
	p.title, p.level, d.id, d.name,
	e.supervisor_id, sup.first_name, sup.last_name
	FROM employees e
	LEFT JOIN positions p ON p.id = e.position_id
	LEFT JOIN departments d ON d.id = p.department_id
	LEFT JOIN employees sup ON sup.id = e.supervisor_id`
