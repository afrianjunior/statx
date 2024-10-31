package pkg

import "time"

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type QueryResult struct {
	URL          string    `json:"url"`
	Timestamp    time.Time `json:"timestamp"`
	Status       int       `json:"status"`
	ResponseTime float64   `json:"response_time"`
}

type Target struct {
	URL      string        `json:"url"`
	Interval time.Duration `json:"interval"`
}
