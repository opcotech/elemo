package service

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTeamService(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    TeamService
		wantErr error
	}{
		{
			name: "new team service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithTeamRepository(repository.NewMockTeamRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			want: &teamService{
				baseService: &baseService{
					logger:            mock.NewMockLogger(nil),
					tracer:            mock.NewMockTracer(nil),
					teamRepo:          repository.NewMockTeamRepository(nil),
					permissionService: NewMockPermissionService(nil),
					licenseService:    mock.NewMockLicenseService(nil),
				},
			},
		},
		{
			name: "new team service with invalid options",
			args: args{
				opts: []Option{
					WithLogger(nil),
					WithTracer(mock.NewMockTracer(nil)),
					WithTeamRepository(repository.NewMockTeamRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new team service with no team repository",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: ErrNoTeamRepository,
		},
		{
			name: "new team service with no permission service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithTeamRepository(repository.NewMockTeamRepository(nil)),
					WithLicenseService(mock.NewMockLicenseService(nil)),
				},
			},
			wantErr: ErrNoPermissionService,
		},
		{
			name: "new team service with no license service",
			args: args{
				opts: []Option{
					WithLogger(mock.NewMockLogger(nil)),
					WithTracer(mock.NewMockTracer(nil)),
					WithTeamRepository(repository.NewMockTeamRepository(nil)),
					WithPermissionService(NewMockPermissionService(nil)),
				},
			},
			wantErr: ErrNoLicenseService,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewTeamService(tt.args.opts...)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

//nolint:revive // test factories take gomock.Controller first
func newTeamServiceTestBase(ctrl *gomock.Controller, ctx context.Context, spanName string) (*baseService, *repository.MockTeamRepository, *MockPermissionService, *mock.MockLicenseService) {
	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))

	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, spanName, gomock.Len(0)).Return(ctx, span)

	teamRepo := repository.NewMockTeamRepository(ctrl)
	permSvc := NewMockPermissionService(ctrl)
	licenseSvc := mock.NewMockLicenseService(ctrl)

	return &baseService{
		logger:            mock.NewMockLogger(ctrl),
		tracer:            tracer,
		teamRepo:          teamRepo,
		permissionService: permSvc,
		licenseService:    licenseSvc,
	}, teamRepo, permSvc, licenseSvc
}

func TestTeamService_Create(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	opts := CreateTeamOpts{Name: "test-team", Description: "test description"}
	repoTeam := testModel.NewRepositoryTeam()
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	t.Run("create new team", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, teamRepo, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		teamRepo.EXPECT().Create(ctx, repository.CreateTeamOpts{
			Name: opts.Name, Description: opts.Description, CreatedBy: userID, BelongsTo: orgID,
		}).Return(repoTeam, nil)

		got, err := (&teamService{baseService: base}).Create(ctx, orgID, opts)
		require.NoError(t, err)
		assert.Equal(t, repoTeam.ID, got.ID)
		assert.Equal(t, repoTeam.Name, got.Name)
	})

	t.Run("create new team with repo error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, teamRepo, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		teamRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, assert.AnError)

		_, err := (&teamService{baseService: base}).Create(ctx, orgID, opts)
		require.ErrorIs(t, err, ErrTeamCreate)
	})

	t.Run("create new team with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		_, err := (&teamService{baseService: base}).Create(ctx, orgID, opts)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("create new team with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false)

		_, err := (&teamService{baseService: base}).Create(ctx, orgID, opts)
		require.ErrorIs(t, err, ErrNoPermission)
	})

	t.Run("create new team with invalid belongs to", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		_, err := (&teamService{baseService: base}).Create(ctx, model.MustNewID(model.ResourceTypeUser), opts)
		require.ErrorIs(t, err, model.ErrInvalidTeamDetails)
	})

	t.Run("create new team with invalid opts", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		_, err := (&teamService{baseService: base}).Create(ctx, orgID, CreateTeamOpts{Name: "ab"})
		require.ErrorIs(t, err, ErrTeamCreate)
	})

	t.Run("create new team with no user", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		emptyCtx := context.Background()
		base, _, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, emptyCtx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(emptyCtx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(emptyCtx, orgID, model.ActionTeamManage).Return(true)

		_, err := (&teamService{baseService: base}).Create(emptyCtx, orgID, opts)
		require.ErrorIs(t, err, ErrNoUser)
	})
}

func TestTeamService_Get(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	repoTeam := testModel.NewRepositoryTeam()
	repoTeam.ID = teamID
	ctx := context.Background()

	t.Run("get team", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, teamRepo, permSvc, _ := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Get")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(true)
		teamRepo.EXPECT().Get(ctx, teamID, orgID, repository.TeamDetailProjection()).Return(repoTeam, nil)

		got, err := (&teamService{baseService: base}).Get(ctx, teamID, orgID)
		require.NoError(t, err)
		assert.Equal(t, teamID, got.ID)
	})

	t.Run("get team with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, _ := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Get")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(false)

		_, err := (&teamService{baseService: base}).Get(ctx, teamID, orgID)
		require.ErrorIs(t, err, ErrNoPermission)
	})

	t.Run("get team with repo error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, teamRepo, permSvc, _ := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Get")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(true)
		teamRepo.EXPECT().Get(ctx, teamID, orgID, repository.TeamDetailProjection()).Return(nil, assert.AnError)

		_, err := (&teamService{baseService: base}).Get(ctx, teamID, orgID)
		require.ErrorIs(t, err, ErrTeamGet)
	})

	t.Run("get team with invalid id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, _ := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Get")

		_, err := (&teamService{baseService: base}).Get(ctx, model.ID{}, orgID)
		require.ErrorIs(t, err, ErrTeamGet)
	})
}

func TestTeamService_ListBelongsTo(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	repoTeam := testModel.NewRepositoryTeam()
	ctx := context.Background()
	page := CursorPage{Size: 10}

	t.Run("list teams", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, teamRepo, permSvc, _ := newTeamServiceTestBase(ctrl, ctx, "service.teamService/ListBelongsTo")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(true)
		teamRepo.EXPECT().ListBelongsTo(ctx, orgID, page, repository.TeamListProjection()).Return(repository.Page[*repository.Team]{
			Items: []*repository.Team{repoTeam},
		}, nil)

		got, err := (&teamService{baseService: base}).ListBelongsTo(ctx, orgID, page)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, repoTeam.ID, got.Items[0].ID)
	})

	t.Run("list teams with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, _ := newTeamServiceTestBase(ctrl, ctx, "service.teamService/ListBelongsTo")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(false)

		_, err := (&teamService{baseService: base}).ListBelongsTo(ctx, orgID, page)
		require.ErrorIs(t, err, ErrNoPermission)
	})
}

func TestTeamService_ListMembers(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	repoUser := testModel.NewRepositoryUser()
	ctx := context.Background()
	page := CursorPage{Size: 10}

	t.Run("list members", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, teamRepo, permSvc, _ := newTeamServiceTestBase(ctrl, ctx, "service.teamService/ListMembers")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(true)
		teamRepo.EXPECT().ListMembers(ctx, teamID, orgID, page).Return(repository.Page[*repository.User]{
			Items: []*repository.User{repoUser},
		}, nil)

		got, err := (&teamService{baseService: base}).ListMembers(ctx, teamID, orgID, page)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, repoUser.ID, got.Items[0].ID)
	})

	t.Run("list members with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, _ := newTeamServiceTestBase(ctrl, ctx, "service.teamService/ListMembers")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(false)

		_, err := (&teamService{baseService: base}).ListMembers(ctx, teamID, orgID, page)
		require.ErrorIs(t, err, ErrNoPermission)
	})
}

func TestTeamService_Update(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	repoTeam := testModel.NewRepositoryTeam()
	repoTeam.Name = "updated-team"
	ctx := context.Background()
	opts := UpdateTeamOpts{Name: optional.Some("updated-team")}

	t.Run("update team", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, teamRepo, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		teamRepo.EXPECT().Update(ctx, teamID, orgID, repository.UpdateTeamOpts{Name: opts.Name}).Return(repoTeam, nil)

		got, err := (&teamService{baseService: base}).Update(ctx, teamID, orgID, opts)
		require.NoError(t, err)
		assert.Equal(t, "updated-team", got.Name)
	})

	t.Run("update team with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		_, err := (&teamService{baseService: base}).Update(ctx, teamID, orgID, opts)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("update team with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false)

		_, err := (&teamService{baseService: base}).Update(ctx, teamID, orgID, opts)
		require.ErrorIs(t, err, ErrNoPermission)
	})
}

func TestTeamService_AddMember(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	memberID := model.MustNewID(model.ResourceTypeUser)
	ctx := context.Background()

	t.Run("add member", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, teamRepo, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		teamRepo.EXPECT().AddMember(ctx, teamID, memberID, orgID).Return(nil)
		permSvc.EXPECT().BumpGeneration(ctx, memberID).Return(nil)

		require.NoError(t, (&teamService{baseService: base}).AddMember(ctx, teamID, memberID, orgID))
	})

	t.Run("add member with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false)

		err := (&teamService{baseService: base}).AddMember(ctx, teamID, memberID, orgID)
		require.ErrorIs(t, err, ErrNoPermission)
	})

	t.Run("add member with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		err := (&teamService{baseService: base}).AddMember(ctx, teamID, memberID, orgID)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})
}

func TestTeamService_RemoveMember(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	memberID := model.MustNewID(model.ResourceTypeUser)
	ctx := context.Background()

	t.Run("remove member", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, teamRepo, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		teamRepo.EXPECT().RemoveMember(ctx, teamID, memberID, orgID).Return(nil)
		permSvc.EXPECT().BumpGeneration(ctx, memberID).Return(nil)

		require.NoError(t, (&teamService{baseService: base}).RemoveMember(ctx, teamID, memberID, orgID))
	})

	t.Run("remove member with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false)

		err := (&teamService{baseService: base}).RemoveMember(ctx, teamID, memberID, orgID)
		require.ErrorIs(t, err, ErrNoPermission)
	})
}

func TestTeamService_Delete(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	ctx := context.Background()

	t.Run("delete team", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, teamRepo, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true)
		teamRepo.EXPECT().Delete(ctx, teamID, orgID).Return(nil)

		require.NoError(t, (&teamService{baseService: base}).Delete(ctx, teamID, orgID))
	})

	t.Run("delete team with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, permSvc, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false)

		err := (&teamService{baseService: base}).Delete(ctx, teamID, orgID)
		require.ErrorIs(t, err, ErrNoPermission)
	})

	t.Run("delete team with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		base, _, _, licenseSvc := newTeamServiceTestBase(ctrl, ctx, "service.teamService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		err := (&teamService{baseService: base}).Delete(ctx, teamID, orgID)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})
}
