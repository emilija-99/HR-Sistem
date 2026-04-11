package types

import (
	"time"
)

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(user User) (uint, error)
	CreateUserWithRole(user User, roleName string) (*User, error)
	EmailExists(email string) string
	AssignRole(userID uint, roleName string) error
	GetUserRole(userID uint) (string, error)
	SaveRefreshToken(userID uint, tokenHash string) error
	GetUserIDByRefreshToken(tokenHash string) (uint, error)
}
type RegisterUserPayload struct {
	FirstName string `json:"firstName" validate:"required,min=3,max=20"`
	LastName  string `json:"lastName"  validate:"required,min=3,max=20"`
	Email     string `json:"email"     validate:"required,email"`
	Password  string `json:"password"  validate:"required,min=8,max=25,strongpwd"`
}

type LoginUserPayload struct {
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=8,max=25,strongpwd"`
}

type User struct {
	ID        uint   `gorm:"primaryKey"`
	FirstName string `gorm:"size:20;not null"`
	LastName  string `gorm:"size:20;not null"`
	Email     string `gorm:"size:255;not null;uniqueIndex"`
	Password  string `gorm:"size:255;not null"` // hashed
	CreatedAt time.Time
}

type CreateUserPayload struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type APIResponse struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}
