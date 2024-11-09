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

type AccountDTO struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	LastLogin time.Time `json:"last_login"`
}

type ConfigMonitorDTO struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Method        string `json:"method"`
	Name          string `json:"name"`
	URL           string `json:"url"`
	Interval      int    `json:"interval"`
	Icon          string `json:"icon"`
	Color         string `json:"color"`
	MaxRetry      int    `json:"max_retry"`
	RetryInterval int    `json:"retry_interval"`
	CallMethod    string `json:"call_method"`
	CallEncoding  string `json:"call_encoding"`
	CallBody      string `json:"call_body"`
	CallHeaders   string `json:"call_headers"`
}
