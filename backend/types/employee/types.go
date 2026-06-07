package types

import "time"

type EmployeeStore interface {
	Create(employee Employee) (*Employee, error)
	GetByID(id int64) (*Employee, error)
	GetByUserID(userID uint) (*Employee, error)
	GetAll() ([]Employee, error)
	Update(id int64, p UpdateEmployeePayload) (*Employee, error)
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
	CreatedAt    time.Time `json:"created_at"`
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
}
