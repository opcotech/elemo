package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/pkg/safepath"
	"github.com/opcotech/elemo/internal/repository"
)

const (
	staticFileRoot = "/"
)

// StaticFileService serves the business logic of creating and retrieving files.
//
//go:generate go tool mockgen -destination=mock/mock_staticfile_gen.go -package=mocksvc . StaticFileService
type StaticFileService interface {
	// Create puts a new file in the static storage for the given path, reading
	// its data from the reader. It returns an error if the operation failed.
	Create(ctx context.Context, path string, data []byte) error
	// Get retrieves an object and writes its data to the designated location.
	// It returns an error if the operation failed.
	Get(ctx context.Context, path string) ([]byte, error)
	// Update replaces the file at the given path with the new data. It returns
	// an error if the operation failed.
	Update(ctx context.Context, path string, data []byte) error
	// Delete removes a file from the static storage, and returns an error if
	// the operation failed.
	Delete(ctx context.Context, path string) error
}

type staticFileService struct {
	runtime
	licenseService LicenseService
	staticFileRepo repository.StaticFileRepository
}

func (s *staticFileService) Create(ctx context.Context, path string, data []byte) error {
	ctx, span := s.tracer.Start(ctx, "service.staticFileService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrStaticFileCreate, license.ErrLicenseExpired)
	}

	safePath, err := safepath.Normalize(staticFileRoot, path)
	if err != nil {
		return errors.Join(ErrStaticFileCreate, ErrStaticFileInvalidPath)
	}

	if err := s.staticFileRepo.Create(ctx, safePath, data); err != nil {
		return errors.Join(ErrStaticFileCreate, err)
	}

	return nil
}

func (s *staticFileService) Get(ctx context.Context, path string) ([]byte, error) {
	ctx, span := s.tracer.Start(ctx, "service.staticFileService/Get")
	defer span.End()

	safePath, err := safepath.Normalize(staticFileRoot, path)
	if err != nil {
		return nil, errors.Join(ErrStaticFileGet, ErrStaticFileInvalidPath)
	}

	data, err := s.staticFileRepo.Get(ctx, safePath)
	if err != nil {
		return nil, errors.Join(ErrStaticFileGet, err)
	}

	return data, nil
}

func (s *staticFileService) Update(ctx context.Context, path string, data []byte) error {
	ctx, span := s.tracer.Start(ctx, "service.staticFileService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrStaticFileUpdate, license.ErrLicenseExpired)
	}

	safePath, err := safepath.Normalize(staticFileRoot, path)
	if err != nil {
		return errors.Join(ErrStaticFileUpdate, ErrStaticFileInvalidPath)
	}

	if err := s.staticFileRepo.Update(ctx, safePath, data); err != nil {
		return errors.Join(ErrStaticFileUpdate, err)
	}

	return nil
}

func (s *staticFileService) Delete(ctx context.Context, path string) error {
	ctx, span := s.tracer.Start(ctx, "service.staticFileService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrStaticFileDelete, license.ErrLicenseExpired)
	}

	safePath, err := safepath.Normalize(staticFileRoot, path)
	if err != nil {
		return errors.Join(ErrStaticFileDelete, ErrStaticFileInvalidPath)
	}

	if err := s.staticFileRepo.Delete(ctx, safePath); err != nil {
		return errors.Join(ErrStaticFileDelete, err)
	}

	return nil
}

// NewStaticFileService returns a new instance of the StaticFileService interface.
func NewStaticFileService(repo repository.StaticFileRepository, licenseService LicenseService, opts ...Option) (StaticFileService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}

	svc := &staticFileService{
		runtime:        rt,
		licenseService: licenseService,
		staticFileRepo: repo,
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	if svc.staticFileRepo == nil {
		return nil, ErrNoStaticFileRepository
	}

	return svc, nil
}
