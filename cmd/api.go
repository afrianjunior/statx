package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/prometheus/tsdb"
	"go.uber.org/zap"
)

type APIError struct {
	Error string `json:"error"`
}

type IApi interface{}

type Api struct {
	httpClient *http.Client
	db         *tsdb.DB
	logger     *zap.SugaredLogger
}

func NewApi(
	httpClient *http.Client,
	db *tsdb.DB,
	logger *zap.SugaredLogger,
) IApi {
	return Api{
		httpClient: httpClient,
		db:         db,
		logger:     logger,
	}
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIError{Error: message})
}

func (s *Api) Start(port string) {
	// Start monitoring each target
	// for _, target := range s.targets {
	// 	go s.checkStatus(target)
	// }

	// Setup and start HTTP server
	router := s.setupRouter()
	s.logger.Infof("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		s.logger.Fatalf("Server error: %v", err)
	}
}

func (s *Api) setupRouter() *chi.Mux {
	r := chi.NewRouter()

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
		// Query status data
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			url := r.URL.Query().Get("url")
			if url == "" {
				s.jsonError(w, "url parameter is required", http.StatusBadRequest)
				return
			}

			timeRange, err := s.parseTimeRange(r)
			if err != nil {
				s.jsonError(w, fmt.Sprintf("invalid time range: %v", err), http.StatusBadRequest)
				return
			}

			results, err := s.QueryStatus(url, timeRange)
			if err != nil {
				s.jsonError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			s.jsonResponse(w, results)
		})

		// List all targets
		r.Get("/targets", func(w http.ResponseWriter, r *http.Request) {
			s.jsonResponse(w, s.targets)
		})

		// Get monitoring stats
		r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
			stats := map[string]interface{}{
				"targets_count":     len(s.targets),
				"retention_period":  s.config.RetentionPeriod.String(),
				"block_duration":    s.config.BlockDuration.String(),
				"max_blocks":        s.config.MaxBlocksToRead,
				"max_samples_daily": s.config.MaxSamplesPerDay,
			}
			s.jsonResponse(w, stats)
		})
	})

	return r
}
