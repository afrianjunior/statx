package config_monitor

import (
	"context"

	"github.com/afrianjunior/statx/internal/pkg"
)

type configMonitorService struct {
	configMonitorRepository ConfigMonitorRepository
}

type ConfigMonitorService interface {
	MutateConfigMonitor(ctx context.Context, payload *pkg.ConfigMonitorDTO) (string, error)
	GetListConfigMonitors(ctx context.Context, limit, offset int) ([]*pkg.ConfigMonitorDTO, int, error)
}

func NewConfigService(
	configMonitorRepository ConfigMonitorRepository,
) ConfigMonitorService {
	return &configMonitorService{
		configMonitorRepository: configMonitorRepository,
	}
}

func (s *configMonitorService) MutateConfigMonitor(ctx context.Context, payload *pkg.ConfigMonitorDTO) (string, error) {
	return s.configMonitorRepository.Insert(ctx, payload)
}

func (s *configMonitorService) GetListConfigMonitors(ctx context.Context, limit, offset int) ([]*pkg.ConfigMonitorDTO, int, error) {
	list, total, err := s.configMonitorRepository.List(ctx, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	return list, total, err
}
