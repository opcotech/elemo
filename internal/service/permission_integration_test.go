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
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type PermissionServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	permissionService service.PermissionService

	ctx context.Context
}

func (s *PermissionServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	var err error
	s.permissionService, err = service.NewPermissionService(
		s.PermissionRepo,
		s.RoleRepo,
	)
	s.Require().NoError(err)
}

func (s *PermissionServiceIntegrationTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *PermissionServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *PermissionServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *PermissionServiceIntegrationTestSuite) createUser() *repository.User {
	user, err := s.UserRepo.Create(s.ctx, testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	return user
}

func (s *PermissionServiceIntegrationTestSuite) userCtx(userID model.ID) context.Context {
	return context.WithValue(s.ctx, pkg.CtxKeyUserID, userID)
}

func (s *PermissionServiceIntegrationTestSuite) TestDirectAllowDeny() {
	owner := s.createUser()
	guest := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)

	allowed, err := s.permissionService.CtxUserHas(s.userCtx(guest.ID), org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().False(allowed)

	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: guest.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	})
	s.Require().NoError(err)

	allowed, err = s.permissionService.CtxUserHas(s.userCtx(guest.ID), org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().True(allowed)
	allowed, err = s.permissionService.CtxUserHas(s.userCtx(guest.ID), org.ID, model.ActionOrganizationDelete)
	s.Require().NoError(err)
	s.Assert().False(allowed)
}

func (s *PermissionServiceIntegrationTestSuite) TestOrganizationCreateGrant() {
	user := s.createUser()
	ctx := s.userCtx(user.ID)
	allowed, err := s.permissionService.CtxUserHas(ctx, model.InstallationID(), model.ActionOrganizationCreate)
	s.Require().NoError(err)
	s.Assert().False(allowed)

	s.Require().NoError(testRepo.GrantOrganizationCreate(user.ID, s.Neo4jDB))
	allowed, err = s.permissionService.CtxUserHas(ctx, model.InstallationID(), model.ActionOrganizationCreate)
	s.Require().NoError(err)
	s.Assert().True(allowed)
}

func (s *PermissionServiceIntegrationTestSuite) TestOrgAdminDoesNotIncludeOrganizationCreate() {
	owner := s.createUser()
	guest := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)

	tmpl, err := model.RoleTemplateByKey(model.RoleKeyOrgAdmin)
	s.Require().NoError(err)
	role, err := s.RoleRepo.Create(s.ctx, repository.CreateRoleOpts{
		Key:         tmpl.Key,
		Name:        tmpl.Name,
		Description: tmpl.Description,
		Actions:     tmpl.ActionStrings(),
		CreatedBy:   owner.ID,
		BelongsTo:   org.ID,
	})
	s.Require().NoError(err)
	s.Require().NoError(s.permissionService.GrantRole(s.ctx, guest.ID, org.ID, role.ID))

	allowed, err := s.permissionService.CtxUserHas(s.userCtx(guest.ID), org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().True(allowed)
	allowed, err = s.permissionService.CtxUserHas(s.userCtx(guest.ID), model.InstallationID(), model.ActionOrganizationCreate)
	s.Require().NoError(err)
	s.Assert().False(allowed)
}

func (s *PermissionServiceIntegrationTestSuite) TestAncestorAndChildScope() {
	owner := s.createUser()
	actor := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)
	siblingNS, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)

	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: actor.ID,
		Scope:     ns.ID,
		Actions:   []model.Action{model.ActionProjectRead, model.ActionNamespaceRead},
	})
	s.Require().NoError(err)

	ctx := s.userCtx(actor.ID)
	allowed, err := s.permissionService.CtxUserHas(ctx, project.ID, model.ActionProjectRead)
	s.Require().NoError(err)
	s.Assert().True(allowed)
	allowed, err = s.permissionService.CtxUserHas(ctx, siblingNS.ID, model.ActionNamespaceRead)
	s.Require().NoError(err)
	s.Assert().False(allowed)
	allowed, err = s.permissionService.CtxUserHas(ctx, org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().False(allowed)
}

func (s *PermissionServiceIntegrationTestSuite) TestTeamGrant() {
	owner := s.createUser()
	member := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)
	team, err := s.TeamRepo.Create(s.ctx, repository.CreateTeamOpts{
		Name:      "svc-team",
		CreatedBy: owner.ID,
		BelongsTo: org.ID,
	})
	s.Require().NoError(err)

	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: team.ID,
		Scope:     project.ID,
		Actions:   []model.Action{model.ActionProjectRead},
	})
	s.Require().NoError(err)

	ctx := s.userCtx(member.ID)
	allowed, err := s.permissionService.CtxUserHas(ctx, project.ID, model.ActionProjectRead)
	s.Require().NoError(err)
	s.Assert().False(allowed)
	s.Require().NoError(s.TeamRepo.AddMember(s.ctx, team.ID, member.ID, org.ID))
	allowed, err = s.permissionService.CtxUserHas(ctx, project.ID, model.ActionProjectRead)
	s.Require().NoError(err)
	s.Assert().True(allowed)
	s.Require().NoError(s.TeamRepo.RemoveMember(s.ctx, team.ID, member.ID, org.ID))
	allowed, err = s.permissionService.CtxUserHas(ctx, project.ID, model.ActionProjectRead)
	s.Require().NoError(err)
	s.Assert().False(allowed)
}

func (s *PermissionServiceIntegrationTestSuite) TestCreatorWithoutGrantDenied() {
	owner := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	allowed, err := s.permissionService.CtxUserHas(s.userCtx(owner.ID), org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().False(allowed)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserCreateRequiresManageAndHeldActions() {
	owner := s.createUser()
	guest := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)

	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: owner.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionPermissionManage, model.ActionOrganizationRead},
	})
	s.Require().NoError(err)

	ownerCtx := s.userCtx(owner.ID)
	_, err = s.permissionService.CtxUserCreate(ownerCtx, service.CreateGrantOpts{
		Principal: guest.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	})
	s.Require().NoError(err)

	_, err = s.permissionService.CtxUserCreate(ownerCtx, service.CreateGrantOpts{
		Principal: guest.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionOrganizationCreate},
	})
	s.Require().ErrorIs(err, model.ErrPrivilegeEscalation)

	_, err = s.permissionService.CtxUserCreate(s.userCtx(guest.ID), service.CreateGrantOpts{
		Principal: guest.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	})
	s.Require().ErrorIs(err, service.ErrNoPermission)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserDeleteRequiresManage() {
	owner := s.createUser()
	guest := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)

	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: owner.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionPermissionManage, model.ActionOrganizationRead},
	})
	s.Require().NoError(err)

	grant, err := s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: guest.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	})
	s.Require().NoError(err)

	s.Require().ErrorIs(s.permissionService.CtxUserDelete(s.userCtx(guest.ID), grant.ID), service.ErrNoPermission)
	s.Require().NoError(s.permissionService.CtxUserDelete(s.userCtx(owner.ID), grant.ID))
	allowed, err := s.permissionService.CtxUserHas(s.userCtx(guest.ID), org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().False(allowed)
}

func (s *PermissionServiceIntegrationTestSuite) TestRevokeAndDisableDeny() {
	owner := s.createUser()
	actor := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)

	grant, err := s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: actor.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	})
	s.Require().NoError(err)
	allowed, err := s.permissionService.CtxUserHas(s.userCtx(actor.ID), org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().True(allowed)

	s.Require().NoError(s.permissionService.Delete(s.ctx, grant.ID))
	allowed, err = s.permissionService.CtxUserHas(s.userCtx(actor.ID), org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().False(allowed)

	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: actor.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	})
	s.Require().NoError(err)
	_, err = s.UserRepo.Update(s.ctx, actor.ID, repository.UpdateUserOpts{
		Status: optional.Some(model.UserStatusInactive),
	})
	s.Require().NoError(err)
	allowed, err = s.permissionService.CtxUserHas(s.userCtx(actor.ID), org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().False(allowed)
}

func (s *PermissionServiceIntegrationTestSuite) TestListDoesNotLeakUngrantedOrgs() {
	owner := s.createUser()
	stranger := s.createUser()
	visible, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	hidden, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)

	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: stranger.ID,
		Scope:     visible.ID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	})
	s.Require().NoError(err)

	page, err := s.OrganizationRepo.ListForUser(s.ctx, repository.OrganizationListQuery{UserID: stranger.ID, Action: model.ActionOrganizationRead, Page: repository.CursorPage{Size: 20}, Order: repository.SortDirectionDesc, Projection: repository.OrganizationListProjection()})
	s.Require().NoError(err)
	ids := make([]model.ID, 0, len(page.Items))
	for _, org := range page.Items {
		ids = append(ids, org.ID)
	}
	s.Assert().Contains(ids, visible.ID)
	s.Assert().NotContains(ids, hidden.ID)
}

func (s *PermissionServiceIntegrationTestSuite) TestIDORHasWithoutGrant() {
	owner := s.createUser()
	stranger := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: owner.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	})
	s.Require().NoError(err)

	allowed, err := s.permissionService.Has(s.ctx, stranger.ID, org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().False(allowed)
}

func (s *PermissionServiceIntegrationTestSuite) TestEffectiveActionsExplainAndListGrantScopes() {
	owner := s.createUser()
	actor := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)

	empty, err := s.permissionService.EffectiveActions(s.ctx, actor.ID, org.ID)
	s.Require().NoError(err)
	s.Assert().Empty(empty)

	denied, err := s.permissionService.Explain(s.ctx, actor.ID, org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().False(denied.Allowed)

	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: actor.ID,
		Scope:     ns.ID,
		Actions:   []model.Action{model.ActionProjectRead},
	})
	s.Require().NoError(err)

	actions, err := s.permissionService.EffectiveActions(s.ctx, actor.ID, project.ID)
	s.Require().NoError(err)
	s.Assert().Contains(actions, model.ActionProjectRead)

	explained, err := s.permissionService.Explain(s.ctx, actor.ID, project.ID, model.ActionProjectRead)
	s.Require().NoError(err)
	s.Assert().True(explained.Allowed)

	scopes, err := s.PermissionRepo.ListGrantScopes(s.ctx, actor.ID, model.ActionProjectRead)
	s.Require().NoError(err)
	s.Assert().Contains(scopes, ns.ID)
}

func (s *PermissionServiceIntegrationTestSuite) TestLinkInScopeOfCycle() {
	owner := s.createUser()
	actor := s.createUser()
	parent, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)
	child, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)

	s.Require().ErrorIs(s.permissionService.LinkInScopeOf(s.ctx, parent.ID, parent.ID), model.ErrGrantCycle)
	s.Require().NoError(s.permissionService.LinkInScopeOf(s.ctx, child.ID, parent.ID))
	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: actor.ID,
		Scope:     parent.ID,
		Actions:   []model.Action{model.ActionOrganizationRead},
	})
	s.Require().NoError(err)
	allowed, err := s.permissionService.Has(s.ctx, actor.ID, child.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().True(allowed)
	s.Require().ErrorIs(s.permissionService.LinkInScopeOf(s.ctx, parent.ID, child.ID), model.ErrGrantCycle)
}

func (s *PermissionServiceIntegrationTestSuite) TestOrgAsPrincipal() {
	ownerA := s.createUser()
	ownerB := s.createUser()
	memberA := s.createUser()
	orgA, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(ownerA.ID))
	s.Require().NoError(err)
	orgB, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(ownerB.ID))
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(s.ctx, orgA.ID, memberA.ID))
	nsB, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(ownerB.ID, orgB.ID))
	s.Require().NoError(err)
	projectB, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(nsB.ID, ownerB.ID))
	s.Require().NoError(err)

	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: orgA.ID,
		Scope:     projectB.ID,
		Actions:   []model.Action{model.ActionProjectRead},
	})
	s.Require().NoError(err)
	allowed, err := s.permissionService.Has(s.ctx, memberA.ID, projectB.ID, model.ActionProjectRead)
	s.Require().NoError(err)
	s.Assert().True(allowed)
}

func (s *PermissionServiceIntegrationTestSuite) TestCtxUserCreateRoleIDEscalation() {
	owner := s.createUser()
	actor := s.createUser()
	guest := s.createUser()
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner.ID))
	s.Require().NoError(err)

	_, err = s.permissionService.Create(s.ctx, service.CreateGrantOpts{
		Principal: actor.ID,
		Scope:     org.ID,
		Actions:   []model.Action{model.ActionPermissionManage, model.ActionOrganizationRead},
	})
	s.Require().NoError(err)

	tmpl, err := model.RoleTemplateByKey(model.RoleKeyOrgAdmin)
	s.Require().NoError(err)
	role, err := s.RoleRepo.Create(s.ctx, repository.CreateRoleOpts{
		Key:         tmpl.Key,
		Name:        tmpl.Name,
		Description: tmpl.Description,
		Actions:     tmpl.ActionStrings(),
		CreatedBy:   owner.ID,
		BelongsTo:   org.ID,
	})
	s.Require().NoError(err)

	_, err = s.permissionService.CtxUserCreate(s.userCtx(actor.ID), service.CreateGrantOpts{
		Principal: guest.ID,
		Scope:     org.ID,
		RoleID:    &role.ID,
	})
	s.Require().ErrorIs(err, model.ErrPrivilegeEscalation)
}

func TestPermissionServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionServiceIntegrationTestSuite))
}
