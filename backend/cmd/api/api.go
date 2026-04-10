package api

import (
	"database/sql"
	"log"
	"main/middleware"
	"main/services/user"
	"main/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore, utils.NewValidator())

	userHandler.RegisterPublicRoutes(subrouter)

	protected := subrouter.PathPrefix("").Subrouter()
	protected.Use(middleware.JWTAuth)

	userHandler.RegisterProtectedRoutes(subrouter)

	log.Println("Listening on: ", s.addr)

	return http.ListenAndServe(s.addr, router)
}
