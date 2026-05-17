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
	GetUserPremissions(roleName PermissionRequest) (*UserPermissions, error)
	ChangeUserStatus(userID uint, isActive bool) (*User, error)
}
type RegisterUserPayload struct {
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=8,max=25,strongpwd"`
}

type LoginUserPayload struct {
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=8,max=25,strongpwd"`
}

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"size:255;not null;uniqueIndex"`
	Password  string `gorm:"size:255;not null"` // hashed
	CreatedAt time.Time
	IsActive  bool `gorm:"default:true"`
}

type CreateUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	ID   int
	Name string
}

type UserPermissions struct {
	Permissions []Permission
}

type PermissionRequest struct {
	RoleName string `json:"roleName"`
}
