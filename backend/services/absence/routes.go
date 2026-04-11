package absence

import (
	"fmt"
	"log"
	types "main/types/absence"
	"main/utils"
	"net/http"

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
