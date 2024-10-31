package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"go.uber.org/zap"
)

type Target struct {
	URL      string        `json:"url"`
	Interval time.Duration `json:"interval"`
}

type Config struct {
	Targets          []Target      `json:"targets"`
	DBPath           string        `json:"db_path"`
	RetentionPeriod  time.Duration `json:"retention_period"`
	BlockDuration    time.Duration `json:"block_duration"`
	ServerPort       string        `json:"server_port"`
	LogLevel         string        `json:"log_level"`
	MaxBlocksToRead  int64         `json:"max_blocks_to_read"`
	CheckTimeout     time.Duration `json:"check_timeout"`
	RetryAttempts    int           `json:"retry_attempts"`
	RetryDelay       time.Duration `json:"retry_delay"`
	MaxSamplesPerDay int64         `json:"max_samples_per_day"`
}

type StatusMonitor struct {
	targets    []Target
	httpClient *http.Client
	db         *tsdb.DB
	logger     *zap.SugaredLogger
	config     *Config
}

type QueryResult struct {
	URL          string    `json:"url"`
	Timestamp    time.Time `json:"timestamp"`
	Status       int       `json:"status"`
	ResponseTime float64   `json:"response_time"`
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type APIError struct {
	Error string `json:"error"`
}

func loadConfig(path string) (*Config, error) {
	defaultConfig := Config{
		Targets: []Target{
			{URL: "https://example.com", Interval: 30 * time.Second},
			{URL: "https://google.com", Interval: 60 * time.Second},
		},
		DBPath:           filepath.Join("data", "tsdb"),
		RetentionPeriod:  7 * 24 * time.Hour,
		BlockDuration:    2 * time.Hour,
		ServerPort:       "8080",
		LogLevel:         "info",
		MaxBlocksToRead:  1000,
		CheckTimeout:     10 * time.Second,
		RetryAttempts:    3,
		RetryDelay:       5 * time.Second,
		MaxSamplesPerDay: 86400,
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		configJSON, _ := json.MarshalIndent(defaultConfig, "", "  ")
		if err := os.WriteFile(path, configJSON, 0644); err != nil {
			return nil, fmt.Errorf("error creating default config: %v", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config: %v", err)
	}

	return &config, nil
}

func setupLogger(level string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	if level == "debug" {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %v", err)
	}
	return logger.Sugar(), nil
}

func NewStatusMonitor(config *Config) (*StatusMonitor, error) {
	logger, err := setupLogger(config.LogLevel)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(config.DBPath, 0755); err != nil {
		return nil, fmt.Errorf("error creating DB directory: %v", err)
	}

	opts := tsdb.DefaultOptions()
	opts.RetentionDuration = config.RetentionPeriod.Milliseconds()
	opts.MaxBlockDuration = config.BlockDuration.Milliseconds()
	opts.MaxBlockChunkSegmentSize = 256 * 1024 * 1024

	db, err := tsdb.Open(config.DBPath, nil, nil, opts, nil)
	if err != nil {
		return nil, fmt.Errorf("error opening TSDB: %v", err)
	}

	return &StatusMonitor{
		targets: config.Targets,
		httpClient: &http.Client{
			Timeout: config.CheckTimeout,
		},
		db:     db,
		logger: logger,
		config: config,
	}, nil
}

func (s *StatusMonitor) writeToTSDB(url string, statusCode int, responseTime float64) error {
	appender := s.db.Appender(context.Background())
	defer appender.Rollback()

	labelSet := labels.Labels{
		{Name: "__name__", Value: "http_status"},
		{Name: "url", Value: url},
	}

	_, err := appender.Append(0, labelSet, time.Now().UnixNano()/int64(time.Millisecond), float64(statusCode))
	if err != nil {
		return fmt.Errorf("error appending sample: %v", err)
	}

	responseTimeLabelSet := labels.Labels{
		{Name: "__name__", Value: "http_response_time"},
		{Name: "url", Value: url},
	}
	_, err = appender.Append(0, responseTimeLabelSet, time.Now().UnixNano()/int64(time.Millisecond), responseTime)
	if err != nil {
		return fmt.Errorf("error appending response time sample: %v", err)
	}

	if err := appender.Commit(); err != nil {
		return fmt.Errorf("error committing sample: %v", err)
	}

	return nil
}

func (s *StatusMonitor) checkStatusWithRetry(target Target) (int, float64, error) {
	var lastErr error
	for attempt := 0; attempt < s.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(s.config.RetryDelay)
		}

		start := time.Now()
		resp, err := s.httpClient.Get(target.URL)
		responseTime := time.Since(start).Seconds() * 1000
		if err == nil {
			defer resp.Body.Close()
			return resp.StatusCode, responseTime, nil
		}
		lastErr = err
		s.logger.Warnf("Attempt %d failed for %s: %v", attempt+1, target.URL, err)
	}
	return 0, 0, lastErr
}

func (s *StatusMonitor) checkStatus(target Target) {
	for {
		statusCode, responseTime, err := s.checkStatusWithRetry(target)
		if err != nil {
			s.logger.Errorf("Error checking %s: %v", target.URL, err)
			if err := s.writeToTSDB(target.URL, 0, 0); err != nil {
				s.logger.Errorf("Error writing to TSDB: %v", err)
			}
		} else {
			s.logger.Infof("Status for %s: %d", target.URL, statusCode)
			if err := s.writeToTSDB(target.URL, statusCode, responseTime); err != nil {
				s.logger.Errorf("Error writing to TSDB: %v", err)
			}
		}
		time.Sleep(target.Interval)
	}
}

func (s *StatusMonitor) parseTimeRange(r *http.Request) (TimeRange, error) {
	now := time.Now()
	defaultRange := TimeRange{
		Start: now.Add(-1 * time.Hour),
		End:   now,
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	duration := r.URL.Query().Get("duration")

	if duration != "" {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return defaultRange, err
		}
		return TimeRange{
			Start: now.Add(-d),
			End:   now,
		}, nil
	}

	if start != "" && end != "" {
		startTime, err := time.Parse(time.RFC3339, start)
		if err != nil {
			return defaultRange, err
		}
		endTime, err := time.Parse(time.RFC3339, end)
		if err != nil {
			return defaultRange, err
		}
		return TimeRange{Start: startTime, End: endTime}, nil
	}

	return defaultRange, nil
}

func (s *StatusMonitor) QueryStatus(url string, timeRange TimeRange) ([]QueryResult, error) {
	querier, err := s.db.Querier(
		timeRange.Start.UnixMilli(),
		timeRange.End.UnixMilli(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating querier: %v", err)
	}
	defer querier.Close()

	statusMatchers := []*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "__name__", "http_status"),
		labels.MustNewMatcher(labels.MatchEqual, "url", url),
	}

	responseTimeMatchers := []*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "__name__", "http_response_time"),
		labels.MustNewMatcher(labels.MatchEqual, "url", url),
	}
	statusSeries := querier.Select(context.Background(), false, nil, statusMatchers...)
	// Query for response time series
	responseTimeSeries := querier.Select(context.Background(), false, nil, responseTimeMatchers...)

	// Collect status results
	statusResults := make(map[int64]int) // timestamp -> status code
	for statusSeries.Next() {
		iter := statusSeries.At().Iterator(nil)
		for iter.Next() == chunkenc.ValFloat {
			ts, val := iter.At()
			statusResults[ts] = int(val)
		}
	}

	// Collect response time results and combine with status results
	var results []QueryResult
	for responseTimeSeries.Next() {
		iter := responseTimeSeries.At().Iterator(nil)
		for iter.Next() == chunkenc.ValFloat {
			ts, val := iter.At()
			responseTime := float64(val)
			status, exists := statusResults[ts]
			if exists {
				results = append(results, QueryResult{
					URL:          url,
					Timestamp:    time.Unix(0, ts*int64(time.Millisecond)),
					Status:       status,
					ResponseTime: responseTime,
				})
			}
		}
	}

	return results, nil
}

func (s *StatusMonitor) setupRouter() *chi.Mux {
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

func (s *StatusMonitor) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *StatusMonitor) jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIError{Error: message})
}

func (s *StatusMonitor) Start() {
	// Start monitoring each target
	for _, target := range s.targets {
		go s.checkStatus(target)
	}

	// Setup and start HTTP server
	router := s.setupRouter()
	s.logger.Infof("Starting server on port %s...", s.config.ServerPort)
	if err := http.ListenAndServe(":"+s.config.ServerPort, router); err != nil {
		s.logger.Fatalf("Server error: %v", err)
	}
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	monitor, err := NewStatusMonitor(config)
	if err != nil {
		log.Fatalf("Error creating monitor: %v", err)
	}

	monitor.Start()
}
