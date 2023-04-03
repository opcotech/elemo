package http

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/transport/http/gen"
)

// SystemController is a controller for system endpoints.
type SystemController interface {
	GetSystemHealth(ctx context.Context, request gen.GetSystemHealthRequestObject) (gen.GetSystemHealthResponseObject, error)
	GetSystemHeartbeat(ctx context.Context, request gen.GetSystemHeartbeatRequestObject) (gen.GetSystemHeartbeatResponseObject, error)
	GetSystemVersion(ctx context.Context, request gen.GetSystemVersionRequestObject) (gen.GetSystemVersionResponseObject, error)
}

// systemController is the concrete implementation of SystemController.
type systemController struct {
	*baseController
}

func (c *systemController) GetSystemHealth(ctx context.Context, _ gen.GetSystemHealthRequestObject) (gen.GetSystemHealthResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemHealth")
	defer span.End()

	health, err := c.systemService.GetHealth(ctx)
	if err != nil {
		return gen.GetSystemHealthResponseObject(nil), err
	}

	return gen.GetSystemHealth200JSONResponse(*healthStatusToDTO(health)), nil
}

func (c *systemController) GetSystemHeartbeat(ctx context.Context, _ gen.GetSystemHeartbeatRequestObject) (gen.GetSystemHeartbeatResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemHeartbeat")
	defer span.End()

	return gen.GetSystemHeartbeat200TextResponse(gen.SystemHeartbeatOK), nil
}

func (c *systemController) GetSystemVersion(ctx context.Context, _ gen.GetSystemVersionRequestObject) (gen.GetSystemVersionResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemVersion")
	defer span.End()

	versionInfo := c.systemService.GetVersion(ctx)

	return gen.GetSystemVersion200JSONResponse(*versionInfoToDTO(versionInfo)), nil
}

// NewSystemController creates a new SystemController.
func NewSystemController(opts ...ControllerOption) (SystemController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &systemController{
		baseController: c,
	}

	if controller.systemService == nil {
		return nil, ErrNoSystemService
	}

	return controller, nil
}

func healthStatusToDTO(status map[model.HealthCheckComponent]model.HealthStatus) *gen.SystemHealth {
	return &gen.SystemHealth{
		GraphDatabase:      gen.SystemHealthGraphDatabase(status[model.HealthCheckComponentGraphDB].String()),
		RelationalDatabase: gen.SystemHealthRelationalDatabase(status[model.HealthCheckComponentRelationalDB].String()),
	}
}

func versionInfoToDTO(version *model.VersionInfo) *gen.SystemVersionInfo {
	return &gen.SystemVersionInfo{
		Version:   version.Version,
		Commit:    version.Commit,
		Date:      version.Date,
		GoVersion: version.GoVersion,
	}
}
