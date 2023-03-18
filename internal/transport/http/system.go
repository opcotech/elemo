package http

import (
	"context"
	"errors"

	oapiTypes "github.com/deepmap/oapi-codegen/pkg/types"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/gen"
)

// SystemController is a controller for system endpoints.
type SystemController interface {
	GetSystemHealth(ctx context.Context, request gen.GetSystemHealthRequestObject) (gen.GetSystemHealthResponseObject, error)
	GetSystemHeartbeat(ctx context.Context, request gen.GetSystemHeartbeatRequestObject) (gen.GetSystemHeartbeatResponseObject, error)
	GetSystemVersion(ctx context.Context, request gen.GetSystemVersionRequestObject) (gen.GetSystemVersionResponseObject, error)
	GetSystemLicense(ctx context.Context, request gen.GetSystemLicenseRequestObject) (gen.GetSystemLicenseResponseObject, error)
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
	_, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemHeartbeat")
	defer span.End()

	return gen.GetSystemHeartbeat200TextResponse(gen.SystemHeartbeatOK), nil
}

func (c *systemController) GetSystemVersion(ctx context.Context, _ gen.GetSystemVersionRequestObject) (gen.GetSystemVersionResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemVersion")
	defer span.End()

	versionInfo := c.systemService.GetVersion(ctx)

	return gen.GetSystemVersion200JSONResponse(*versionInfoToDTO(versionInfo)), nil
}

func (c *systemController) GetSystemLicense(ctx context.Context, _ gen.GetSystemLicenseRequestObject) (gen.GetSystemLicenseResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemLicense")
	defer span.End()

	l, err := c.licenseService.GetLicense(ctx)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return gen.GetSystemLicense401JSONResponse{N401JSONResponse: permissionDenied}, nil
		}
		return gen.GetSystemLicenseResponseObject(nil), err
	}

	return gen.GetSystemLicense200JSONResponse(*licenseToDTO(&l)), nil
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
		License:            gen.SystemHealthLicense(status[model.HealthCheckComponentLicense].String()),
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

func licenseToDTO(l *license.License) *gen.SystemLicense {
	type licenseQuota = struct {
		Documents     int `json:"documents"`
		Namespaces    int `json:"namespaces"`
		Organizations int `json:"organizations"`
		Projects      int `json:"projects"`
		Roles         int `json:"roles"`
		Users         int `json:"users"`
	}

	systemLicense := &gen.SystemLicense{
		Id:           l.ID.String(),
		Organization: l.Organization,
		Email:        oapiTypes.Email(l.Email),
		Quotas: licenseQuota{
			Documents:     int(l.Quotas[license.QuotaDocuments]),
			Namespaces:    int(l.Quotas[license.QuotaNamespaces]),
			Organizations: int(l.Quotas[license.QuotaOrganizations]),
			Projects:      int(l.Quotas[license.QuotaProjects]),
			Roles:         int(l.Quotas[license.QuotaRoles]),
			Users:         int(l.Quotas[license.QuotaUsers]),
		},
		ExpiresAt: l.ExpiresAt,
	}

	for _, feature := range l.Features {
		systemLicense.Features = append(systemLicense.Features, gen.SystemLicenseFeatures(feature))
	}

	return systemLicense
}
