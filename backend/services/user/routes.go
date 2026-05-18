package user

import (
	"fmt"
	"log"
	"main/services/auth"
	types "main/types/user"
	"main/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	store     types.UserStore
	validator *utils.Validator
}

// NewHandler creates a new user Handler with the provided store and validator.
func NewHandler(store types.UserStore, v *utils.Validator) *Handler {
	return &Handler{store: store, validator: v}
}

// RegisterPublicRoutes registers public (unauthenticated) user routes on the given router.
func (h *Handler) RegisterPublicRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/register", h.handleRegister).Methods("POST")
	router.HandleFunc("/refresh", h.handleRefresh).Methods("POST")
	router.HandleFunc("/permission", h.handlePremissions).Methods("GET")
	router.HandleFunc("/change-status", h.handleChangeStatus).Methods("PUT")
	router.HandleFunc("/users", h.handleGetAllUsers).Methods("GET")
	router.HandleFunc("/users/{id}", h.hadnleGetUserByIdWithRole).Methods("GET")

}

// RegisterProtectedRoutes registers routes that require authentication.
func (h *Handler) RegisterProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/me", h.handleMe).Methods("GET")
}

// handleRefresh issues a new access token when a valid refresh token cookie is presented.
//
// @Summary Refresh access token
// @Description Issues a new access token when provided a valid refresh token cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param refreshToken cookie string true "Refresh token"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /refresh [post]
func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Missing refresh token", "")
		return
	}

	tokenHash := auth.HashToken(cookie.Value)

	// find token in DB
	userID, err := h.store.GetUserIDByRefreshToken(tokenHash)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid refresh token", "")
		return
	}

	// load user + role
	user, _ := h.store.GetUserByID(int(userID))
	role, _ := h.store.GetUserRole(userID)

	// issue NEW access token
	accessToken, err := auth.GenerateToken(user.ID, user.Email, role)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Token generation failed", "")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"accessToken": accessToken,
	})
}

// handleLogin authenticates a user and returns an access token and sets a refresh token cookie.
//
// @Summary User login
// @Description Authenticates a user with email and password. Returns access token and sets refresh token cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body types.LoginUserPayload true "Login payload"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /login [post]
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.LoginUserPayload
	// log.Print(payload)
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	// log.Print(payload)
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

	// 5) generate refresh token
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Token generation failed", err.Error())
		return
	}

	// hash before storing
	refreshTokenHash := auth.HashToken(refreshToken)

	// store in DB
	err = h.store.SaveRefreshToken(user.ID, refreshTokenHash)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Session creation failed", err.Error())
		return
	}

	// response
	response := map[string]any{
		"status": "success",
		"data": map[string]any{
			"accessToken": token,
			"user": map[string]any{
				"id":        user.ID,
				"email":     user.Email,
				"role":      role,
				"createdAt": user.CreatedAt,
				"isActive":  user.IsActive,
			},
		},
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/refresh",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	utils.WriteJSON(w, http.StatusOK, response)
}

// handleRegister creates a new user and assigns the default role.
//
// @Summary User registration
// @Description Registers a new user with email and password and assigns the "user" role.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body types.RegisterUserPayload true "Registration payload"
// @Success 201 {object} types.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /register [post]
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload types.RegisterUserPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		// log.Print(err)
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
	// log.Print("ER")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	createdUser, err := h.store.CreateUserWithRole(types.User{
		Email:    payload.Email,
		Password: string(hash),
	}, "user")
	log.Print("createdUser: %s\n", err)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	log.Print("createdUser-- %v\n", createdUser)
	response := types.UserResponse{
		ID:    createdUser.ID,
		Email: createdUser.Email,
	}
	log.Print("response %+v", response)
	utils.WriteJSON(w, http.StatusCreated, response)
}

// handleMe returns a simple message for authenticated users.
//
// @Summary Current user info
// @Description Returns a message indicating access to a protected route. Replace with actual user info as needed.
// @Tags users
// @Produce json
// @Success 200 {object} map[string]string
// @Router /me [get]
func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "You accessed protected route",
	})
}

// handlePremissions returns permissions for a user or role.
//
// @Summary Get permissions
// @Description Retrieves permissions for a given user or role.
// @Tags users
// @Accept json
// @Produce json
// @Param payload body types.PermissionRequest true "Permission request"
// @Success 200 {object} interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /permission [get]
func (h *Handler) handlePremissions(w http.ResponseWriter, r *http.Request) {
	var payload types.PermissionRequest

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	fmt.Printf("ID: %s", payload)

	permissions, err := h.store.GetUserPremissions(payload)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, permissions)
}

// handleChangeStatus activates or deactivates a user account.
//
// @Summary Change user status
// @Description Activate or deactivate a user account by ID.
// @Tags users
// @Accept json
// @Produce json
// @Param payload body object true "Status payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /change-status [put]
func (h *Handler) handleChangeStatus(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UserID   uint `json:"id"`
		IsActive bool `json:"isActive"`
	}

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	user, err := h.store.ChangeUserStatus(payload.UserID, payload.IsActive)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to change user status", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "User status changed successfully",
		"user":    fmt.Sprintf("ID: %d, Email: %s, IsActive: %t", user.ID, user.Email, user.IsActive),
	})
}

func (h *Handler) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.GetAllUsers()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get users", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, users)
}

func (h *Handler) hadnleGetUserByIdWithRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	user, role, err := h.store.GetUserByIDWithRole(id)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get user", err.Error())
		return
	}

	response := map[string]any{
		"user": user,
		"role": role,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
