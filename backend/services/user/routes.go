package user

import (
	"database/sql"
	"fmt"
	"log"
	"main/middleware"
	"main/services/auth"
	typesAudit "main/types/audit"
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
	db         *sql.DB
	store      types.UserStore
	validator  *utils.Validator
	auditStore typesAudit.AuditStore
}

// NewHandler creates a new user Handler with the provided store and validator.
func NewHandler(db *sql.DB, store types.UserStore, auditStore typesAudit.AuditStore, v *utils.Validator) *Handler {
	return &Handler{db: db, store: store, auditStore: auditStore, validator: v}
}

// RegisterPublicRoutes registers public (unauthenticated) user routes on the given router.
// Only authentication-related endpoints are public.
func (h *Handler) RegisterPublicRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/register", h.handleRegister).Methods("POST")
	router.HandleFunc("/refresh", h.handleRefresh).Methods("POST")
}

// RegisterProtectedRoutes registers routes that require authentication.
// Admin-only user-management endpoints are guarded with permission middleware.
func (h *Handler) RegisterProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/me", h.handleMe).Methods("GET")
	router.HandleFunc("/logout", h.handleLogout).Methods("POST")
	router.HandleFunc("/change-password", h.handleChangePassword).Methods("PUT")

	// admin / platform-owner endpoints — permission-protected
	router.Handle("/permission", middleware.RequirePermission(h.db, "users.manage", http.HandlerFunc(h.handlePremissions))).Methods("GET")
	router.Handle("/change-status", middleware.RequirePermission(h.db, "users.manage", http.HandlerFunc(h.handleChangeStatus))).Methods("PUT")
	router.Handle("/users", middleware.RequirePermission(h.db, "users.manage", http.HandlerFunc(h.handleGetAllUsers))).Methods("GET")
	router.Handle("/users/{id}", middleware.RequirePermission(h.db, "users.manage", http.HandlerFunc(h.hadnleGetUserByIdWithRole))).Methods("GET")
	router.Handle("/users/{id}/role", middleware.RequirePermission(h.db, "roles.manage", http.HandlerFunc(h.handleSetUserRole))).Methods("PUT")
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
		utils.WriteError(w, http.StatusUnauthorized, "Nedostaje refresh token.", "")
		return
	}

	tokenHash := auth.HashToken(cookie.Value)

	// find token in DB
	userID, err := h.store.GetUserIDByRefreshToken(tokenHash)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Nevažeći refresh token.", "")
		return
	}

	// load user + role
	user, err := h.store.GetUserByID(int(userID))
	if err != nil || user == nil || !user.IsActive {
		utils.WriteError(w, http.StatusUnauthorized, "Nalog nije dostupan.", "")
		return
	}

	role, err := h.store.GetUserRole(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Uloga nije pronađena.", "")
		return
	}

	// issue NEW access token
	accessToken, err := auth.GenerateToken(user.ID, user.Email, role)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška pri kreiranju tokena.", "")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "Token je osvežen.", map[string]string{
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
		utils.WriteError(w, http.StatusBadRequest, "Neispravan zahtev.", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Podaci nisu ispravni.", err.Error())
		return
	}

	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	// log.Print(payload)
	// 1) find user
	user, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška u bazi podataka.", err.Error())
		return
	}

	if user == nil {
		utils.WriteError(w, http.StatusUnauthorized, "Lozinka ili email adresa ne postoje. Molimo vas unesite ponovo.", "")
		return
	}

	// 2) check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Lozinka ili email adresa ne postoje. Molimo vas unesite ponovo.", "")
		return
	}

	// 3) reject deactivated accounts
	if !user.IsActive {
		utils.WriteError(w, http.StatusForbidden, "Nalog je deaktiviran.", "")
		return
	}

	// 4) load role
	role, err := h.store.GetUserRole(user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Uloga nije pronađena.", err.Error())
		return
	}

	// 4) generate JWT
	token, err := auth.GenerateToken(user.ID, user.Email, role)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška pri kreiranju tokena.", err.Error())
		return
	}

	// 5) generate refresh token
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška pri kreiranju tokena.", err.Error())
		return
	}

	// hash before storing
	refreshTokenHash := auth.HashToken(refreshToken)

	// store in DB
	err = h.store.SaveRefreshToken(user.ID, refreshTokenHash, int(auth.SessionTTL.Minutes()))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška pri kreiranju sesije.", err.Error())
		return
	}

	// response
	response := map[string]any{
		"accessToken": token,
		"user": map[string]any{
			"id":        user.ID,
			"email":     user.Email,
			"role":      role,
			"createdAt": user.CreatedAt,
			"isActive":  user.IsActive,
		},
	}

	// Cookie path must cover both /api/v1/refresh and /api/v1/logout so the
	// token is sent to (and can be revoked by) the logout endpoint.
	// Bez `Expires`/`MaxAge` → ovo je **„session cookie“**: pretraživač ga briše
	// kada se zatvori, pa je sledeći pristup ponovo login stranica.
	// Trajanje sesije **ne odreduje kolačić** nego server (`auth.SessionTTL`, preko
	// `refresh_tokens.expires_at`), da se ne može produžiti sa klijenta.
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1",
	})

	utils.WriteSuccess(w, http.StatusOK, "Prijava uspešna.", response)
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
		utils.WriteError(w, http.StatusBadRequest, "Neispravan zahtev.", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Neispravan zahtev.", err.Error())
		// log.Print(err)
		return
	}
	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))

	log.Printf("paylod: %s", payload.Email)
	existingUser, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška u bazi podataka.", err.Error())
		return
	}

	log.Printf("HERE PRINT %+v\n", existingUser)
	if existingUser != nil {
		utils.WriteError(w, http.StatusConflict, "Korisnik sa ovom email adresom već postoji.", "")
		return
	}
	log.Print(existingUser)
	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	// log.Print("ER")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška na serveru.", err.Error())
		return
	}

	createdUser, err := h.store.CreateUserWithRole(types.User{
		Email:    payload.Email,
		Password: string(hash),
	}, "EMPLOYEE", nil)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška na serveru.", err.Error())
		return
	}

	log.Printf("User: %+v", createdUser)
	response := types.UserResponse{
		ID:    createdUser.ID,
		Email: createdUser.Email,
	}

	if err := h.auditStore.Log(typesAudit.AuditEntry{
		Action:    "user.register",
		Entity:    "user",
		EntityID:  createdUser.ID,
		ActorID:   nil,
		Details:   map[string]any{"email": payload.Email, "role": "EMPLOYEE"},
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}); err != nil {
		log.Printf("WARNING: Failed to write audit log: %v", err)
	}

	log.Printf("Response in handleRegister: %+v", response)
	utils.WriteSuccess(w, http.StatusCreated, "Korisnik je registrovan.", response)
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
	utils.WriteSuccess(w, http.StatusOK, "OK", map[string]string{
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
		utils.WriteError(w, http.StatusBadRequest, "Neispravan zahtev.", err.Error())
		return
	}

	fmt.Printf("ID: %s", payload)

	permissions, err := h.store.GetUserPremissions(payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška na serveru.", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", permissions)
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
		utils.WriteError(w, http.StatusBadRequest, "Neispravan zahtev.", err.Error())
		return
	}

	actorID, _, ok := middleware.RoleFromContext(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Nevažeći token.", "")
		return
	}

	// prevent an admin from deactivating their own account (self-lockout)
	if payload.UserID == actorID {
		utils.WriteError(w, http.StatusBadRequest, "Ne možete menjati sopstveni status.", "")
		return
	}

	user, err := h.store.ChangeUserStatus(payload.UserID, payload.IsActive)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška pri promeni statusa korisnika.", err.Error())
		return
	}

	// revoke all refresh tokens so a deactivated user cannot obtain new access tokens
	if !payload.IsActive {
		_ = h.store.RevokeAllUserTokens(payload.UserID)
	}

	h.auditStore.Log(typesAudit.AuditEntry{
		Action:    "user.change_status",
		Entity:    "user",
		EntityID:  payload.UserID,
		ActorID:   &actorID,
		Details:   map[string]any{"is_active": payload.IsActive, "email": user.Email},
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})

	utils.WriteSuccess(w, http.StatusOK, "Status korisnika je promenjen.", map[string]string{
		"user": fmt.Sprintf("ID: %d, Email: %s, IsActive: %t", user.ID, user.Email, user.IsActive),
	})
}

func (h *Handler) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.GetAllUsers()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška pri dohvatanju korisnika.", err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", users)
}

func (h *Handler) hadnleGetUserByIdWithRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Neispravan ID korisnika.", err.Error())
		return
	}

	user, role, err := h.store.GetUserByIDWithRole(id)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška pri dohvatanju korisnika.", err.Error())
		return
	}

	response := map[string]any{
		"user": user,
		"role": role,
	}

	utils.WriteSuccess(w, http.StatusOK, "OK", response)
}

// handleLogout revokes the refresh token and clears the cookie.
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refreshToken")
	if err == nil {
		_ = h.store.RevokeRefreshToken(auth.HashToken(cookie.Value))
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		HttpOnly: true,
		Path:     "/api/v1",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})

	// audit
	userID, _, ok := middleware.RoleFromContext(r)
	if ok {
		h.auditStore.Log(typesAudit.AuditEntry{
			Action:    "user.logout",
			Entity:    "user",
			EntityID:  userID,
			ActorID:   &userID,
			IP:        r.RemoteAddr,
			UserAgent: r.UserAgent(),
		})
	}

	utils.WriteSuccess(w, http.StatusOK, "Odjava uspešna.", nil)
}

// handleChangePassword lets an authenticated user change their own password.
func (h *Handler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := middleware.RoleFromContext(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Nevažeći token.", "")
		return
	}

	var payload types.ChangePasswordPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Neispravan zahtev.", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Podaci nisu ispravni.", err.Error())
		return
	}

	user, err := h.store.GetUserByID(int(userID))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška u bazi podataka.", err.Error())
		return
	}

	// verify current password — 400 (a ne 401) da klijent ne pomisli da je sesija
	// istekla i ne pokušava tiho osvežavanje tokena
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.CurrentPassword)); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Trenutna lozinka nije ispravna.", "")
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), 12)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška na serveru.", err.Error())
		return
	}

	if err := h.store.ChangePassword(userID, string(newHash)); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Greška pri promeni lozinke.", err.Error())
		return
	}

	// invalidate all existing refresh tokens after a password change
	_ = h.store.RevokeAllUserTokens(userID)

	h.auditStore.Log(typesAudit.AuditEntry{
		Action:    "user.change_password",
		Entity:    "user",
		EntityID:  userID,
		ActorID:   &userID,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})

	utils.WriteSuccess(w, http.StatusOK, "Lozinka je promenjena.", nil)
}

// handleSetUserRole changes the role of a user (platform owner only).
func (h *Handler) handleSetUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Neispravan ID korisnika.", err.Error())
		return
	}

	var payload types.AssignRolePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Neispravan zahtev.", err.Error())
		return
	}

	if err := h.validator.V.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Podaci nisu ispravni.", err.Error())
		return
	}

	actorID, _, ok := middleware.RoleFromContext(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "Nevažeći token.", "")
		return
	}

	// prevent an admin from changing their own role (self-lockout)
	if uint(id) == actorID {
		utils.WriteError(w, http.StatusBadRequest, "Ne možete menjati sopstvenu ulogu.", "")
		return
	}

	if err := h.store.SetUserRole(uint(id), payload.RoleName); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Greška pri dodeli uloge.", err.Error())
		return
	}

	h.auditStore.Log(typesAudit.AuditEntry{
		Action:    "user.set_role",
		Entity:    "user",
		EntityID:  uint(id),
		ActorID:   &actorID,
		Details:   map[string]any{"role": payload.RoleName},
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})

	utils.WriteSuccess(w, http.StatusOK, "Uloga je promenjena.", map[string]string{"role": payload.RoleName})
}
