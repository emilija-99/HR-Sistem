package employee

import (
	"database/sql"
	"fmt"
	"log"
	"main/middleware"
	"main/services/auth"
	auditTypes "main/types/audit"
	types "main/types/employee"
	"main/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type Handler struct {
	db         *sql.DB
	store      types.EmployeeStore
	auditStore auditTypes.AuditStore
	validator  *utils.Validator
}

func NewHandler(db *sql.DB, store types.EmployeeStore, auditStore auditTypes.AuditStore, v *utils.Validator) *Handler {
	return &Handler{db: db, store: store, auditStore: auditStore, validator: v}
}

func (h *Handler) logAudit(action string, entityID uint, actorID *uint, details map[string]any, r *http.Request) {
	if h.auditStore == nil {
		return
	}
	if err := h.auditStore.Log(auditTypes.AuditEntry{
		Action:    action,
		Entity:    "employee",
		EntityID:  entityID,
		ActorID:   actorID,
		Details:   details,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}); err != nil {
		log.Printf("WARNING: audit log failed: %v", err)
	}
}

func (h *Handler) RegisterProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/employees", h.handleCreate).Methods("POST")
	router.HandleFunc("/employees/me", h.handleGetMe).Methods("GET")
	router.HandleFunc("/employees/me", h.handleUpdateMe).Methods("PATCH")

	// admin routes guarded by permission
	router.Handle("/employees", middleware.RequirePermission(h.db, "employees.read", http.HandlerFunc(h.handleGetAll))).Methods("GET")
	router.Handle("/employees/{id}", middleware.RequirePermission(h.db, "employees.read", http.HandlerFunc(h.handleGetByID))).Methods("GET")
	router.Handle("/employees/{id}", middleware.RequirePermission(h.db, "employees.write", http.HandlerFunc(h.handleUpdate))).Methods("PATCH")
	router.Handle("/admin/employees", middleware.RequirePermission(h.db, "employees.write", http.HandlerFunc(h.handleCreateByAdmin))).Methods("POST")

	// reference data (any authenticated user)
	router.HandleFunc("/countries", h.handleGetCountries).Methods("GET")
	router.HandleFunc("/positions", h.handleGetPositions).Methods("GET")
	router.HandleFunc("/departments", h.handleGetDepartments).Methods("GET")
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
	h.logAudit("employee.create", emp.ID, &userID,
		map[string]any{"first_name": emp.FirstName, "last_name": emp.LastName}, r)
	utils.WriteSuccess(w, http.StatusCreated, "Created", emp)
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
	utils.WriteSuccess(w, http.StatusOK, "OK", emp)
}

// PATCH /api/v1/employees/me
func (h *Handler) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
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

	var payload types.UpdateEmployeePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// employees cannot reassign their own supervisor
	payload.SupervisorID = nil

	updated, err := h.store.Update(int64(emp.ID), payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update employee", err.Error())
		return
	}
	h.logAudit("employee.update", updated.ID, &userID, map[string]any{"fields": payload}, r)
	utils.WriteSuccess(w, http.StatusOK, "OK", updated)
}

// GET /api/v1/employees
func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	emps, err := h.store.GetAll()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch employees", err.Error())
		return
	}
	utils.WriteSuccess(w, http.StatusOK, "OK", emps)
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
	utils.WriteSuccess(w, http.StatusOK, "OK", emp)
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
	utils.WriteSuccess(w, http.StatusOK, "OK", emp)
}

// POST /api/v1/admin/employees — HR/platform admin creates a user + employee profile
func (h *Handler) handleCreateByAdmin(w http.ResponseWriter, r *http.Request) {
	actorID, err := extractUserID(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	var payload types.CreateEmployeeByHRPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}
	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(payload.Email))
	hash, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	emp, err := h.store.CreateWithUser(email, hash, "EMPLOYEE", &actorID, types.Employee{
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
		SupervisorID: payload.SupervisorID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			utils.WriteError(w, http.StatusConflict, "Email already exists", "")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Failed to create employee", err.Error())
		return
	}

	log.Printf("Employee %d created by user %d (email %s)", emp.ID, actorID, email)
	h.logAudit("employee.create", emp.ID, &actorID,
		map[string]any{
			"email":       email,
			"first_name":  emp.FirstName,
			"last_name":   emp.LastName,
			"position_id": emp.PositionID,
		}, r)
	utils.WriteSuccess(w, http.StatusCreated, "Created", emp)
}

// ── reference data ────────────────────────────────────────────

func (h *Handler) handleGetCountries(w http.ResponseWriter, r *http.Request) {
	countries, err := h.store.GetCountries()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch countries", err.Error())
		return
	}
	utils.WriteSuccess(w, http.StatusOK, "OK", countries)
}

func (h *Handler) handleGetPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := h.store.GetPositions()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch positions", err.Error())
		return
	}
	utils.WriteSuccess(w, http.StatusOK, "OK", positions)
}

func (h *Handler) handleGetDepartments(w http.ResponseWriter, r *http.Request) {
	departments, err := h.store.GetDepartments()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch departments", err.Error())
		return
	}
	utils.WriteSuccess(w, http.StatusOK, "OK", departments)
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
