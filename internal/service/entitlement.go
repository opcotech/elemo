package service

import (
	"context"

	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/model"
)

// EntitlementService exposes administrator entitlement status.
//
//go:generate go tool mockgen -destination=mock/mock_entitlement_gen.go -package=mocksvc . EntitlementService
type EntitlementService interface {
	Get(ctx context.Context) (entitlement.Status, error)
}

type entitlementService struct {
	runtime
	policy            entitlement.Reporter
	permissionService PermissionService
}

func (s *entitlementService) Get(ctx context.Context) (entitlement.Status, error) {
	ctx, span := s.tracer.Start(ctx, "service.entitlementService/Get")
	defer span.End()

	if err := requireAction(ctx, s.permissionService, model.InstallationID(), model.ActionOrganizationCreate); err != nil {
		return entitlement.Status{}, err
	}

	status, err := s.policy.Status(ctx)
	if err != nil {
		return entitlement.Status{}, err
	}
	return status, nil
}

func NewEntitlementService(policy entitlement.Reporter, permissionService PermissionService, opts ...Option) (EntitlementService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, entitlement.ErrNoSeatPolicy
	}
	if permissionService == nil {
		return nil, ErrNoPermissionService
	}
	return &entitlementService{
		runtime:           rt,
		policy:            policy,
		permissionService: permissionService,
	}, nil
}
