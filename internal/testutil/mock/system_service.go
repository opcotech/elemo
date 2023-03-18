package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type SystemService struct {
	mock.Mock
}

func (m *SystemService) GetHeartbeat(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *SystemService) GetHealth(ctx context.Context) (map[model.HealthCheckComponent]model.HealthStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[model.HealthCheckComponent]model.HealthStatus), args.Error(1)
}

func (m *SystemService) GetVersion(ctx context.Context) *model.VersionInfo {
	args := m.Called(ctx)
	return args.Get(0).(*model.VersionInfo)
}

type PingableResource struct {
	mock.Mock
}

func (m *PingableResource) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
