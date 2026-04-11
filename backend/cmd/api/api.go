package api

import (
	"database/sql"
	"log"
	"main/middleware"
	"main/services/absence"
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

	router.Use(middleware.CORS)
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	absenceStore := absence.NewStore(s.db)

	absenceHandler := absence.NewHandler(absenceStore, utils.NewValidator())

	userHandler := user.NewHandler(userStore, utils.NewValidator())

	userHandler.RegisterPublicRoutes(subrouter)
	absenceHandler.RegisterPublicRoutes(subrouter)

	protected := subrouter.PathPrefix("").Subrouter()
	protected.Use(middleware.JWTAuth)

	userHandler.RegisterProtectedRoutes(protected)

	log.Println("Listening on: ", s.addr)

	return http.ListenAndServe(s.addr, router)
}
