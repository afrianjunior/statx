package cmd

import (
	"database/sql"
	"net/http"

	"github.com/afrianjunior/statx/internal/exposer"
	_ "github.com/glebarez/go-sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/prometheus/tsdb"
	"go.uber.org/zap"
)

type APIError struct {
	Error string `json:"error"`
}

type Rest interface {
	Start(port string)
}

type rest struct {
	httpClient *http.Client
	tsdb       *tsdb.DB
	db         *sql.DB
	logger     *zap.SugaredLogger
}

func NewRest(
	httpClient *http.Client,
	tsdb *tsdb.DB,
	db *sql.DB,
	logger *zap.SugaredLogger,
) Rest {
	return &rest{
		httpClient: httpClient,
		tsdb:       tsdb,
		db:         db,
		logger:     logger,
	}
}

func (s *rest) Start(port string) {
	router := s.setupRouter()
	s.logger.Infof("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		s.logger.Fatalf("Server error: %v", err)
	}
}

func (s *rest) setupRouter() *chi.Mux {
	r := chi.NewRouter()
	exposerService := exposer.NewExposerService(s.tsdb)

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API Routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/status", exposer.StatusHandler(exposerService))
	})

	return r
}
