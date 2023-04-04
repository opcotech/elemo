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
	GetHealth(ctx context.Context) (map[model.HealthCheckComponent]model.HealthStatus, error)
	// GetVersion returns a version response.
	GetVersion(ctx context.Context) *model.VersionInfo
}

// systemService is the concrete implementation of SystemService.
type systemService struct {
	*baseService
	versionInfo *model.VersionInfo
	resources   map[model.HealthCheckComponent]Pingable
}

func (s *systemService) checkStatus(
	ctx context.Context,
	name model.HealthCheckComponent,
	resource Pingable,
	response map[model.HealthCheckComponent]model.HealthStatus,
	errCh chan error,
	wg *sync.WaitGroup,
	lock *sync.RWMutex,
) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("Check %s health", name))

	defer wg.Done()

	status := model.HealthStatusHealthy

	if err := resource.Ping(ctx); err != nil {
		status = model.HealthStatusUnhealthy
		errCh <- errors.Join(ErrSystemHealthCheck, err)
	}

	lock.Lock()
	defer lock.Unlock()

	response[name] = status
}

func (s *systemService) GetHeartbeat(ctx context.Context) error {
	_, span := s.tracer.Start(ctx, "service.systemService/GetHeartbeat")
	defer span.End()

	return nil
}

func (s *systemService) GetHealth(ctx context.Context) (map[model.HealthCheckComponent]model.HealthStatus, error) {
	ctx, span := s.tracer.Start(ctx, "service.systemService/GetHealth")
	defer span.End()

	var wg sync.WaitGroup
	var lock sync.RWMutex

	response := make(map[model.HealthCheckComponent]model.HealthStatus)

	for name := range s.resources {
		response[name] = model.HealthStatusUnknown
	}

	wg.Add(len(s.resources))
	errCh := make(chan error, len(s.resources))

	for name, resource := range s.resources {
		go s.checkStatus(ctx, name, resource, response, errCh, &wg, &lock)
	}

	wg.Wait()
	close(errCh)

	return response, <-errCh
}

func (s *systemService) GetVersion(ctx context.Context) *model.VersionInfo {
	_, span := s.tracer.Start(ctx, "service.systemService/GetVersion")
	defer span.End()

	return &model.VersionInfo{
		Version:   s.versionInfo.Version,
		Commit:    s.versionInfo.Commit,
		Date:      s.versionInfo.Date,
		GoVersion: s.versionInfo.GoVersion,
	}
}

// NewSystemService creates a new SystemService.
func NewSystemService(resources map[model.HealthCheckComponent]Pingable, version *model.VersionInfo, opts ...Option) (SystemService, error) {
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
