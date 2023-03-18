package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type LicenseRepository struct {
	mock.Mock
}

func (l *LicenseRepository) ActiveUserCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (l *LicenseRepository) ActiveOrganizationCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (l *LicenseRepository) DocumentCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (l *LicenseRepository) NamespaceCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (l *LicenseRepository) ProjectCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (l *LicenseRepository) RoleCount(ctx context.Context) (int, error) {
	args := l.Called(ctx)
	return args.Int(0), args.Error(1)
}
