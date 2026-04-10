package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"main/types"
	"main/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestRegister_UserServiceHandlers(t *testing.T) {
	userStore := &mockUserStore{}
	validator := utils.NewValidator()
	handler := NewHandler(userStore, validator)

	payload := types.RegisterUserPayload{
		FirstName: "test1",
		LastName:  "test1",
		Email:     "asd@gmail.com",
		Password:  "test1",
	}

	marshalled, err := json.Marshal(payload)

	if err != nil {
		log.Fatal("bad marshal")
	}

	t.Run("should fail if the user payload is invalid", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(marshalled))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := mux.NewRouter()

		router.HandleFunc("/register", handler.handleRegister)

		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}

type mockUserStore struct {
	getUserErr    error
	createErr     error
	assignRoleErr error

	createdUser  *types.User
	assignedRole string
	nextUserID   uint
}

func newTestHandler(store *mockUserStore) *Handler {
	v := utils.NewValidator()
	return NewHandler(store, v)
}

func TestRegister_InvalidJSON(t *testing.T) {
	store := &mockUserStore{}
	h := newTestHandler(store)

	rr := performRequest(h.handleRegister, []byte(`{invalid json`))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
func TestRegister_ValidationFail(t *testing.T) {
	store := &mockUserStore{}
	h := newTestHandler(store)

	p := validPayload()
	p.Password = "123" // invalid

	body, _ := json.Marshal(p)
	rr := performRequest(h.handleRegister, body)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	store := &mockUserStore{
		getUserErr: errors.New("found"),
	}
	h := newTestHandler(store)

	body, _ := json.Marshal(validPayload())
	rr := performRequest(h.handleRegister, body)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRegister_CreateUserFails(t *testing.T) {
	store := &mockUserStore{
		createErr: errors.New("db down"),
	}
	h := newTestHandler(store)

	body, _ := json.Marshal(validPayload())
	rr := performRequest(h.handleRegister, body)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
func performRequest(handler http.HandlerFunc, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/register", handler)
	router.ServeHTTP(rr, req)
	return rr
}

func validPayload() types.RegisterUserPayload {
	return types.RegisterUserPayload{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@doe.com",
		Password:  "strongpass123",
	}
}
func (m *mockUserStore) GetUserByEmail(email string) (*types.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	return nil, nil
}

func (m *mockUserStore) GetUserByID(id int) (*types.User, error) { return nil, nil }

func (m *mockUserStore) CreateUser(u types.User) (uint, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}

	if m.assignRoleErr != nil {
		return 0, m.assignRoleErr
	}

	if m.nextUserID == 0 {
		m.nextUserID = 1
	}

	return m.nextUserID, nil
}

func (m *mockUserStore) EmailExists(email string) string { return "" }
func (s *mockUserStore) AssignRole(userID uint, roleName string) error {
	return nil
}
func TestRegister_Success(t *testing.T) {
	store := &mockUserStore{}
	h := newTestHandler(store)

	body, _ := json.Marshal(validPayload())
	rr := performRequest(h.handleRegister, body)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	if store.createdUser == nil {
		t.Fatal("user was not created")
	}

	if store.createdUser.Password == "strongpass123" {
		t.Fatal("password not hashed")
	}
}
