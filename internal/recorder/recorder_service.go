package recorder

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/afrianjunior/statx/internal/pkg"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb"
	"go.uber.org/zap"
)

type recorderService struct {
	db         *tsdb.DB
	config     *pkg.Config
	httpClient *http.Client
	logger     *zap.SugaredLogger
}

type RecorderService interface {
	WriteUpTimeRecord(ctx context.Context, url string, statusCode int, responseTime float64) error
	CheckUptimeWithRetry(target pkg.Target) (int, float64, error)
}

func NewRecorderService(
	db *tsdb.DB,
	config *pkg.Config,
	httpClient *http.Client,
	logger *zap.SugaredLogger,
) RecorderService {
	return &recorderService{
		db,
		config,
		httpClient,
		logger,
	}
}

func (s *recorderService) WriteUpTimeRecord(ctx context.Context, url string, statusCode int, responseTime float64) error {
	appender := s.db.Appender(ctx)
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

func (s *recorderService) CheckUptimeWithRetry(target pkg.Target) (int, float64, error) {
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
