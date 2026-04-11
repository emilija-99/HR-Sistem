package user

import (
	"log"
	"main/services/auth"
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

func (h *Handler) RegisterPublicRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/register", h.handleRegister).Methods("POST")
}

func (h *Handler) RegisterProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/me", h.handleMe).Methods("GET")
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.LoginUserPayload
	log.Print(payload)
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	log.Print(payload)
	// 1) find user
	user, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	if user == nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid credentials", "")
		return
	}

	// 2) check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid credentials", "")
		return
	}

	// 3) load role
	role, err := h.store.GetUserRole(user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Role not found", err.Error())
		return
	}

	// 4) generate JWT
	token, err := auth.GenerateToken(user.ID, user.Email, role)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Token generation failed", err.Error())
		return
	}

	// 5) response DTO
	response := map[string]any{
		"status": "success",
		"data": map[string]any{
			"token": token,
			"user": map[string]any{
				"id":        user.ID,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
				"email":     user.Email,
				"role":      role,
			},
		},
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		log.Print(err)
		return
	}
	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))

	log.Printf("paylod: %s", payload.Email)
	existingUser, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	log.Printf("HERE PRINT %+v\n", existingUser)
	if existingUser != nil {
		utils.WriteError(w, http.StatusBadRequest, "User already exists", "")
		return
	}
	log.Print(existingUser)
	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	log.Print("ER")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	createdUser, err := h.store.CreateUserWithRole(types.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  string(hash),
	}, "user")
	log.Print("createdUser: %s\n", err)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	log.Print("createdUser-- %v\n", createdUser)
	response := types.UserResponse{
		ID:        createdUser.ID,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
		Email:     createdUser.Email,
	}
	log.Print("response %+v", response)
	utils.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "You accessed protected route",
	})
}
