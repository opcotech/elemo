package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
)

// Grant is a scoped authorization relationship returned by the service.
type Grant struct {
	ID        model.ID
	Principal model.ID
	Scope     model.ID
	RoleID    *model.ID
	Actions   []model.Action
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// CreateGrantOpts holds the data required to create a grant.
type CreateGrantOpts struct {
	Principal model.ID
	Scope     model.ID
	RoleID    *model.ID
	Actions   []model.Action
}

// Validate reports whether the grant has a principal, a scope, and either a
// role or at least one action.
func (o CreateGrantOpts) Validate() error {
	return repository.CreateGrantOpts{
		Principal: o.Principal,
		Scope:     o.Scope,
		RoleID:    o.RoleID,
		Actions:   o.Actions,
	}.Validate()
}

// PermissionService is the single authorization API for request-time checks,
// listings, and grant administration.
//
//go:generate go tool mockgen -destination=permission_mock_gen.go -package=service -mock_names PermissionService=MockPermissionService . PermissionService
type PermissionService interface {
	// Has reports whether actor may perform action on resource, walking
	// MEMBER_OF (depth 0 or 1) and IN_SCOPE_OF ancestry.
	Has(ctx context.Context, actor, resource model.ID, action model.Action) (bool, error)
	// CtxUserHas is Has for the user ID stored in ctx. It returns false when
	// the user is missing or the check fails.
	CtxUserHas(ctx context.Context, resource model.ID, action model.Action) bool
	// EffectiveActions returns the union of grant and role-bundle actions the
	// actor holds on resource, including inherited scopes.
	EffectiveActions(ctx context.Context, actor, resource model.ID) ([]model.Action, error)
	// CtxUserEffectiveActions is EffectiveActions for the user ID stored in
	// ctx. It returns ErrNoUser when the context has no user ID.
	CtxUserEffectiveActions(ctx context.Context, resource model.ID) ([]model.Action, error)
	// Explain returns a Decision for whether actor may perform action on
	// resource. A deny leaves Principal, Scope, GrantID, and RoleID unset.
	Explain(ctx context.Context, actor, resource model.ID, action model.Action) (*repository.Decision, error)

	// Create records a grant without checking the caller's existing permissions.
	Create(ctx context.Context, opts CreateGrantOpts) (*Grant, error)
	// CtxUserCreate creates a grant for the context user. The caller must hold
	// permission.manage on the scope and every action being granted, including
	// those implied by RoleID.
	CtxUserCreate(ctx context.Context, opts CreateGrantOpts) (*Grant, error)
	// Get returns a grant by ID.
	Get(ctx context.Context, id model.ID) (*Grant, error)
	// ListByPrincipal returns grants issued to principal.
	ListByPrincipal(ctx context.Context, principal model.ID) ([]*Grant, error)
	// ListByScope returns grants whose scope is the given resource.
	ListByScope(ctx context.Context, scope model.ID) ([]*Grant, error)
	// Delete removes a grant by ID.
	Delete(ctx context.Context, id model.ID) error
	// CtxUserDelete deletes a grant if the context user holds permission.manage
	// on the grant's scope.
	CtxUserDelete(ctx context.Context, id model.ID) error

	// LinkInScopeOf creates (child)-[:IN_SCOPE_OF]->(parent). It rejects a link
	// that would introduce a cycle.
	LinkInScopeOf(ctx context.Context, child, parent model.ID) error
	// BootstrapCreator grants actions to creator on resource without an
	// authorization check. Use only when creating a resource for its owner.
	BootstrapCreator(ctx context.Context, creator, resource model.ID, actions []model.Action) error
	// GrantRole creates a grant that binds principal to roleID on scope.
	GrantRole(ctx context.Context, principal, scope model.ID, roleID model.ID) error
	// BumpGeneration increments the per-principal authz generation key used by
	// cached membership paths. Evaluator results are not cached.
	BumpGeneration(ctx context.Context, principal model.ID) error
}

type permissionService struct {
	*baseService
	permissionRepo repository.PermissionRepository
}

func grantFromRepository(g *repository.Grant) *Grant {
	if g == nil {
		return nil
	}
	return &Grant{
		ID:        g.ID,
		Principal: g.Principal,
		Scope:     g.Scope,
		RoleID:    g.RoleID,
		Actions:   g.Actions,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

func grantsFromRepository(grants []*repository.Grant) []*Grant {
	out := make([]*Grant, len(grants))
	for i, g := range grants {
		out[i] = grantFromRepository(g)
	}
	return out
}

func (s *permissionService) Has(ctx context.Context, actor, resource model.ID, action model.Action) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/Has")
	defer span.End()

	allowed, err := s.permissionRepo.Has(ctx, actor, resource, action)
	if err != nil {
		return false, errors.Join(ErrPermissionHasPermission, err)
	}
	return allowed, nil
}

func (s *permissionService) CtxUserHas(ctx context.Context, resource model.ID, action model.Action) bool {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/CtxUserHas")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return false
	}

	allowed, err := s.Has(ctx, userID, resource, action)
	if err != nil && !errors.Is(err, repository.ErrPermissionRead) {
		return false
	}
	return allowed
}

func (s *permissionService) EffectiveActions(ctx context.Context, actor, resource model.ID) ([]model.Action, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/EffectiveActions")
	defer span.End()

	actions, err := s.permissionRepo.EffectiveActions(ctx, actor, resource)
	if err != nil {
		return nil, errors.Join(ErrPermissionGet, err)
	}
	return actions, nil
}

func (s *permissionService) CtxUserEffectiveActions(ctx context.Context, resource model.ID) ([]model.Action, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/CtxUserEffectiveActions")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return nil, ErrNoUser
	}
	return s.EffectiveActions(ctx, userID, resource)
}

func (s *permissionService) Explain(ctx context.Context, actor, resource model.ID, action model.Action) (*repository.Decision, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/Explain")
	defer span.End()

	decision, err := s.permissionRepo.Explain(ctx, actor, resource, action)
	if err != nil {
		return nil, errors.Join(ErrPermissionHasPermission, err)
	}
	return decision, nil
}

func (s *permissionService) Create(ctx context.Context, opts CreateGrantOpts) (*Grant, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/Create")
	defer span.End()

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	grant, err := s.permissionRepo.Create(ctx, repository.CreateGrantOpts{
		Principal: opts.Principal,
		Scope:     opts.Scope,
		RoleID:    opts.RoleID,
		Actions:   opts.Actions,
	})
	if err != nil {
		return nil, errors.Join(ErrPermissionCreate, err)
	}
	_ = s.permissionRepo.BumpGeneration(ctx, opts.Principal)
	return grantFromRepository(grant), nil
}

func (s *permissionService) heldActions(ctx context.Context, actor, scope model.ID) (map[model.Action]struct{}, error) {
	actions, err := s.EffectiveActions(ctx, actor, scope)
	if err != nil {
		return nil, err
	}
	held := make(map[model.Action]struct{}, len(actions))
	for _, action := range actions {
		held[action] = struct{}{}
	}
	return held, nil
}

func (s *permissionService) CtxUserCreate(ctx context.Context, opts CreateGrantOpts) (*Grant, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/CtxUserCreate")
	defer span.End()

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return nil, errors.Join(ErrPermissionCreate, ErrNoUser)
	}

	if !s.CtxUserHas(ctx, opts.Scope, model.ActionPermissionManage) {
		return nil, errors.Join(ErrPermissionCreate, ErrNoPermission)
	}

	held, err := s.heldActions(ctx, userID, opts.Scope)
	if err != nil {
		return nil, errors.Join(ErrPermissionCreate, err)
	}

	actions := opts.Actions
	if len(actions) == 0 && opts.RoleID != nil {
		parsed, parseErr := s.actionsForRole(ctx, *opts.RoleID)
		if parseErr != nil {
			return nil, errors.Join(ErrPermissionCreate, parseErr)
		}
		actions = parsed
	}

	for _, action := range actions {
		if _, ok := held[action]; !ok {
			return nil, errors.Join(ErrPermissionCreate, model.ErrPrivilegeEscalation)
		}
	}

	return s.Create(ctx, opts)
}

func (s *permissionService) actionsForRole(ctx context.Context, roleID model.ID) ([]model.Action, error) {
	if s.roleRepo == nil {
		return nil, ErrNoRoleRepository
	}
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return model.ParseActions(role.Actions)
}

func (s *permissionService) Get(ctx context.Context, id model.ID) (*Grant, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/Get")
	defer span.End()

	grant, err := s.permissionRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrPermissionGet, err)
	}
	return grantFromRepository(grant), nil
}

func (s *permissionService) ListByPrincipal(ctx context.Context, principal model.ID) ([]*Grant, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/ListByPrincipal")
	defer span.End()

	grants, err := s.permissionRepo.ListByPrincipal(ctx, principal)
	if err != nil {
		return nil, errors.Join(ErrPermissionGetBySubject, err)
	}
	return grantsFromRepository(grants), nil
}

func (s *permissionService) ListByScope(ctx context.Context, scope model.ID) ([]*Grant, error) {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/ListByScope")
	defer span.End()

	grants, err := s.permissionRepo.ListByScope(ctx, scope)
	if err != nil {
		return nil, errors.Join(ErrPermissionGetByTarget, err)
	}
	return grantsFromRepository(grants), nil
}

func (s *permissionService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/Delete")
	defer span.End()

	if err := s.permissionRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrPermissionDelete, err)
	}
	return nil
}

func (s *permissionService) CtxUserDelete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/CtxUserDelete")
	defer span.End()

	if _, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID); !ok {
		return errors.Join(ErrPermissionDelete, ErrNoUser)
	}

	grant, err := s.Get(ctx, id)
	if err != nil {
		return errors.Join(ErrPermissionDelete, err)
	}

	if !s.CtxUserHas(ctx, grant.Scope, model.ActionPermissionManage) {
		return errors.Join(ErrPermissionDelete, ErrNoPermission)
	}

	return s.Delete(ctx, id)
}

func (s *permissionService) LinkInScopeOf(ctx context.Context, child, parent model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/LinkInScopeOf")
	defer span.End()

	if err := s.permissionRepo.LinkInScopeOf(ctx, child, parent); err != nil {
		return errors.Join(ErrPermissionCreate, err)
	}
	return nil
}

func (s *permissionService) BootstrapCreator(ctx context.Context, creator, resource model.ID, actions []model.Action) error {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/BootstrapCreator")
	defer span.End()

	_, err := s.Create(ctx, CreateGrantOpts{
		Principal: creator,
		Scope:     resource,
		Actions:   actions,
	})
	return err
}

func (s *permissionService) GrantRole(ctx context.Context, principal, scope model.ID, roleID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/GrantRole")
	defer span.End()

	_, err := s.Create(ctx, CreateGrantOpts{
		Principal: principal,
		Scope:     scope,
		RoleID:    &roleID,
	})
	return err
}

func (s *permissionService) BumpGeneration(ctx context.Context, principal model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.permissionService/BumpGeneration")
	defer span.End()

	if err := s.permissionRepo.BumpGeneration(ctx, principal); err != nil {
		return errors.Join(ErrPermissionUpdate, err)
	}
	return nil
}

func roleTemplateActions(key string) ([]model.Action, error) {
	tmpl, err := model.RoleTemplateByKey(key)
	if err != nil {
		return nil, err
	}
	return tmpl.Actions, nil
}

// NewPermissionService creates a PermissionService that evaluates and
// administers grants through permissionRepo.
func NewPermissionService(permissionRepo repository.PermissionRepository, opts ...Option) (PermissionService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &permissionService{
		baseService:    s,
		permissionRepo: permissionRepo,
	}

	if svc.permissionRepo == nil {
		return nil, ErrNoPermissionRepository
	}

	return svc, nil
}
