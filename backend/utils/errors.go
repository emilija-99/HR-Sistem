package utils

type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}

var (
	ErrBadRequest         = "bad request"
	ErrInvalidJSON        = "invalid JSON payload"
	ErrInternalServer     = "internal server error"
	ErrUserNotFound       = "user not found"
	ErrEmailAlreadyExists = "email already exists"
	ErrUnauthorized       = "unauthorized"
)
