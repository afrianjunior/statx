package pkg

import "time"

type Config struct {
	Targets          []Target      `json:"targets"`
	StoragePath      string        `json:"storage_path"`
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
