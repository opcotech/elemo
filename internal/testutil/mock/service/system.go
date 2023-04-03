package service

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/model"
)

type MockSystemService struct {
	mock.Mock
}

func (m *MockSystemService) GetHeartbeat(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSystemService) GetHealth(ctx context.Context) (map[model.HealthCheckComponent]model.HealthStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[model.HealthCheckComponent]model.HealthStatus), args.Error(1)
}

func (m *MockSystemService) GetVersion(ctx context.Context) *model.VersionInfo {
	args := m.Called(ctx)
	return args.Get(0).(*model.VersionInfo)
}

type MockPingable struct {
	mock.Mock
}

func (m *MockPingable) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
