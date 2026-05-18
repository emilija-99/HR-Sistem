package employee

import (
	types "main/types/employee"
	"main/utils"
)

type Handler struct {
	store     types.EmployeeStore
	validator *utils.Validator
}

func NewHandler(store types.Employee, v *utils.Validator) *Handler {
	return &Handler{store: store, validator: v}
}
