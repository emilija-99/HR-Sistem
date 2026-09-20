package attendance

import (
	"database/sql"
	"fmt"
	"log"
	"main/middleware"
	types "main/types/attendance"
	empTypes "main/types/employee"
	"main/utils"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type Handler struct {
	db            *sql.DB
	store         types.AttendanceStore
	employeeStore empTypes.EmployeeStore
}

func NewHandler(db *sql.DB, store types.AttendanceStore, empStore empTypes.EmployeeStore) *Handler {
	return &Handler{db: db, store: store, employeeStore: empStore}
}

func (h *Handler) RegisterProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/attendance/clock-in", h.handleClockIn).Methods("POST")
	router.HandleFunc("/attendance/clock-out", h.handleClockOut).Methods("POST")
	router.HandleFunc("/attendance/status", h.handleGetStatus).Methods("GET")
	router.HandleFunc("/attendance/me", h.handleGetMine).Methods("GET")

	// admin overview — same as other reports (HR_ADMIN, PLATFORM_ADMIN)
	router.Handle("/attendance", middleware.RequirePermission(h.db, "reports.view", http.HandlerFunc(h.handleGetAll))).Methods("GET")
}

func (h *Handler) extractUserID(r *http.Request) (uint, error) {
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

func (h *Handler) resolveEmployeeID(userID uint) (uint, error) {
	emp, err := h.employeeStore.GetByUserID(userID)
	if err != nil {
		return 0, fmt.Errorf("employee profile not found for user %d", userID)
	}
	return emp.ID, nil
}

// POST /api/v1/attendance/clock-in
func (h *Handler) handleClockIn(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	record, err := h.store.ClockIn(employeeID)
	if err != nil {
		utils.WriteError(w, http.StatusConflict, "Clock-in failed", err.Error())
		return
	}

	log.Printf("Clock-in: employee=%d record=%d", employeeID, record.ID)
	utils.WriteSuccess(w, http.StatusCreated, "Created", record)
}

// POST /api/v1/attendance/clock-out
func (h *Handler) handleClockOut(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	record, err := h.store.ClockOut(employeeID)
	if err != nil {
		utils.WriteError(w, http.StatusConflict, "Clock-out failed", err.Error())
		return
	}

	log.Printf("Clock-out: employee=%d record=%d", employeeID, record.ID)
	utils.WriteSuccess(w, http.StatusOK, "OK", record)
}

// GET /api/v1/attendance/status
func (h *Handler) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	record, err := h.store.GetCurrentStatus(employeeID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch status", err.Error())
		return
	}

	if record == nil {
		utils.WriteSuccess(w, http.StatusOK, "OK", map[string]any{"is_working": false})
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", map[string]any{
		"is_working": true,
		"record":     record,
	})
}

// GET /api/v1/attendance/me
func (h *Handler) handleGetMine(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	records, err := h.store.GetByEmployee(employeeID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch attendance", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", records)
}

// GET /api/v1/attendance (admin — reports.view)
func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	records, err := h.store.GetAll()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch attendance", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", records)
}
