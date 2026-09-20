package absence

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"main/middleware"
	types "main/types/absence"
	auditTypes "main/types/audit"
	empTypes "main/types/employee"
	"main/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type Handler struct {
	db            *sql.DB
	store         types.AbsenceStore
	employeeStore empTypes.EmployeeStore
	auditStore    auditTypes.AuditStore
	validator     *utils.Validator
}

func NewHandler(db *sql.DB, store types.AbsenceStore, empStore empTypes.EmployeeStore, auditStore auditTypes.AuditStore, v *utils.Validator) *Handler {
	return &Handler{db: db, store: store, employeeStore: empStore, auditStore: auditStore, validator: v}
}

func (h *Handler) RegisterPublicRoutes(router *mux.Router) {
	router.HandleFunc("/absences/types", h.handleGetAbsenceTypes).Methods("GET")
}

func (h *Handler) RegisterProtectedRoutes(router *mux.Router) {
	// self-service (any authenticated user)
	router.HandleFunc("/absences/types", h.handleGetAbsenceTypes).Methods("GET")
	router.HandleFunc("/absences/requests", h.handleCreateRequest).Methods("POST")
	router.HandleFunc("/absences/requests/me", h.handleGetMyRequests).Methods("GET")
	router.HandleFunc("/absences/requests/{id}", h.handleUpdateRequest).Methods("PATCH")
	router.HandleFunc("/absences/requests/{id}/submit", h.handleSubmitRequest).Methods("POST")
	router.HandleFunc("/absences/requests/{id}/cancel", h.handleCancelRequest).Methods("PUT")
	router.HandleFunc("/absences/balance/me", h.handleGetMyBalance).Methods("GET")
	router.HandleFunc("/absences/balance/my-policies", h.handleGetMyPolicies).Methods("GET")

	// review / approval (HR_ADMIN, MANAGER_PORTAL_ACCESS, PLATFORM_ADMIN)
	router.Handle("/absences/requests", middleware.RequirePermission(h.db, "absence.read.all", http.HandlerFunc(h.handleGetAllRequests))).Methods("GET")
	router.Handle("/absences/requests/{id}/approve", middleware.RequirePermission(h.db, "absence.approve", http.HandlerFunc(h.handleApproveRequest))).Methods("PUT")
	router.Handle("/absences/requests/{id}/reject", middleware.RequirePermission(h.db, "absence.reject", http.HandlerFunc(h.handleRejectRequest))).Methods("PUT")

	// leave administration — balances and policies (HR_ADMIN, PLATFORM_ADMIN)
	router.Handle("/absences/balance/{id}", middleware.RequirePermission(h.db, "leave.manage", http.HandlerFunc(h.handleGetBalanceByID))).Methods("GET")
	router.Handle("/absences/balance/grant", middleware.RequirePermission(h.db, "leave.manage", http.HandlerFunc(h.handleGrantBalance))).Methods("POST")
	router.Handle("/absences/balance/adjust", middleware.RequirePermission(h.db, "leave.manage", http.HandlerFunc(h.handleAdjustBalance))).Methods("PUT")
	router.Handle("/absences/balance/rollover", middleware.RequirePermission(h.db, "leave.manage", http.HandlerFunc(h.handleRollover))).Methods("POST")
	router.Handle("/absences/policies", middleware.RequirePermission(h.db, "leave.manage", http.HandlerFunc(h.handleGetPolicies))).Methods("GET")
	router.Handle("/absences/policies/employee/{id}", middleware.RequirePermission(h.db, "leave.manage", http.HandlerFunc(h.handleGetEmployeePolicies))).Methods("GET")
	router.Handle("/absences/policies/assign", middleware.RequirePermission(h.db, "leave.manage", http.HandlerFunc(h.handleAssignPolicy))).Methods("POST")
	router.Handle("/absences/maintenance/run", middleware.RequirePermission(h.db, "leave.manage", http.HandlerFunc(h.handleRunMaintenance))).Methods("POST")
}

// ── helpers ───────────────────────────────────────────────────

func extractClaims(r *http.Request) (userID uint, role string, err error) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		return 0, "", fmt.Errorf("invalid token claims")
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, "", fmt.Errorf("user_id not found in token")
	}
	role, _ = claims["role"].(string)
	return uint(userIDFloat), role, nil
}

// businessDays counts working days (Mon–Fri, excluding holidays) in the
// inclusive [startDate, endDate] range.
func businessDays(startDate, endDate string, holidays map[string]bool) (float64, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0, fmt.Errorf("invalid start_date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return 0, fmt.Errorf("invalid end_date: %w", err)
	}
	if end.Before(start) {
		return 0, fmt.Errorf("end_date cannot be before start_date")
	}

	days := 0.0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		switch d.Weekday() {
		case time.Saturday, time.Sunday:
			continue
		}
		if holidays[d.Format("2006-01-02")] {
			continue
		}
		days++
	}
	return days, nil
}

// isWeekend reports whether the given date is a Saturday or Sunday.
func isWeekend(d time.Time) bool {
	return d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
}

// validateRequestDates enforces the self-service rules for an absence period:
//   - the end may not be before the start,
//   - the period may not start in the past,
//   - neither boundary may fall on a Saturday or Sunday.
//
// `today` is a parameter so the rule stays deterministic in tests.
func validateRequestDates(startDate, endDate string, today time.Time) error {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("invalid start_date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("invalid end_date: %w", err)
	}
	if end.Before(start) {
		return fmt.Errorf("end_date cannot be before start_date")
	}

	day := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if start.Before(day) {
		return fmt.Errorf("start_date cannot be in the past")
	}
	if isWeekend(start) {
		return fmt.Errorf("start_date cannot be a Saturday or Sunday")
	}
	if isWeekend(end) {
		return fmt.Errorf("end_date cannot be a Saturday or Sunday")
	}
	return nil
}

// computeBusinessDays validates the dates and returns the billable working
// days, rejecting periods that contain no working days at all.
func (h *Handler) computeBusinessDays(employeeID uint, start, end string) (float64, error) {
	if err := validateRequestDates(start, end, time.Now()); err != nil {
		return 0, err
	}
	holidays, err := h.store.GetHolidays(employeeID, start, end)
	if err != nil {
		return 0, fmt.Errorf("failed to load holidays: %w", err)
	}
	days, err := businessDays(start, end, holidays)
	if err != nil {
		return 0, err
	}
	if days == 0 {
		return 0, fmt.Errorf("selected period contains no working days")
	}
	return days, nil
}

// ErrInsufficientBalance marks a failed balance check (as opposed to an
// infrastructure error while resolving the policy or the ledger).
var ErrInsufficientBalance = errors.New("insufficient balance")

// balanceRequired reports whether the requested days must be covered by the
// leave ledger.
//
// Fail-closed: a nil policy means "no rules are configured for this absence
// type", so the days are still checked against the ledger. Only an explicit
// `requires_balance = false` policy (e.g. Sick unlimited) turns the check off.
func balanceRequired(policy *types.LeavePolicy) bool {
	return policy == nil || policy.RequiresBalance
}

// enforceBalance is the single balance gate for create, edit, submit and
// approve: a request must never be saved as PENDING or approved without enough
// available days.
func (h *Handler) enforceBalance(employeeID, absenceTypeID uint, atDate string, days float64) error {
	policy, err := h.store.GetActivePolicy(employeeID, absenceTypeID, atDate)
	if err != nil {
		return fmt.Errorf("failed to resolve leave policy: %w", err)
	}
	if !balanceRequired(policy) {
		return nil
	}

	avail, err := h.store.GetAvailableForType(employeeID, absenceTypeID)
	if err != nil {
		return fmt.Errorf("failed to load balance: %w", err)
	}
	if avail < days {
		return fmt.Errorf("%w: available %.1f days, requested %.1f days",
			ErrInsufficientBalance, avail, days)
	}
	return nil
}

// requireBalance runs enforceBalance and writes the matching error response.
// It reports whether the caller may continue.
func (h *Handler) requireBalance(w http.ResponseWriter, employeeID, absenceTypeID uint, atDate string, days float64) bool {
	err := h.enforceBalance(employeeID, absenceTypeID, atDate, days)
	if err == nil {
		return true
	}
	if errors.Is(err, ErrInsufficientBalance) {
		utils.WriteError(w, http.StatusConflict, "Insufficient balance", err.Error())
	} else {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to check balance", err.Error())
	}
	return false
}

func (h *Handler) resolveEmployeeID(userID uint) (uint, error) {
	emp, err := h.employeeStore.GetByUserID(userID)
	if err != nil {
		return 0, fmt.Errorf("employee profile not found for user %d", userID)
	}
	return emp.ID, nil
}

// logAudit writes an entry to MongoDB, ignoring failures (audit must never break the request)
func (h *Handler) logAudit(action, entity string, entityID uint, actorID *uint, details map[string]any, r *http.Request) {
	if h.auditStore == nil {
		return
	}
	if err := h.auditStore.Log(auditTypes.AuditEntry{
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		ActorID:   actorID,
		Details:   details,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}); err != nil {
		log.Printf("WARNING: audit log failed for %s/%d: %v", entity, entityID, err)
	}
}

// ── absence types ─────────────────────────────────────────────

func (h *Handler) handleGetAbsenceTypes(w http.ResponseWriter, r *http.Request) {
	payload, err := h.store.GetAllAbsenceTypes()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get absence types", err.Error())
		return
	}
	utils.WriteSuccess(w, http.StatusOK, "OK", payload.Data)
}

// ── create request ────────────────────────────────────────────

func (h *Handler) handleCreateRequest(w http.ResponseWriter, r *http.Request) {
	userID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	var payload types.CreateAbsenceRequestPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	days, err := h.computeBusinessDays(employeeID, payload.StartDate, payload.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid dates", err.Error())
		return
	}

	status := "PENDING"
	if payload.Status != nil && *payload.Status == "DRAFT" {
		status = "DRAFT"
	}

	// Drafts may be saved regardless of balance; submitting requires enough days.
	if status == "PENDING" {
		if !h.requireBalance(w, employeeID, payload.AbsenceTypeID, payload.StartDate, days) {
			return
		}
	}

	req, err := h.store.CreateRequest(types.AbsenceRequest{
		EmployeeID:    employeeID,
		AbsenceTypeID: payload.AbsenceTypeID,
		StartDate:     payload.StartDate,
		EndDate:       payload.EndDate,
		TotalDays:     days,
		Reason:        payload.Reason,
		Status:        status,
		CreatedBy:     employeeID,
	})
	if err != nil {
		if errors.Is(err, ErrOverlap) {
			utils.WriteError(w, http.StatusConflict, "Overlapping request", err.Error())
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Failed to create request", err.Error())
		return
	}

	log.Printf("Absence request created: id=%d employee=%d type=%d days=%.1f status=%s",
		req.ID, employeeID, payload.AbsenceTypeID, days, status)

	h.logAudit("absence.request.create", "absence_request", req.ID, &employeeID,
		map[string]any{
			"employee_id":     employeeID,
			"absence_type_id": payload.AbsenceTypeID,
			"start_date":      payload.StartDate,
			"end_date":        payload.EndDate,
			"total_days":      days,
			"status":          status,
		}, r)

	utils.WriteSuccess(w, http.StatusCreated, "Created", req)
}

// ── my requests ───────────────────────────────────────────────

func (h *Handler) handleGetMyRequests(w http.ResponseWriter, r *http.Request) {
	userID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	requests, err := h.store.GetRequestsByEmployee(employeeID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch requests", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", requests)
}

// ── all requests (admin) ──────────────────────────────────────

func (h *Handler) handleGetAllRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.store.GetAllRequests()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch requests", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", requests)
}

// ── approve / reject (admin) ──────────────────────────────────

func (h *Handler) handleApproveRequest(w http.ResponseWriter, r *http.Request) {
	h.changeStatusByAdmin(w, r, "APPROVED")
}

func (h *Handler) handleRejectRequest(w http.ResponseWriter, r *http.Request) {
	h.changeStatusByAdmin(w, r, "REJECTED")
}

func (h *Handler) changeStatusByAdmin(w http.ResponseWriter, r *http.Request, newStatus string) {
	userID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request ID", err.Error())
		return
	}

	adminEmployeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Admin employee profile required", err.Error())
		return
	}

	// When approving, enforce requires_balance from the active leave policy
	if newStatus == "APPROVED" {
		existing, err := h.store.GetRequestByID(id)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch request", err.Error())
			return
		}
		if existing == nil {
			utils.WriteError(w, http.StatusNotFound, "Request not found", "")
			return
		}
		if existing.Status != "APPROVED" {
			if !h.requireBalance(w, existing.EmployeeID, existing.AbsenceTypeID, existing.StartDate, existing.TotalDays) {
				return
			}
		}
	}

	req, err := h.store.UpdateRequestStatus(id, newStatus, &adminEmployeeID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update request", err.Error())
		return
	}

	log.Printf("Absence request %d -> %s by employee %d", id, newStatus, adminEmployeeID)
	h.logAudit("absence.request."+newStatus, "absence_request", uint(id), &adminEmployeeID,
		map[string]any{"request_id": id, "status": newStatus}, r)
	utils.WriteSuccess(w, http.StatusOK, "OK", req)
}

// ── cancel (owner, pending only) ──────────────────────────────

func (h *Handler) handleCancelRequest(w http.ResponseWriter, r *http.Request) {
	userID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request ID", err.Error())
		return
	}

	req, err := h.store.GetRequestByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch request", err.Error())
		return
	}
	if req == nil {
		utils.WriteError(w, http.StatusNotFound, "Request not found", "")
		return
	}

	// only the owner can cancel, and only while PENDING
	if req.EmployeeID != employeeID {
		utils.WriteError(w, http.StatusForbidden, "Forbidden", "can only cancel own requests")
		return
	}
	if req.Status != "PENDING" && req.Status != "DRAFT" {
		utils.WriteError(w, http.StatusConflict, "Cannot cancel", "only PENDING or DRAFT requests can be cancelled")
		return
	}

	updated, err := h.store.UpdateRequestStatus(id, "CANCELLED", nil)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to cancel request", err.Error())
		return
	}

	log.Printf("Absence request %d cancelled by employee %d", id, employeeID)
	h.logAudit("absence.request.cancel", "absence_request", uint(id), &employeeID,
		map[string]any{"request_id": id}, r)
	utils.WriteSuccess(w, http.StatusOK, "OK", updated)
}

// ── edit / submit draft (owner) ───────────────────────────────

// PATCH /api/v1/absences/requests/{id} — edit own DRAFT or PENDING request
func (h *Handler) handleUpdateRequest(w http.ResponseWriter, r *http.Request) {
	userID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request ID", err.Error())
		return
	}

	req, err := h.store.GetRequestByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch request", err.Error())
		return
	}
	if req == nil {
		utils.WriteError(w, http.StatusNotFound, "Request not found", "")
		return
	}
	if req.EmployeeID != employeeID {
		utils.WriteError(w, http.StatusForbidden, "Forbidden", "can only edit own requests")
		return
	}
	if req.Status != "DRAFT" && req.Status != "PENDING" {
		utils.WriteError(w, http.StatusConflict, "Cannot edit", "only DRAFT or PENDING requests can be edited")
		return
	}

	var payload types.UpdateAbsenceRequestPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// merge the patch onto the stored request
	if payload.AbsenceTypeID != nil {
		req.AbsenceTypeID = *payload.AbsenceTypeID
	}
	if payload.StartDate != nil {
		req.StartDate = *payload.StartDate
	}
	if payload.EndDate != nil {
		req.EndDate = *payload.EndDate
	}
	if payload.Reason != nil {
		req.Reason = payload.Reason
	}

	days, err := h.computeBusinessDays(employeeID, req.StartDate, req.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid dates", err.Error())
		return
	}
	req.TotalDays = days

	if req.Status == "PENDING" {
		if !h.requireBalance(w, employeeID, req.AbsenceTypeID, req.StartDate, days) {
			return
		}
	}

	updated, err := h.store.UpdateRequest(*req)
	if err != nil {
		if errors.Is(err, ErrOverlap) {
			utils.WriteError(w, http.StatusConflict, "Overlapping request", err.Error())
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, "Request not found", "")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update request", err.Error())
		return
	}

	h.logAudit("absence.request.update", "absence_request", uint(id), &employeeID,
		map[string]any{
			"request_id":      id,
			"absence_type_id": updated.AbsenceTypeID,
			"start_date":      updated.StartDate,
			"end_date":        updated.EndDate,
			"total_days":      updated.TotalDays,
		}, r)
	utils.WriteSuccess(w, http.StatusOK, "OK", updated)
}

// POST /api/v1/absences/requests/{id}/submit — move own DRAFT to PENDING
func (h *Handler) handleSubmitRequest(w http.ResponseWriter, r *http.Request) {
	userID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request ID", err.Error())
		return
	}

	req, err := h.store.GetRequestByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch request", err.Error())
		return
	}
	if req == nil {
		utils.WriteError(w, http.StatusNotFound, "Request not found", "")
		return
	}
	if req.EmployeeID != employeeID {
		utils.WriteError(w, http.StatusForbidden, "Forbidden", "can only submit own requests")
		return
	}
	if req.Status != "DRAFT" {
		utils.WriteError(w, http.StatusConflict, "Cannot submit", "only DRAFT requests can be submitted")
		return
	}

	// recompute in case dates or holidays changed since the draft was created
	days, err := h.computeBusinessDays(employeeID, req.StartDate, req.EndDate)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid dates", err.Error())
		return
	}
	if !h.requireBalance(w, employeeID, req.AbsenceTypeID, req.StartDate, days) {
		return
	}

	req.TotalDays = days
	req.Status = "PENDING"

	updated, err := h.store.UpdateRequest(*req)
	if err != nil {
		if errors.Is(err, ErrOverlap) {
			utils.WriteError(w, http.StatusConflict, "Overlapping request", err.Error())
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Failed to submit request", err.Error())
		return
	}

	h.logAudit("absence.request.submit", "absence_request", uint(id), &employeeID,
		map[string]any{"request_id": id, "total_days": days}, r)
	utils.WriteSuccess(w, http.StatusOK, "OK", updated)
}

// ── leave balance ─────────────────────────────────────────────

// GET /api/v1/absences/balance/me — own balance summary
func (h *Handler) handleGetMyBalance(w http.ResponseWriter, r *http.Request) {
	userID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	balance, err := h.store.GetBalanceByEmployee(employeeID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch balance", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", balance)
}

// GET /api/v1/absences/balance/{id} — any employee's balance (admin)
func (h *Handler) handleGetBalanceByID(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid employee ID", err.Error())
		return
	}

	balance, err := h.store.GetBalanceByEmployee(uint(employeeID))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch balance", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", balance)
}

// POST /api/v1/absences/balance/grant — admin grants days (ACCRUAL)
func (h *Handler) handleGrantBalance(w http.ResponseWriter, r *http.Request) {
	actorID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	var payload types.GrantBalancePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	year := payload.AccrualYear
	if year == 0 {
		year = time.Now().Year()
	}

	entry, err := h.store.AddLedgerEntry(types.LeaveBalanceEntry{
		EmployeeID:    payload.EmployeeID,
		AbsenceTypeID: payload.AbsenceTypeID,
		EntryType:     "ACCRUAL",
		Days:          payload.Days,
		AccrualYear:   year,
		ExpiresAt:     payload.ExpiresAt,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to grant days", err.Error())
		return
	}

	log.Printf("Granted %.1f days of type %d to employee %d (year %d)",
		payload.Days, payload.AbsenceTypeID, payload.EmployeeID, year)
	h.logAudit("balance.grant", "leave_balance", payload.EmployeeID, &actorID,
		map[string]any{
			"employee_id":     payload.EmployeeID,
			"absence_type_id": payload.AbsenceTypeID,
			"days":            payload.Days,
			"year":            year,
		}, r)
	utils.WriteSuccess(w, http.StatusCreated, "Created", entry)
}

// PUT /api/v1/absences/balance/adjust — admin manual adjustment (+/-)
func (h *Handler) handleAdjustBalance(w http.ResponseWriter, r *http.Request) {
	actorID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	var payload types.AdjustBalancePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	year := payload.AccrualYear
	if year == 0 {
		year = time.Now().Year()
	}

	entry, err := h.store.AddLedgerEntry(types.LeaveBalanceEntry{
		EmployeeID:    payload.EmployeeID,
		AbsenceTypeID: payload.AbsenceTypeID,
		EntryType:     "MANUAL_ADJUST",
		Days:          payload.Days,
		AccrualYear:   year,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to adjust days", err.Error())
		return
	}

	log.Printf("Manual adjustment %.1f days of type %d for employee %d",
		payload.Days, payload.AbsenceTypeID, payload.EmployeeID)
	h.logAudit("balance.adjust", "leave_balance", payload.EmployeeID, &actorID,
		map[string]any{
			"employee_id":     payload.EmployeeID,
			"absence_type_id": payload.AbsenceTypeID,
			"days":            payload.Days,
		}, r)
	utils.WriteSuccess(w, http.StatusOK, "OK", entry)
}

// ── leave policies + rollover (admin) ──────────────────────────

// GET /api/v1/absences/balance/my-policies — active policy per absence type for the caller
func (h *Handler) handleGetMyPolicies(w http.ResponseWriter, r *http.Request) {
	userID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	employeeID, err := h.resolveEmployeeID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Employee profile required", err.Error())
		return
	}

	// resolve the active policy for each absence type, using today as reference
	result := []map[string]any{}
	today := time.Now().Format("2006-01-02")
	for _, tid := range []uint{1, 2, 3, 4, 5, 6} {
		p, err := h.store.GetActivePolicy(employeeID, tid, today)
		if err != nil || p == nil {
			continue // no policy mapped → skip
		}
		result = append(result, map[string]any{
			"absence_type_id": tid,
			"policy":          p,
		})
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", result)
}

// POST /api/v1/absences/balance/rollover — run annual accrual/carry-over/expiry
func (h *Handler) handleRollover(w http.ResponseWriter, r *http.Request) {
	actorID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	var payload types.RolloverRequest
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Manual rollover is time-boxed: last week of December (current year) or
	// first week of January (previous year).
	allowedYear, allowed := rolloverWindow(time.Now())
	if !allowed {
		utils.WriteError(w, http.StatusConflict, "Rollover unavailable",
			"manual rollover is only allowed in the last week of December (current year) or the first week of January (previous year)")
		return
	}

	year := payload.Year
	if year == 0 {
		year = allowedYear
	}
	if year != allowedYear {
		utils.WriteError(w, http.StatusConflict, "Invalid rollover year",
			fmt.Sprintf("only year %d can be rolled over at this time", allowedYear))
		return
	}

	report, err := h.store.RolloverYear(year)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Rollover failed", err.Error())
		return
	}

	log.Printf("Rollover for %d: accrued=%d carried=%d expired=%d skipped=%d",
		year, len(report.Accrued), len(report.Carried), len(report.Expired), report.SkippedCnt)
	h.logAudit("balance.rollover", "leave_balance", 0, &actorID,
		map[string]any{"year": year}, r)
	utils.WriteSuccess(w, http.StatusOK, "OK", report)
}

// rolloverWindow reports the year a manual rollover may target, and whether a
// manual rollover is allowed right now:
//   - last week of December (25–31) → current year
//   - first week of January (1–7)   → previous year
//
// Outside that window a manual rollover is not offered (the scheduler still
// performs the accrual automatically).
func rolloverWindow(now time.Time) (year int, allowed bool) {
	switch now.Month() {
	case time.December:
		if now.Day() >= 25 {
			return now.Year(), true
		}
	case time.January:
		if now.Day() <= 7 {
			return now.Year() - 1, true
		}
	}
	return 0, false
}

// GET /api/v1/absences/policies — list all leave policies
func (h *Handler) handleGetPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.store.GetPolicies()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch policies", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", policies)
}

// GET /api/v1/absences/policies/employee/{id} — assignments for an employee
func (h *Handler) handleGetEmployeePolicies(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid employee ID", err.Error())
		return
	}

	assignments, err := h.store.GetEmployeePolicyAssignments(uint(employeeID))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch assignments", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", assignments)
}

// POST /api/v1/absences/policies/assign — assign a policy to an employee
func (h *Handler) handleAssignPolicy(w http.ResponseWriter, r *http.Request) {
	actorID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	var payload types.AssignEmployeePolicyPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	if err := h.store.AssignEmployeePolicy(types.EmployeePolicyAssignment{
		EmployeeID:    payload.EmployeeID,
		AbsenceTypeID: payload.AbsenceTypeID,
		PolicyID:      payload.PolicyID,
		ValidFrom:     payload.ValidFrom,
		ValidTo:       payload.ValidTo,
	}); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to assign policy", err.Error())
		return
	}

	log.Printf("Policy %d assigned to employee %d for absence type %d from %s",
		payload.PolicyID, payload.EmployeeID, payload.AbsenceTypeID, payload.ValidFrom)
	h.logAudit("policy.assign", "employee_leave_policy", payload.EmployeeID, &actorID,
		map[string]any{
			"employee_id":     payload.EmployeeID,
			"absence_type_id": payload.AbsenceTypeID,
			"policy_id":       payload.PolicyID,
			"valid_from":      payload.ValidFrom,
		}, r)
	utils.WriteSuccess(w, http.StatusCreated, "Created", map[string]string{"message": "Policy assigned"})
}

// POST /api/v1/absences/maintenance/run — run the scheduled leave jobs now
// (annual accrual + carry-over, monthly accrual, expiry). Every job is
// idempotent, so this is safe to call at any time.
func (h *Handler) handleRunMaintenance(w http.ResponseWriter, r *http.Request) {
	actorID, _, err := extractClaims(r)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	year := time.Now().Year()

	annual, err := h.store.RolloverYear(year)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Annual rollover failed", err.Error())
		return
	}
	monthly, err := h.store.RunMonthlyAccrual(time.Now().Format("2006-01"))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Monthly accrual failed", err.Error())
		return
	}
	expired, err := h.store.RunExpiration()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Expiration failed", err.Error())
		return
	}

	h.logAudit("maintenance.run", "leave_balance", 0, &actorID,
		map[string]any{"year": year}, r)

	utils.WriteSuccess(w, http.StatusOK, "OK", map[string]any{
		"annual_rollover": annual,
		"monthly_accrual": monthly,
		"expiration":      expired,
	})
}
