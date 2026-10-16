package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	oapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/opcotech/elemo/internal/deployment"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/entitlement/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// SystemController is a controller for system endpoints.
type SystemController interface {
	V1SystemHealth(ctx context.Context, request api.V1SystemHealthRequestObject) (api.V1SystemHealthResponseObject, error)
	V1SystemHeartbeat(ctx context.Context, request api.V1SystemHeartbeatRequestObject) (api.V1SystemHeartbeatResponseObject, error)
	V1SystemVersion(ctx context.Context, request api.V1SystemVersionRequestObject) (api.V1SystemVersionResponseObject, error)
	V1SystemEntitlements(ctx context.Context, request api.V1SystemEntitlementsRequestObject) (api.V1SystemEntitlementsResponseObject, error)
}

type systemController struct {
	*baseController
	systemService      service.SystemService
	entitlementService service.EntitlementService
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

func (c *systemController) V1SystemEntitlements(ctx context.Context, _ api.V1SystemEntitlementsRequestObject) (api.V1SystemEntitlementsResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetSystemEntitlements")
	defer span.End()

	status, err := c.entitlementService.Get(ctx)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1SystemEntitlements403JSONResponse{N403JSONResponse: permissionDenied}, nil
		default:
			c.logger.Error(ctx, "failed to get entitlement status", log.WithError(err))
			return api.V1SystemEntitlements500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: http.StatusText(http.StatusInternalServerError),
			}}, nil
		}
	}

	dto, err := entitlementsToDTO(status)
	if err != nil {
		c.logger.Error(ctx, "failed to map entitlement status", log.WithError(err))
		return api.V1SystemEntitlements500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: http.StatusText(http.StatusInternalServerError),
		}}, nil
	}

	return api.V1SystemEntitlements200JSONResponse(*dto), nil
}

func NewSystemController(
	systemService service.SystemService,
	entitlementService service.EntitlementService,
	opts ...ControllerOption,
) (SystemController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	if systemService == nil {
		return nil, ErrNoSystemService
	}

	if entitlementService == nil {
		return nil, entitlement.ErrNoEntitlementService
	}

	return &systemController{
		baseController:     c,
		systemService:      systemService,
		entitlementService: entitlementService,
	}, nil
}

func healthStatusToDTO(status map[model.HealthCheckComponent]model.HealthStatus) *api.SystemHealth {
	return &api.SystemHealth{
		CacheDatabase:      api.SystemHealthCacheDatabase(status[model.HealthCheckComponentCacheDB].String()),
		GraphDatabase:      api.SystemHealthGraphDatabase(status[model.HealthCheckComponentGraphDB].String()),
		RelationalDatabase: api.SystemHealthRelationalDatabase(status[model.HealthCheckComponentRelationalDB].String()),
		MessageQueue:       api.SystemHealthMessageQueue(status[model.HealthCheckComponentMessageQueue].String()),
		Search:             api.SystemHealthSearch(status[model.HealthCheckComponentSearch].String()),
	}
}

func versionInfoToDTO(version *model.VersionInfo) *api.SystemVersion {
	date, _ := time.Parse(time.RFC3339, version.Date)

	return &api.SystemVersion{
		Version:   version.Version,
		Commit:    version.Commit,
		Date:      date,
		GoVersion: version.GoVersion,
	}
}

func entitlementsToDTO(status entitlement.Status) (*api.SystemEntitlements, error) {
	out := &api.SystemEntitlements{
		DeploymentMode: api.SystemEntitlementsDeploymentModeSelfHosted,
	}
	if status.DeploymentMode == deployment.ModeAirGap {
		out.DeploymentMode = api.SystemEntitlementsDeploymentModeAirgap
	}
	if status.AirGap == nil {
		return out, nil
	}

	air := status.AirGap
	dto := &api.SystemAirGapEntitlement{
		State:       airGapStateToDTO(air.State),
		SeatsActive: air.SeatsActive,
	}
	parsedInstallationID, err := uuid.Parse(air.InstallationID)
	if err != nil {
		return nil, fmt.Errorf("invalid installation ID: %w", err)
	}
	dto.InstallationId = oapiTypes.UUID(parsedInstallationID)
	if air.Reason != "" {
		reason := air.Reason
		dto.Reason = &reason
	}
	if air.Customer != "" {
		customer := air.Customer
		dto.Customer = &customer
	}
	if air.KeyID != "" {
		keyID := air.KeyID
		dto.KeyId = &keyID
	}
	if air.LicenseID != "" {
		if parsed, err := uuid.Parse(air.LicenseID); err == nil {
			id := oapiTypes.UUID(parsed)
			dto.LicenseId = &id
		}
	}
	if air.SeatsLicensed > 0 {
		seats := air.SeatsLicensed
		dto.SeatsLicensed = &seats
	}
	dto.IssuedAt = air.IssuedAt
	dto.NotBefore = air.NotBefore
	dto.ExpiresAt = air.ExpiresAt
	dto.GraceEndsAt = air.GraceEndsAt
	out.Airgap = dto
	return out, nil
}

func airGapStateToDTO(state license.State) api.SystemAirGapEntitlementState {
	switch state {
	case license.StateValid:
		return api.SystemAirGapEntitlementStateValid
	case license.StateGrace:
		return api.SystemAirGapEntitlementStateGrace
	case license.StateExpired:
		return api.SystemAirGapEntitlementStateExpired
	case license.StateNotYetValid:
		return api.SystemAirGapEntitlementStateNotYetValid
	case license.StateMissing:
		return api.SystemAirGapEntitlementStateMissing
	default:
		return api.SystemAirGapEntitlementStateInvalid
	}
}
