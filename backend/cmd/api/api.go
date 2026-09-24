package api

import (
	"database/sql"
	"log"
	"main/middleware"
	"main/services/absence"
	"main/services/attendance"
	"main/services/audit"
	"main/services/employee"
	"main/services/user"
	"main/utils"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	// Create router and global middleware
	router := mux.NewRouter()
	router.Use(middleware.CORS)
	router.Use(middleware.PrometheusMetrics)

	// Prometheus scrape endpoint (unauthenticated, for the collector)
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	empStore := employee.NewStore(s.db)
	auditStore := audit.NewStore(s.mongo)
	absenceStore := absence.NewStore(s.db)

	userHandler := user.NewHandler(s.db, userStore, auditStore, utils.NewValidator())
	empHandler := employee.NewHandler(s.db, empStore, auditStore, utils.NewValidator(), absenceStore)
	absenceHandler := absence.NewHandler(s.db, absenceStore, empStore, auditStore, utils.NewValidator())
	attendanceHandler := attendance.NewHandler(s.db, attendance.NewStore(s.db), empStore)

	userHandler.RegisterPublicRoutes(subrouter)
	absenceHandler.RegisterPublicRoutes(subrouter)

	protected := subrouter.PathPrefix("").Subrouter()
	protected.Use(middleware.JWTAuth(s.db))

	userHandler.RegisterProtectedRoutes(protected)
	empHandler.RegisterProtectedRoutes(protected)
	absenceHandler.RegisterProtectedRoutes(protected)
	attendanceHandler.RegisterProtectedRoutes(protected)

	auditHandler := audit.NewHandler(s.db, auditStore)
	auditHandler.RegisterProtectedRoutes(protected)

	log.Println("Listening on: ", s.addr)
	return http.ListenAndServe(s.addr, router)
}
