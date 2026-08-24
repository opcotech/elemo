package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// Team represents a team returned by the service.
type Team struct {
	ID          model.ID
	Name        string
	Description string
	MemberCount *int64
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// CreateTeamOpts holds the data required to create a team.
type CreateTeamOpts struct {
	Name        string `json:"name" validate:"required,min=3,max=120"`
	Description string `json:"description" validate:"omitempty,min=5,max=500"`
}

// Validate validates the create options.
func (o *CreateTeamOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidTeamDetails, err)
	}
	return nil
}

// UpdateTeamOpts holds the fields that can be updated on a team.
// Undefined fields (Defined == false) are left unchanged.
type UpdateTeamOpts struct {
	Name        optional.Optional[string]
	Description optional.Optional[string]
}

// TeamService is the interface that provides methods for managing teams.
//
//go:generate go tool mockgen -destination=mock/mock_team_gen.go -package=mocksvc . TeamService
type TeamService interface {
	// Create creates a new team under an organization or project. The caller
	// must have team.manage on the parent resource.
	Create(ctx context.Context, belongsTo model.ID, opts CreateTeamOpts) (*Team, error)
	// Get returns a team by its ID. If the team does not exist, an error is
	// returned.
	Get(ctx context.Context, id, belongsTo model.ID) (*Team, error)
	// ListBelongsTo returns a cursor-paginated page of teams for a resource.
	ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage) (Page[*Team], error)
	// ListMembers returns a cursor-paginated page of members of a team that
	// belongs to a resource. If the resource does not exist, an error is
	// returned.
	ListMembers(ctx context.Context, id, belongsTo model.ID, page CursorPage) (Page[*User], error)
	// Update updates a team. If the team does not exist, an error is returned.
	Update(ctx context.Context, id, belongsTo model.ID, opts UpdateTeamOpts) (*Team, error)
	// AddMember adds a member to a team. If the member is already a member of
	// the team, the membership is refreshed.
	AddMember(ctx context.Context, teamID, memberID, belongsToID model.ID) error
	// RemoveMember removes a member from a team. If the member is not a member
	// of the team, an error is returned.
	RemoveMember(ctx context.Context, teamID, memberID, belongsToID model.ID) error
	// Delete deletes a team from the system.
	Delete(ctx context.Context, id, belongsTo model.ID) error
}

// teamService implements TeamService interface.
type teamService struct {
	runtime
	teamRepo          repository.TeamRepository
	permissionService PermissionService
	licenseService    LicenseService
}

func teamFromRepository(t *repository.Team) *Team {
	if t == nil {
		return nil
	}
	return &Team{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		MemberCount: t.MemberCount,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func isTeamScope(id model.ID) bool {
	return id.Type == model.ResourceTypeOrganization || id.Type == model.ResourceTypeProject
}

func (s *teamService) requireScopeRead(ctx context.Context, belongsTo model.ID) error {
	action, ok := model.ReadActionFor(belongsTo.Type)
	if !ok {
		return ErrNoPermission
	}
	if err := requireAction(ctx, s.permissionService, belongsTo, action); err != nil {
		return err
	}
	return nil
}

func (s *teamService) requireTeamManage(ctx context.Context, belongsTo model.ID) error {
	if err := requireAction(ctx, s.permissionService, belongsTo, model.ActionTeamManage); err != nil {
		return err
	}
	return nil
}

func (s *teamService) Create(ctx context.Context, belongsTo model.ID, opts CreateTeamOpts) (*Team, error) {
	ctx, span := s.tracer.Start(ctx, "service.teamService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrTeamCreate, license.ErrLicenseExpired)
	}

	if err := belongsTo.Validate(); err != nil {
		return nil, errors.Join(ErrTeamCreate, err)
	}

	if !isTeamScope(belongsTo) {
		return nil, errors.Join(ErrTeamCreate, model.ErrInvalidTeamDetails)
	}

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrTeamCreate, err)
	}

	if err := s.requireTeamManage(ctx, belongsTo); err != nil {
		return nil, errors.Join(ErrTeamCreate, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, errors.Join(ErrTeamCreate, err)
	}

	team, err := s.teamRepo.Create(ctx, repository.CreateTeamOpts{
		Name:        opts.Name,
		Description: opts.Description,
		CreatedBy:   userID,
		BelongsTo:   belongsTo,
	})
	if err != nil {
		return nil, errors.Join(ErrTeamCreate, err)
	}

	return teamFromRepository(team), nil
}

func (s *teamService) Get(ctx context.Context, id, belongsTo model.ID) (*Team, error) {
	ctx, span := s.tracer.Start(ctx, "service.teamService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrTeamGet, err)
	}

	if err := belongsTo.Validate(); err != nil {
		return nil, errors.Join(ErrTeamGet, err)
	}

	if err := s.requireScopeRead(ctx, belongsTo); err != nil {
		return nil, errors.Join(ErrTeamGet, err)
	}

	team, err := s.teamRepo.Get(ctx, id, belongsTo, repository.TeamDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrTeamGet, err)
	}

	return teamFromRepository(team), nil
}

func (s *teamService) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage) (Page[*Team], error) {
	ctx, span := s.tracer.Start(ctx, "service.teamService/ListBelongsTo")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return Page[*Team]{}, errors.Join(ErrTeamGetBelongsTo, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Team]{}, errors.Join(ErrTeamGetBelongsTo, err)
	}

	if err := s.requireScopeRead(ctx, belongsTo); err != nil {
		return Page[*Team]{}, errors.Join(ErrTeamGetBelongsTo, err)
	}

	teams, err := s.teamRepo.ListBelongsTo(
		ctx,
		belongsTo,
		normalized,
		repository.TeamListProjection(),
	)
	if err != nil {
		return Page[*Team]{}, errors.Join(ErrTeamGetBelongsTo, err)
	}

	return mapPage(teams, teamFromRepository), nil
}

func (s *teamService) ListMembers(ctx context.Context, id, belongsTo model.ID, page CursorPage) (Page[*User], error) {
	ctx, span := s.tracer.Start(ctx, "service.teamService/ListMembers")
	defer span.End()

	if err := id.Validate(); err != nil {
		return Page[*User]{}, errors.Join(ErrTeamGetBelongsTo, err)
	}

	if err := belongsTo.Validate(); err != nil {
		return Page[*User]{}, errors.Join(ErrTeamGetBelongsTo, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*User]{}, errors.Join(ErrTeamGetBelongsTo, err)
	}

	if err := s.requireScopeRead(ctx, belongsTo); err != nil {
		return Page[*User]{}, errors.Join(ErrTeamGetBelongsTo, err)
	}

	members, err := s.teamRepo.ListMembers(ctx, id, belongsTo, normalized)
	if err != nil {
		return Page[*User]{}, errors.Join(ErrTeamGetBelongsTo, err)
	}

	return mapPage(members, userFromRepository), nil
}

func (s *teamService) Update(ctx context.Context, id, belongsTo model.ID, opts UpdateTeamOpts) (*Team, error) {
	ctx, span := s.tracer.Start(ctx, "service.teamService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrTeamUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrTeamUpdate, err)
	}

	if err := belongsTo.Validate(); err != nil {
		return nil, errors.Join(ErrTeamUpdate, err)
	}

	if err := s.requireTeamManage(ctx, belongsTo); err != nil {
		return nil, errors.Join(ErrTeamUpdate, err)
	}

	team, err := s.teamRepo.Update(ctx, id, belongsTo, repository.UpdateTeamOpts{
		Name:        opts.Name,
		Description: opts.Description,
	})
	if err != nil {
		return nil, errors.Join(ErrTeamUpdate, err)
	}

	return teamFromRepository(team), nil
}

func (s *teamService) AddMember(ctx context.Context, teamID, memberID, belongsToID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.teamService/AddMember")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrTeamAddMember, license.ErrLicenseExpired)
	}

	if err := teamID.Validate(); err != nil {
		return errors.Join(ErrTeamAddMember, err)
	}

	if err := memberID.Validate(); err != nil {
		return errors.Join(ErrTeamAddMember, err)
	}

	if err := belongsToID.Validate(); err != nil {
		return errors.Join(ErrTeamAddMember, err)
	}

	if err := s.requireTeamManage(ctx, belongsToID); err != nil {
		return errors.Join(ErrTeamAddMember, err)
	}

	if err := s.teamRepo.AddMember(ctx, teamID, memberID, belongsToID); err != nil {
		return errors.Join(ErrTeamAddMember, err)
	}

	if s.permissionService != nil {
		_ = s.permissionService.BumpGeneration(ctx, memberID)
	}

	return nil
}

func (s *teamService) RemoveMember(ctx context.Context, teamID, memberID, belongsToID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.teamService/RemoveMember")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrTeamRemoveMember, license.ErrLicenseExpired)
	}

	if err := teamID.Validate(); err != nil {
		return errors.Join(ErrTeamRemoveMember, err)
	}

	if err := memberID.Validate(); err != nil {
		return errors.Join(ErrTeamRemoveMember, err)
	}

	if err := belongsToID.Validate(); err != nil {
		return errors.Join(ErrTeamRemoveMember, err)
	}

	if err := s.requireTeamManage(ctx, belongsToID); err != nil {
		return errors.Join(ErrTeamRemoveMember, err)
	}

	if err := s.teamRepo.RemoveMember(ctx, teamID, memberID, belongsToID); err != nil {
		return errors.Join(ErrTeamRemoveMember, err)
	}

	if s.permissionService != nil {
		_ = s.permissionService.BumpGeneration(ctx, memberID)
	}

	return nil
}

func (s *teamService) Delete(ctx context.Context, id, belongsTo model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.teamService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrTeamDelete, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrTeamDelete, err)
	}

	if err := belongsTo.Validate(); err != nil {
		return errors.Join(ErrTeamDelete, err)
	}

	if err := s.requireTeamManage(ctx, belongsTo); err != nil {
		return errors.Join(ErrTeamDelete, err)
	}

	if err := s.teamRepo.Delete(ctx, id, belongsTo); err != nil {
		return errors.Join(ErrTeamDelete, err)
	}

	return nil
}

// NewTeamService creates a new TeamService that provides methods
// for managing teams.
func NewTeamService(
	teamRepo repository.TeamRepository,
	permissionService PermissionService,
	licenseService LicenseService,
	opts ...Option,
) (TeamService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}

	svc := &teamService{
		runtime:           rt,
		teamRepo:          teamRepo,
		permissionService: permissionService,
		licenseService:    licenseService,
	}

	if svc.teamRepo == nil {
		return nil, ErrNoTeamRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	return svc, nil
}
