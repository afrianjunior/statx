package cmd

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/afrianjunior/statx/internal/pkg"
	"github.com/afrianjunior/statx/internal/recorder"
	"github.com/prometheus/prometheus/tsdb"
	"go.uber.org/zap"
)

type worker struct {
	tsdb       *tsdb.DB
	db         *sql.DB
	targets    []pkg.Target
	httpClient *http.Client
	logger     *zap.SugaredLogger
	config     *pkg.Config
}

type Worker interface {
	Start(ctx context.Context)
}

func NewWorker(
	tsdb *tsdb.DB,
	db *sql.DB,
	targets []pkg.Target,
	httpClient *http.Client,
	logger *zap.SugaredLogger,
	config *pkg.Config,
) Worker {
	return &worker{
		tsdb,
		db,
		targets,
		httpClient,
		logger,
		config,
	}
}

func (s *worker) Start(ctx context.Context) {
	recorderService := recorder.NewRecorderService(s.tsdb, s.db, s.config, s.httpClient, s.logger)

	recoderUptimeJob := recorder.NewUptimeJob(recorderService, s.targets, s.httpClient, s.logger)

	recoderUptimeJob.Start(ctx)
}
