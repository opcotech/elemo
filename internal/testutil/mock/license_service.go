package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	elemoLicense "github.com/opcotech/elemo/internal/license"
)

type LicenseService struct {
	mock.Mock
}

func (l *LicenseService) Expired(ctx context.Context) bool {
	args := l.Called(ctx)
	return args.Bool(0)
}

func (l *LicenseService) HasFeature(ctx context.Context, feature elemoLicense.Feature) bool {
	args := l.Called(ctx, feature)
	return args.Bool(0)
}

func (l *LicenseService) WithinQuota(ctx context.Context, name elemoLicense.Quota, current int) bool {
	args := l.Called(ctx, name, current)
	return args.Bool(0)
}
