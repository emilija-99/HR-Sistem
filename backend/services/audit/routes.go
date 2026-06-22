package audit

import (
	types "main/types/audit"
	"main/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	store types.AuditStore
}

func NewHandler(store types.AuditStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/audit/{entity}/{id}", h.handleGetAuditLogs).Methods("GET")
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
