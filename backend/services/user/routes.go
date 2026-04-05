package user

import (
	"fmt"
	"main/types"
	"main/utils"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	store     types.UserStore
	validator *utils.Validator
}

func NewHandler(store types.UserStore, v *utils.Validator) *Handler {
	return &Handler{store: store, validator: v}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/register", h.handleRegister).Methods("POST")
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrInvalidJSON)
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteErrorCustom(w, http.StatusBadRequest, err)
		return
	}
	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))

	_, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteErrorCustom(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrInternalServer)
		return
	}

	userID, err := h.store.CreateUser(types.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  string(hash),
	})

	err = h.store.AssignRole(userID, "user")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrInternalServer)
		return
	}

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrInternalServer)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, nil)
}
