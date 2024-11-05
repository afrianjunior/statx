package recorder

import (
	"context"
	"net/http"
	"time"

	"github.com/afrianjunior/statx/internal/pkg"
	"go.uber.org/zap"
)

type genericJob struct {
	recorderService RecorderService
	targets         []pkg.Target
	httpClient      *http.Client
	logger          *zap.SugaredLogger
}

type GenericJob interface {
	Start(ctx context.Context)
}

func NewGenericJob(
	recorderService RecorderService,
	targets []pkg.Target,
	httpClient *http.Client,
	logger *zap.SugaredLogger,
) GenericJob {
	return &genericJob{
		recorderService: recorderService,
		httpClient:      httpClient,
		logger:          logger,
		targets:         targets,
	}
}

func (s *genericJob) Start(ctx context.Context) {
	for _, target := range s.targets {
		go s.checkStatus(ctx, target)
	}
}

func (s *genericJob) checkStatus(ctx context.Context, target pkg.Target) {
	for {
		statusCode, responseTime, err := s.recorderService.CheckUptimeWithRetry(target)
		if err != nil {
			s.logger.Errorf("Error checking %s: %v", target.URL, err)
			if err := s.recorderService.WriteUpTimeRecord(ctx, target.URL, 0, 0); err != nil {
				s.logger.Errorf("Error writing to TSDB: %v", err)
			}
		} else {
			s.logger.Infof("Status for %s: %d", target.URL, statusCode)
			if err := s.recorderService.WriteUpTimeRecord(ctx, target.URL, statusCode, responseTime); err != nil {
				s.logger.Errorf("Error writing to TSDB: %v", err)
			}
		}
		time.Sleep(target.Interval)
	}
}
