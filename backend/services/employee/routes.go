package employee

import (
	"fmt"
	"log"
	"main/middleware"
	types "main/types/employee"
	"main/utils"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.EmployeeStore
	validator *utils.Validator
}

func NewHandler(store types.EmployeeStore, v *utils.Validator) *Handler {
	return &Handler{store: store, validator: v}
}

func (h *Handler) RegisterProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/employees", h.handleCreate).Methods("POST")
	router.HandleFunc("/employees", h.handleGetAll).Methods("GET")
	router.HandleFunc("/employees/me", h.handleGetMe).Methods("GET")
	router.HandleFunc("/employees/{id}", h.handleGetByID).Methods("GET")
	router.HandleFunc("/employees/{id}", h.handleUpdate).Methods("PATCH")
}

// POST /api/v1/employees
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	var payload types.CreateEmployeePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	emp, err := h.store.Create(types.Employee{
		UserID:       userID,
		FirstName:    payload.FirstName,
		LastName:     payload.LastName,
		PhoneNumber:  payload.PhoneNumber,
		PrivateEmail: payload.PrivateEmail,
		Street:       payload.Street,
		Country:      payload.Country,
		City:         payload.City,
		DateOfBirth:  payload.DateOfBirth,
		HireDate:     payload.HireDate,
		PositionID:   payload.PositionID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to create employee", err.Error())
		return
	}

	log.Printf("Employee created: id=%d for user=%d", emp.ID, userID)
	utils.WriteJSON(w, http.StatusCreated, emp)
}

// GET /api/v1/employees/me
func (h *Handler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	emp, err := h.store.GetByUserID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Employee profile not found", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, emp)
}

// GET /api/v1/employees
func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	emps, err := h.store.GetAll()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch employees", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, emps)
}

// GET /api/v1/employees/{id}
func (h *Handler) handleGetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	emp, err := h.store.GetByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Employee not found", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, emp)
}

// PATCH /api/v1/employees/{id}
func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var payload types.UpdateEmployeePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	emp, err := h.store.Update(id, payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update employee", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, emp)
}

// extractUserID pulls the user_id from JWT claims stored in context
func extractUserID(r *http.Request) (uint, error) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id not found in token")
	}
	return uint(userIDFloat), nil
}
