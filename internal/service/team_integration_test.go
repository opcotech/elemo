package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type TeamServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	teamService       service.TeamService
	permissionService service.PermissionService

	owner        *repository.User
	organization *repository.Organization
	ctx          context.Context
}

func (s *TeamServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)

	permissionService, err := service.NewPermissionService(s.PermissionRepo)
	s.Require().NoError(err)
	s.permissionService = permissionService

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		service.WithPermissionService(permissionService),
	)
	s.Require().NoError(err)

	s.teamService, err = service.NewTeamService(
		service.WithTeamRepository(s.TeamRepo),
		service.WithPermissionService(permissionService),
		service.WithLicenseService(licenseService),
	)
	s.Require().NoError(err)
}

func (s *TeamServiceIntegrationTestSuite) SetupTest() {
	var err error
	s.owner, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)

	s.organization, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.owner.ID))
	s.Require().NoError(err)

	_, err = s.PermissionRepo.Create(context.Background(), repository.CreateGrantOpts{
		Principal: s.owner.ID,
		Scope:     s.organization.ID,
		Actions:   testModel.OrgAdminActions(),
	})
	s.Require().NoError(err)
}

func (s *TeamServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *TeamServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *TeamServiceIntegrationTestSuite) TestCreate() {
	team, err := s.teamService.Create(s.ctx, s.organization.ID, service.CreateTeamOpts{
		Name:        "test-team",
		Description: "test team description",
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(team.ID)
	s.Assert().NotNil(team.CreatedAt)
}

func (s *TeamServiceIntegrationTestSuite) TestGet() {
	created, err := s.teamService.Create(s.ctx, s.organization.ID, service.CreateTeamOpts{
		Name: "get-team", Description: "get team description",
	})
	s.Require().NoError(err)

	team, err := s.teamService.Get(s.ctx, created.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, team.ID)
	s.Assert().Equal(created.Name, team.Name)
}

func (s *TeamServiceIntegrationTestSuite) TestListBelongsTo() {
	_, err := s.teamService.Create(s.ctx, s.organization.ID, service.CreateTeamOpts{
		Name: "team-one", Description: "team one description",
	})
	s.Require().NoError(err)
	_, err = s.teamService.Create(s.ctx, s.organization.ID, service.CreateTeamOpts{
		Name: "team-two", Description: "team two description",
	})
	s.Require().NoError(err)

	teams, err := s.teamService.ListBelongsTo(s.ctx, s.organization.ID, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(len(teams.Items), 2)
}

func (s *TeamServiceIntegrationTestSuite) TestUpdate() {
	created, err := s.teamService.Create(s.ctx, s.organization.ID, service.CreateTeamOpts{
		Name: "upd-team", Description: "upd team description",
	})
	s.Require().NoError(err)

	team, err := s.teamService.Update(s.ctx, created.ID, s.organization.ID, service.UpdateTeamOpts{
		Name: optional.Some("updated-team"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("updated-team", team.Name)
	s.Assert().NotNil(team.UpdatedAt)
}

func (s *TeamServiceIntegrationTestSuite) TestAddMember() {
	created, err := s.teamService.Create(s.ctx, s.organization.ID, service.CreateTeamOpts{
		Name: "member-team", Description: "member team description",
	})
	s.Require().NoError(err)

	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), s.organization.ID, member.ID))

	s.Require().NoError(s.teamService.AddMember(s.ctx, created.ID, member.ID, s.organization.ID))

	members, err := s.teamService.ListMembers(s.ctx, created.ID, s.organization.ID, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(len(members.Items), 1)
}

func (s *TeamServiceIntegrationTestSuite) TestRemoveMember() {
	created, err := s.teamService.Create(s.ctx, s.organization.ID, service.CreateTeamOpts{
		Name: "rm-team", Description: "rm team description",
	})
	s.Require().NoError(err)

	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), s.organization.ID, member.ID))
	s.Require().NoError(s.teamService.AddMember(s.ctx, created.ID, member.ID, s.organization.ID))
	s.Require().NoError(s.teamService.RemoveMember(s.ctx, created.ID, member.ID, s.organization.ID))
}

func (s *TeamServiceIntegrationTestSuite) TestDelete() {
	created, err := s.teamService.Create(s.ctx, s.organization.ID, service.CreateTeamOpts{
		Name: "del-team", Description: "del team description",
	})
	s.Require().NoError(err)
	s.Require().NoError(s.teamService.Delete(s.ctx, created.ID, s.organization.ID))
}

func (s *TeamServiceIntegrationTestSuite) TestGuestDenied() {
	guest, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	guestCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, guest.ID)

	_, err = s.teamService.Create(guestCtx, s.organization.ID, service.CreateTeamOpts{
		Name: "denied-team", Description: "denied team description",
	})
	s.Require().ErrorIs(err, service.ErrNoPermission)

	_, err = s.teamService.Get(guestCtx, model.MustNewID(model.ResourceTypeTeam), s.organization.ID)
	s.Require().ErrorIs(err, service.ErrNoPermission)
}

func (s *TeamServiceIntegrationTestSuite) TestTeamGrantAuthorizesMember() {
	ns, err := s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.owner.ID, s.organization.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(ns.ID, s.owner.ID))
	s.Require().NoError(err)

	team, err := s.teamService.Create(s.ctx, s.organization.ID, service.CreateTeamOpts{
		Name: "rebac-team", Description: "rebac team description",
	})
	s.Require().NoError(err)

	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), s.organization.ID, member.ID))

	_, err = s.PermissionRepo.Create(context.Background(), repository.CreateGrantOpts{
		Principal: team.ID,
		Scope:     project.ID,
		Actions:   []model.Action{model.ActionProjectRead},
	})
	s.Require().NoError(err)

	memberCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, member.ID)
	s.Assert().False(s.permissionService.CtxUserHas(memberCtx, project.ID, model.ActionProjectRead))

	s.Require().NoError(s.teamService.AddMember(s.ctx, team.ID, member.ID, s.organization.ID))
	s.Assert().True(s.permissionService.CtxUserHas(memberCtx, project.ID, model.ActionProjectRead))
}

func TestTeamServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TeamServiceIntegrationTestSuite))
}
