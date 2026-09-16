package audit

import (
	"database/sql"
	"main/middleware"
	types "main/types/audit"
	"main/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	db    *sql.DB
	store types.AuditStore
}

func NewHandler(db *sql.DB, store types.AuditStore) *Handler {
	return &Handler{db: db, store: store}
}

func (h *Handler) RegisterProtectedRoutes(router *mux.Router) {
	// audit log is viewable by roles with reports.view (HR_ADMIN, PLATFORM_ADMIN)
	router.Handle("/audit/recent", middleware.RequirePermission(h.db, "reports.view", http.HandlerFunc(h.handleGetRecent))).Methods("GET")
	router.Handle("/audit/actions", middleware.RequirePermission(h.db, "reports.view", http.HandlerFunc(h.handleGetActions))).Methods("GET")
	router.Handle("/audit/{entity}/{id}", middleware.RequirePermission(h.db, "reports.view", http.HandlerFunc(h.handleGetAuditLogs))).Methods("GET")
}

func (h *Handler) handleGetAuditLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	entries, err := h.store.GetByEntity(vars["entity"], uint(id))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch audit logs", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, entries)
}

func (h *Handler) handleGetRecent(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	entity := r.URL.Query().Get("entity")
	action := r.URL.Query().Get("action")

	entries, err := h.store.GetRecent(limit, entity, action)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch audit logs", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, entries)
}

func (h *Handler) handleGetActions(w http.ResponseWriter, r *http.Request) {
	actions, err := h.store.DistinctActions()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch audit actions", err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, actions)
}
