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

	"github.com/afrianjunior/statx/cmd"
	"github.com/afrianjunior/statx/internal/pkg"
	"github.com/prometheus/prometheus/tsdb"
	"go.uber.org/zap"
)

type App struct {
	targets    []pkg.Target
	httpClient *http.Client
	db         *tsdb.DB
	logger     *zap.SugaredLogger
	config     *pkg.Config
}

func loadConfig(path string) (*pkg.Config, error) {
	defaultConfig := pkg.Config{
		Targets: []pkg.Target{
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

	var config pkg.Config
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

func NewApp(config *pkg.Config) (*App, error) {
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

	return &App{
		targets: config.Targets,
		httpClient: &http.Client{
			Timeout: config.CheckTimeout,
		},
		db:     db,
		logger: logger,
		config: config,
	}, nil
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	app, err := NewApp(config)
	if err != nil {
		log.Fatalf("Error creating monitor: %v", err)
	}

	worker := cmd.NewWorker(
		app.db,
		app.targets,
		app.httpClient,
		app.logger,
		app.config,
	)

	worker.Start(context.Background())

	rest := cmd.NewRest(
		app.httpClient,
		app.db,
		app.logger,
	)

	rest.Start(app.config.ServerPort)
}
