package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type LicenseRepositoryOld struct {
	mock.Mock
}

func (l *LicenseRepositoryOld) ActiveUserCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (l *LicenseRepositoryOld) ActiveOrganizationCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (l *LicenseRepositoryOld) DocumentCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (l *LicenseRepositoryOld) NamespaceCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (l *LicenseRepositoryOld) ProjectCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (l *LicenseRepositoryOld) RoleCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}
