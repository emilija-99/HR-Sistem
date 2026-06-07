package api

import (
	"database/sql"
	"log"
	"main/middleware"
	"main/services/audit"
	"main/services/employee"
	"main/services/user"
	"main/utils"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

type APIServer struct {
	addr  string
	db    *sql.DB
	mongo *mongo.Database
}

func NewAPIServer(addr string, db *sql.DB, mongoDB *mongo.Database) *APIServer {
	return &APIServer{addr: addr, db: db, mongo: mongoDB}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	router.Use(middleware.CORS)
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	empStore := employee.NewStore(s.db)
	auditStore := audit.NewStore(s.mongo)

	userHandler := user.NewHandler(userStore, auditStore, utils.NewValidator())
	empHandler := employee.NewHandler(empStore, utils.NewValidator())

	userHandler.RegisterPublicRoutes(subrouter)

	protected := subrouter.PathPrefix("").Subrouter()
	protected.Use(middleware.JWTAuth)

	userHandler.RegisterProtectedRoutes(protected)
	empHandler.RegisterProtectedRoutes(protected)

	auditHandler := audit.NewHandler(auditStore)
	auditHandler.RegisterProtectedRoutes(protected)

	log.Println("Listening on: ", s.addr)
	return http.ListenAndServe(s.addr, router)
}
