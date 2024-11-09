package config_monitor

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/afrianjunior/statx/internal/pkg"
	_ "github.com/glebarez/go-sqlite"
)

type configMonitorRepository struct {
	db *sql.DB
}

type ConfigMonitorRepository interface {
	Insert(ctx context.Context, config *pkg.ConfigMonitorDTO) (string, error)
	GetByID(ctx context.Context, id string) (*pkg.ConfigMonitorDTO, error)
	List(ctx context.Context, limit, offset int) ([]*pkg.ConfigMonitorDTO, int, error)
}

func NewConfigMonitorRepository(
	db *sql.DB,
) ConfigMonitorRepository {
	return &configMonitorRepository{
		db,
	}
}

func (r *configMonitorRepository) Insert(ctx context.Context, config *pkg.ConfigMonitorDTO) (string, error) {
	query := `
		INSERT INTO config_monitor (
			type, method, name, url, interval, icon, color, 
			max_retry, retry_interval, call_method, call_encoding, 
			call_body, call_headers
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		config.Type,
		config.Method,
		config.Name,
		config.URL,
		config.Interval,
		config.Icon,
		config.Color,
		config.MaxRetry,
		config.RetryInterval,
		config.CallMethod,
		config.CallEncoding,
		config.CallBody,
		config.CallHeaders,
	)

	if err != nil {
		return "", fmt.Errorf("error inserting config monitor: %w", err)
	}

	// Get the last inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("error getting last insert ID: %w", err)
	}

	// Fetch the generated UUID
	var uuid string
	err = r.db.QueryRowContext(ctx, "SELECT id FROM config_monitor WHERE rowid = ?", id).Scan(&uuid)
	if err != nil {
		return "", fmt.Errorf("error fetching generated UUID: %w", err)
	}

	return uuid, nil
}

func (r *configMonitorRepository) GetByID(ctx context.Context, id string) (*pkg.ConfigMonitorDTO, error) {
	query := `
		SELECT id, type, method, name, url, interval, icon, color, 
			   max_retry, retry_interval, call_method, call_encoding, 
			   call_body, call_headers
		FROM config_monitor
		WHERE id = ?
	`

	var config pkg.ConfigMonitorDTO
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&config.ID,
		&config.Type,
		&config.Method,
		&config.Name,
		&config.URL,
		&config.Interval,
		&config.Icon,
		&config.Color,
		&config.MaxRetry,
		&config.RetryInterval,
		&config.CallMethod,
		&config.CallEncoding,
		&config.CallBody,
		&config.CallHeaders,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("error fetching config monitor: %w", err)
		}
		return nil, fmt.Errorf("error fetching config monitor: %w", err)
	}

	return &config, nil
}

func (r *configMonitorRepository) List(ctx context.Context, limit, offset int) ([]*pkg.ConfigMonitorDTO, int, error) {
	query := `
		SELECT id, type, method, name, url, interval, icon, color, 
			   max_retry, retry_interval, call_method, call_encoding, 
			   call_body, call_headers
		FROM config_monitor
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying config monitors: %w", err)
	}
	defer rows.Close()

	var configs []*pkg.ConfigMonitorDTO
	for rows.Next() {
		var config pkg.ConfigMonitorDTO
		err := rows.Scan(
			&config.ID,
			&config.Type,
			&config.Method,
			&config.Name,
			&config.URL,
			&config.Interval,
			&config.Icon,
			&config.Color,
			&config.MaxRetry,
			&config.RetryInterval,
			&config.CallMethod,
			&config.CallEncoding,
			&config.CallBody,
			&config.CallHeaders,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning config monitor: %w", err)
		}
		configs = append(configs, &config)
	}

	// Get total count
	var total int
	err = r.db.QueryRow("SELECT COUNT(*) FROM config_monitor").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting total count: %w", err)
	}

	return configs, total, nil
}
