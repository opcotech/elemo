package http

import (
	"context"
	"errors"
	"time"

	oapiTypes "github.com/deepmap/oapi-codegen/pkg/types"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// SystemController is a controller for system endpoints.
type SystemController interface {
	V1SystemHealth(ctx context.Context, request api.V1SystemHealthRequestObject) (api.V1SystemHealthResponseObject, error)
	V1SystemHeartbeat(ctx context.Context, request api.V1SystemHeartbeatRequestObject) (api.V1SystemHeartbeatResponseObject, error)
	V1SystemVersion(ctx context.Context, request api.V1SystemVersionRequestObject) (api.V1SystemVersionResponseObject, error)
	V1SystemLicense(ctx context.Context, request api.V1SystemLicenseRequestObject) (api.V1SystemLicenseResponseObject, error)
}

// systemController is the concrete implementation of SystemController.
type systemController struct {
	*baseController
}

func (c *systemController) V1SystemHealth(ctx context.Context, _ api.V1SystemHealthRequestObject) (api.V1SystemHealthResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemHealth")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	health, _ := c.systemService.GetHealth(ctx)
	return api.V1SystemHealth200JSONResponse(*healthStatusToDTO(health)), nil
}

func (c *systemController) V1SystemHeartbeat(ctx context.Context, _ api.V1SystemHeartbeatRequestObject) (api.V1SystemHeartbeatResponseObject, error) {
	_, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemHeartbeat")
	defer span.End()

	return api.V1SystemHeartbeat200TextResponse("OK"), nil
}

func (c *systemController) V1SystemVersion(ctx context.Context, _ api.V1SystemVersionRequestObject) (api.V1SystemVersionResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemVersion")
	defer span.End()

	versionInfo := c.systemService.GetVersion(ctx)

	return api.V1SystemVersion200JSONResponse(*versionInfoToDTO(versionInfo)), nil
}

func (c *systemController) V1SystemLicense(ctx context.Context, _ api.V1SystemLicenseRequestObject) (api.V1SystemLicenseResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemLicense")
	defer span.End()

	l, err := c.licenseService.GetLicense(ctx)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1SystemLicense403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1SystemLicenseResponseObject(nil), err
	}

	return api.V1SystemLicense200JSONResponse(*licenseToDTO(&l)), nil
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

func healthStatusToDTO(status map[model.HealthCheckComponent]model.HealthStatus) *api.SystemHealth {
	return &api.SystemHealth{
		CacheDatabase:      api.SystemHealthCacheDatabase(status[model.HealthCheckComponentCacheDB].String()),
		GraphDatabase:      api.SystemHealthGraphDatabase(status[model.HealthCheckComponentGraphDB].String()),
		RelationalDatabase: api.SystemHealthRelationalDatabase(status[model.HealthCheckComponentRelationalDB].String()),
		License:            api.SystemHealthLicense(status[model.HealthCheckComponentLicense].String()),
		MessageQueue:       api.SystemHealthMessageQueue(status[model.HealthCheckComponentMessageQueue].String()),
	}
}

func versionInfoToDTO(version *model.VersionInfo) *api.SystemVersion {
	// The date is set by ldflags, so it's always in the same format.
	date, _ := time.Parse(time.RFC3339, version.Date)

	return &api.SystemVersion{
		Version:   version.Version,
		Commit:    version.Commit,
		Date:      date,
		GoVersion: version.GoVersion,
	}
}

func licenseToDTO(l *license.License) *api.SystemLicense {
	type licenseQuota = struct {
		Documents     int `json:"documents"`
		Namespaces    int `json:"namespaces"`
		Organizations int `json:"organizations"`
		Projects      int `json:"projects"`
		Roles         int `json:"roles"`
		Users         int `json:"users"`
	}

	systemLicense := &api.SystemLicense{
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
		systemLicense.Features = append(systemLicense.Features, api.SystemLicenseFeatures(feature))
	}

	return systemLicense
}
