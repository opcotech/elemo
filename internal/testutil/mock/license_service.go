package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	elemoLicense "github.com/opcotech/elemo/internal/license"
)

type LicenseService struct {
	mock.Mock
}

func (l *LicenseService) Expired(ctx context.Context) (bool, error) {
	args := l.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (l *LicenseService) HasFeature(ctx context.Context, feature elemoLicense.Feature) (bool, error) {
	args := l.Called(ctx, feature)
	return args.Bool(0), args.Error(1)
}

func (l *LicenseService) WithinThreshold(ctx context.Context, name elemoLicense.Quota) (bool, error) {
	args := l.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (l *LicenseService) Ping(ctx context.Context) error {
	args := l.Called(ctx)
	return args.Error(0)
}
