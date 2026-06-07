package types

import (
	"time"
)

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(user User) (uint, error)
	CreateUserWithRole(user User, roleName string, createdBy *uint) (*User, error)
	EmailExists(email string) bool
	AssignRole(userID uint, roleName string) error
	GetUserRole(userID uint) (string, error)
	SaveRefreshToken(userID uint, tokenHash string) error
	GetUserIDByRefreshToken(tokenHash string) (uint, error)
	GetUserPremissions(roleName PermissionRequest) (*UserPermissions, error)
	ChangeUserStatus(userID uint, isActive bool) (*User, error)
	GetAllUsers() ([]User, error)
	GetUserByIDWithRole(id int64) (*User, string, error)
}

// Users - domain model
type User struct {
	ID        uint       `json:"id"`
	Email     string     `json:"email"`
	Password  string     `json:"-"` // password_hash in DB
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	CreatedBy *uint      `json:"created_by"`
	UpdatedAt *time.Time `json:"updated_at"`
	UpdatedBy *uint      `json:"updated_by"`
}

type RegisterUserPayload struct {
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password_hash"  validate:"required,min=8,max=25,strongpwd"`
}

type LoginUserPayload struct {
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password_hash"  validate:"required,min=8,max=25,strongpwd"`
}

type CreateUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password_hash"`
}

type UserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

type APIResponse struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type Permission struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
}

type UserPermissions struct {
	Permissions []Permission `json:"permissions"`
}

type PermissionRequest struct {
	RoleName string `json:"roleName"`
}
