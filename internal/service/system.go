package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrSystemHealthCheck = errors.New("system health check failed") // system health check failed
	ErrNoVersionInfo     = errors.New("no version info provided")   // no version info provided
	ErrNoResources       = errors.New("no resources provided")      // no resources provided
)

// Pingable defines the interface for a driver that can be pinged.
type Pingable interface {
	Ping(ctx context.Context) error
}

// SystemService serves the business logic of interacting with the server
// through drivers.
type SystemService interface {
	// GetHeartbeat returns a heartbeat response.
	GetHeartbeat(ctx context.Context) error
	// GetHealth returns a health response.
	GetHealth(ctx context.Context) (map[string]model.HealthStatus, error)
	// GetVersion returns a version response.
	GetVersion(ctx context.Context) (*model.VersionInfo, error)
}

// systemService is the concrete implementation of SystemService.
type systemService struct {
	*baseService
	versionInfo *model.VersionInfo
	resources   map[string]Pingable
}

func (s *systemService) checkStatus(ctx context.Context, label string, resource Pingable, response map[string]model.HealthStatus, errCh chan error, wg *sync.WaitGroup, lock *sync.RWMutex) {
	span := trace.SpanFromContext(ctx)
	lock.Lock()

	defer wg.Done()
	defer lock.Unlock()

	span.AddEvent(fmt.Sprintf("Check %s health", label))
	if err := resource.Ping(ctx); err != nil {
		response[label] = model.HealthStatusUnhealthy
		errCh <- errors.Join(ErrSystemHealthCheck, err)
		return
	}

	response[label] = model.HealthStatusHealthy
}

func (s *systemService) GetHeartbeat(ctx context.Context) error {
	_, span := s.tracer.Start(ctx, "core.baseService.system/GetHeartbeat")
	defer span.End()

	return nil
}

func (s *systemService) GetHealth(ctx context.Context) (map[string]model.HealthStatus, error) {
	ctx, span := s.tracer.Start(ctx, "core.baseService.system/GetHealth")
	defer span.End()

	var wg sync.WaitGroup
	var lock sync.RWMutex

	errCh := make(chan error, 1)
	response := make(map[string]model.HealthStatus)

	wg.Add(len(s.resources))
	for name, resource := range s.resources {
		response[name] = model.HealthStatusUnknown
		go s.checkStatus(ctx, name, resource, response, errCh, &wg, &lock)
	}

	wg.Wait()
	close(errCh)

	return response, <-errCh
}

func (s *systemService) GetVersion(ctx context.Context) (*model.VersionInfo, error) {
	_, span := s.tracer.Start(ctx, "core.baseService.system/GetVersion")
	defer span.End()

	return &model.VersionInfo{
		Version:   s.versionInfo.Version,
		Commit:    s.versionInfo.Commit,
		Date:      s.versionInfo.Date,
		GoVersion: s.versionInfo.GoVersion,
	}, nil
}

// NewSystemService creates a new SystemService.
func NewSystemService(resources map[string]Pingable, version *model.VersionInfo, opts ...Option) (SystemService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &systemService{
		baseService: s,
		versionInfo: version,
		resources:   resources,
	}

	if svc.versionInfo == nil {
		return nil, ErrNoVersionInfo
	}

	if svc.resources == nil || len(svc.resources) == 0 {
		return nil, ErrNoResources
	}

	return svc, nil
}
