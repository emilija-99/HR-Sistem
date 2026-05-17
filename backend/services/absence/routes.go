package absence

import (
	"fmt"
	"log"
	types "main/types/absence"
	"main/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.AbsenceStore
	validator *utils.Validator
}

func NewHandler(store types.AbsenceStore, v *utils.Validator) *Handler {
	return &Handler{store: store, validator: v}
}

func (h *Handler) RegisterPublicRoutes(router *mux.Router) {
	log.Print("Registering public routes for absence")
	router.HandleFunc("/absence", h.handleGetAbsence).Methods("GET")
	router.HandleFunc("/absence/{id}", h.handleGetAbsenceById).Methods("GET")
	router.HandleFunc("/absence/{id}", h.handlePatchAbsenceById).Methods("PATCH")
	router.HandleFunc("/absence/change-status/{id}", h.handleChangeStatusOfAbsence).Methods("PUT")

}

func (h *Handler) RegisterProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/absence", h.handleGetAbsence).Methods("GET")
}

func (h *Handler) handleGetAbsence(w http.ResponseWriter, r *http.Request) {
	var payload *types.AbsenceResponse
	var err error
	payload, err = h.store.GetAllAbsenceTypes()

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get absence types", err.Error())
		return
	}

	fmt.Printf("Payload: %+v\n", payload)
	utils.WriteJSON(w, http.StatusOK, payload)

}

func (h *Handler) handleGetAbsenceById(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)
	var error error
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid absence type ID", err.Error())
		return
	}

	fmt.Printf("ID: %v", id)

	log.Print("Handling GET /absence/{id} request")

	var payload *types.AbsenceTypes

	payload, error = h.store.GetAbsenceTypeById(id)

	if error != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get absence type", error.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) handlePatchAbsenceById(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)
	var error error
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid absence type ID", err.Error())
		return
	}

	var payload types.AbsenceTypePatchRequest
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}
	log.Printf("Payload after parsing: %+v", payload)

	if payload.TypeName == nil && payload.Code == nil && payload.IsPaid == nil && payload.Status == nil {
		utils.WriteError(w, http.StatusBadRequest, "At least one field must be provided for update", "No fields to update")
		return
	}

	log.Printf("Handling PUT /absence/{id} request with payload: %+v", payload)

	var response *types.AbsenceTypes

	response, error = h.store.PatchAbsenceTypeById(id, payload)

	if error != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update absence type", error.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleChangeStatusOfAbsence(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)
	var error error
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	log.Printf("ID for changing status: %d", id)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid absence type ID", err.Error())
		return
	}

	var payload *types.ChangeStatusRequest
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	log.Printf("Payload after parsing: %+v", payload)
	log.Printf("Handling PATCH /absence/{id} request with payload: %+v", payload)

	var response *types.AbsenceTypes

	response, error = h.store.ChangeStatusOfAbstenceType(id, payload.Status)

	if error != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to update absence type", error.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
