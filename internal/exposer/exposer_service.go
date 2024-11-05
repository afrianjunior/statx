package exposer

import (
	"context"
	"fmt"
	"time"

	"github.com/afrianjunior/statx/internal/pkg"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
)

type exposerService struct {
	tsdb *tsdb.DB
}

type ExposerService interface {
	QueryUpTimeStatus(ctx context.Context, url string, timeRange pkg.TimeRange) ([]pkg.QueryResult, error)
}

func NewExposerService(db *tsdb.DB) ExposerService {
	return &exposerService{
		db,
	}
}

func (s *exposerService) QueryUpTimeStatus(ctx context.Context, url string, timeRange pkg.TimeRange) ([]pkg.QueryResult, error) {
	querier, err := s.tsdb.Querier(
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
	statusSeries := querier.Select(ctx, false, nil, statusMatchers...)
	responseTimeSeries := querier.Select(ctx, false, nil, responseTimeMatchers...)

	statusResults := make(map[int64]int)
	for statusSeries.Next() {
		iter := statusSeries.At().Iterator(nil)
		for iter.Next() == chunkenc.ValFloat {
			ts, val := iter.At()
			statusResults[ts] = int(val)
		}
	}

	var results []pkg.QueryResult
	for responseTimeSeries.Next() {
		iter := responseTimeSeries.At().Iterator(nil)
		for iter.Next() == chunkenc.ValFloat {
			ts, val := iter.At()
			responseTime := float64(val)
			status, exists := statusResults[ts]
			if exists {
				results = append(results, pkg.QueryResult{
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
