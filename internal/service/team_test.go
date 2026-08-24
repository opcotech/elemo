package service_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func TestNewTeamService(t *testing.T) {
	tests := []struct {
		name    string
		build   func(ctrl *gomock.Controller) (service.TeamService, error)
		wantErr error
	}{
		{
			name: "new team service",
			build: func(ctrl *gomock.Controller) (service.TeamService, error) {
				return service.NewTeamService(mockrepo.NewMockTeamRepository(nil), mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
		},
		{
			name: "new team service with no team repository",
			build: func(ctrl *gomock.Controller) (service.TeamService, error) {
				return service.NewTeamService(nil, mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoTeamRepository,
		},
		{
			name: "new team service with no permission service",
			build: func(ctrl *gomock.Controller) (service.TeamService, error) {
				return service.NewTeamService(mockrepo.NewMockTeamRepository(nil), nil, mocksvc.NewMockLicenseService(nil), service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoPermissionService,
		},
		{
			name: "new team service with no license service",
			build: func(ctrl *gomock.Controller) (service.TeamService, error) {
				return service.NewTeamService(mockrepo.NewMockTeamRepository(nil), mocksvc.NewMockPermissionService(nil), nil, service.WithLogger(mocklog.NewMockLogger(ctrl)), service.WithTracer(mocktrace.NewMockTracer(ctrl)))
			},
			wantErr: service.ErrNoLicenseService,
		},
		{
			name: "new team service with invalid options",
			build: func(_ *gomock.Controller) (service.TeamService, error) {
				return service.NewTeamService(mockrepo.NewMockTeamRepository(nil), mocksvc.NewMockPermissionService(nil), mocksvc.NewMockLicenseService(nil), service.WithLogger(nil))
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			got, err := tt.build(ctrl)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, got)
			}
		})
	}
}

//nolint:revive // test factories take gomock.Controller first
func newTeamServiceForTest(ctrl *gomock.Controller, ctx context.Context, spanName string) (service.TeamService, *mockrepo.MockTeamRepository, *mocksvc.MockPermissionService, *mocksvc.MockLicenseService) {
	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))

	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, spanName, gomock.Len(0)).Return(ctx, span)

	teamRepo := mockrepo.NewMockTeamRepository(ctrl)
	permSvc := mocksvc.NewMockPermissionService(ctrl)
	licenseSvc := mocksvc.NewMockLicenseService(ctrl)

	return func() service.TeamService {
		svc, err := service.NewTeamService(
			teamRepo,
			permSvc,
			licenseSvc,
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		if err != nil {
			panic(err)
		}
		return svc
	}(), teamRepo, permSvc, licenseSvc
}

func TestTeamService_Create(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	opts := service.CreateTeamOpts{Name: "test-team", Description: "test description"}
	repoTeam := testModel.NewRepositoryTeam()
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	t.Run("create new team", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, teamRepo, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		teamRepo.EXPECT().Create(ctx, repository.CreateTeamOpts{
			Name: opts.Name, Description: opts.Description, CreatedBy: userID, BelongsTo: orgID,
		}).Return(repoTeam, nil)

		got, err := s.Create(ctx, orgID, opts)
		require.NoError(t, err)
		assert.Equal(t, repoTeam.ID, got.ID)
		assert.Equal(t, repoTeam.Name, got.Name)
	})

	t.Run("create new team with repo error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, teamRepo, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		teamRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, assert.AnError)

		_, err := s.Create(ctx, orgID, opts)
		require.ErrorIs(t, err, service.ErrTeamCreate)
	})

	t.Run("create new team with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		_, err := s.Create(ctx, orgID, opts)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("create new team with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false, nil)

		_, err := s.Create(ctx, orgID, opts)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})

	t.Run("create new team with invalid belongs to", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		_, err := s.Create(ctx, model.MustNewID(model.ResourceTypeUser), opts)
		require.ErrorIs(t, err, model.ErrInvalidTeamDetails)
	})

	t.Run("create new team with invalid opts", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		_, err := s.Create(ctx, orgID, service.CreateTeamOpts{Name: "ab"})
		require.ErrorIs(t, err, service.ErrTeamCreate)
	})

	t.Run("create new team with no user", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		emptyCtx := context.Background()
		s, _, permSvc, licenseSvc := newTeamServiceForTest(ctrl, emptyCtx, "service.teamService/Create")
		licenseSvc.EXPECT().Expired(emptyCtx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(emptyCtx, orgID, model.ActionTeamManage).Return(true, nil)

		_, err := s.Create(emptyCtx, orgID, opts)
		require.ErrorIs(t, err, service.ErrNoUser)
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
		s, teamRepo, permSvc, _ := newTeamServiceForTest(ctrl, ctx, "service.teamService/Get")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(true, nil)
		teamRepo.EXPECT().Get(ctx, teamID, orgID, repository.TeamDetailProjection()).Return(repoTeam, nil)

		got, err := s.Get(ctx, teamID, orgID)
		require.NoError(t, err)
		assert.Equal(t, teamID, got.ID)
	})

	t.Run("get team with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, _ := newTeamServiceForTest(ctrl, ctx, "service.teamService/Get")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(false, nil)

		_, err := s.Get(ctx, teamID, orgID)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})

	t.Run("get team with repo error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, teamRepo, permSvc, _ := newTeamServiceForTest(ctrl, ctx, "service.teamService/Get")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(true, nil)
		teamRepo.EXPECT().Get(ctx, teamID, orgID, repository.TeamDetailProjection()).Return(nil, assert.AnError)

		_, err := s.Get(ctx, teamID, orgID)
		require.ErrorIs(t, err, service.ErrTeamGet)
	})

	t.Run("get team with invalid id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, _ := newTeamServiceForTest(ctrl, ctx, "service.teamService/Get")

		_, err := s.Get(ctx, model.ID{}, orgID)
		require.ErrorIs(t, err, service.ErrTeamGet)
	})
}

func TestTeamService_ListBelongsTo(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	repoTeam := testModel.NewRepositoryTeam()
	ctx := context.Background()
	page := service.CursorPage{Size: 10}

	t.Run("list teams", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, teamRepo, permSvc, _ := newTeamServiceForTest(ctrl, ctx, "service.teamService/ListBelongsTo")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(true, nil)
		teamRepo.EXPECT().ListBelongsTo(ctx, orgID, page, repository.TeamListProjection()).Return(repository.Page[*repository.Team]{
			Items: []*repository.Team{repoTeam},
		}, nil)

		got, err := s.ListBelongsTo(ctx, orgID, page)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, repoTeam.ID, got.Items[0].ID)
	})

	t.Run("list teams with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, _ := newTeamServiceForTest(ctrl, ctx, "service.teamService/ListBelongsTo")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(false, nil)

		_, err := s.ListBelongsTo(ctx, orgID, page)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestTeamService_ListMembers(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	repoUser := testModel.NewRepositoryUser()
	ctx := context.Background()
	page := service.CursorPage{Size: 10}

	t.Run("list members", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, teamRepo, permSvc, _ := newTeamServiceForTest(ctrl, ctx, "service.teamService/ListMembers")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(true, nil)
		teamRepo.EXPECT().ListMembers(ctx, teamID, orgID, page).Return(repository.Page[*repository.User]{
			Items: []*repository.User{repoUser},
		}, nil)

		got, err := s.ListMembers(ctx, teamID, orgID, page)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, repoUser.ID, got.Items[0].ID)
	})

	t.Run("list members with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, _ := newTeamServiceForTest(ctrl, ctx, "service.teamService/ListMembers")
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(false, nil)

		_, err := s.ListMembers(ctx, teamID, orgID, page)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestTeamService_Update(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	repoTeam := testModel.NewRepositoryTeam()
	repoTeam.Name = "updated-team"
	ctx := context.Background()
	opts := service.UpdateTeamOpts{Name: optional.Some("updated-team")}

	t.Run("update team", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, teamRepo, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		teamRepo.EXPECT().Update(ctx, teamID, orgID, repository.UpdateTeamOpts{Name: opts.Name}).Return(repoTeam, nil)

		got, err := s.Update(ctx, teamID, orgID, opts)
		require.NoError(t, err)
		assert.Equal(t, "updated-team", got.Name)
	})

	t.Run("update team with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		_, err := s.Update(ctx, teamID, orgID, opts)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})

	t.Run("update team with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Update")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false, nil)

		_, err := s.Update(ctx, teamID, orgID, opts)
		require.ErrorIs(t, err, service.ErrNoPermission)
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
		s, teamRepo, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		teamRepo.EXPECT().AddMember(ctx, teamID, memberID, orgID).Return(nil)
		permSvc.EXPECT().BumpGeneration(ctx, memberID).Return(nil)

		require.NoError(t, s.AddMember(ctx, teamID, memberID, orgID))
	})

	t.Run("add member with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false, nil)

		err := s.AddMember(ctx, teamID, memberID, orgID)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})

	t.Run("add member with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/AddMember")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		err := s.AddMember(ctx, teamID, memberID, orgID)
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
		s, teamRepo, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		teamRepo.EXPECT().RemoveMember(ctx, teamID, memberID, orgID).Return(nil)
		permSvc.EXPECT().BumpGeneration(ctx, memberID).Return(nil)

		require.NoError(t, s.RemoveMember(ctx, teamID, memberID, orgID))
	})

	t.Run("remove member with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/RemoveMember")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false, nil)

		err := s.RemoveMember(ctx, teamID, memberID, orgID)
		require.ErrorIs(t, err, service.ErrNoPermission)
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
		s, teamRepo, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(true, nil)
		teamRepo.EXPECT().Delete(ctx, teamID, orgID).Return(nil)

		require.NoError(t, s.Delete(ctx, teamID, orgID))
	})

	t.Run("delete team with no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, permSvc, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(ctx, orgID, model.ActionTeamManage).Return(false, nil)

		err := s.Delete(ctx, teamID, orgID)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})

	t.Run("delete team with expired license", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		s, _, _, licenseSvc := newTeamServiceForTest(ctrl, ctx, "service.teamService/Delete")
		licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

		err := s.Delete(ctx, teamID, orgID)
		require.ErrorIs(t, err, license.ErrLicenseExpired)
	})
}
