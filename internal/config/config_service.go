package config

import "database/sql"

type configService struct {
	db *sql.DB
}

type ConfigService interface {
}

func NewConfigService(db *sql.DB) ConfigService {
	return &configService{
		db,
	}
}
