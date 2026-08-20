package repository_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type PermissionRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	ctx context.Context
}

func (s *PermissionRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *PermissionRepositoryIntegrationTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *PermissionRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *PermissionRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *PermissionRepositoryIntegrationTestSuite) createUser() *repository.User {
	user, err := s.UserRepo.Create(s.ctx, testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	return user
}

func (s *PermissionRepositoryIntegrationTestSuite) createOrg(owner model.ID) *repository.Organization {
	org, err := s.OrganizationRepo.Create(s.ctx, testModel.NewCreateOrganizationOpts(owner))
	s.Require().NoError(err)
	return org
}

func (s *PermissionRepositoryIntegrationTestSuite) grant(principal, scope model.ID, actions ...model.Action) *repository.Grant {
	grant, err := s.PermissionRepo.Create(s.ctx, repository.CreateGrantOpts{
		Principal: principal,
		Scope:     scope,
		Actions:   actions,
	})
	s.Require().NoError(err)
	return grant
}

func (s *PermissionRepositoryIntegrationTestSuite) has(actor, resource model.ID, action model.Action) bool {
	allowed, err := s.PermissionRepo.Has(s.ctx, actor, resource, action)
	s.Require().NoError(err)
	return allowed
}

func (s *PermissionRepositoryIntegrationTestSuite) exec(cypher string, params map[string]any) {
	s.Require().NoError(repository.Neo4jExecuteWriteAndConsume(s.ctx, s.Neo4jDB, cypher, params))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestDirectAllowDeny() {
	user := s.createUser()
	owner := s.createUser()
	org := s.createOrg(owner.ID)

	s.Assert().False(s.has(user.ID, org.ID, model.ActionOrganizationRead))

	s.grant(user.ID, org.ID, model.ActionOrganizationRead)
	s.Assert().True(s.has(user.ID, org.ID, model.ActionOrganizationRead))
	s.Assert().False(s.has(user.ID, org.ID, model.ActionOrganizationDelete))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestOrganizationCreateRequiresInstallationGrant() {
	user := s.createUser()
	s.Assert().False(s.has(user.ID, model.InstallationID(), model.ActionOrganizationCreate))

	s.Require().NoError(testRepo.GrantOrganizationCreate(user.ID, s.Neo4jDB))
	s.Assert().True(s.has(user.ID, model.InstallationID(), model.ActionOrganizationCreate))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestOrgAdminDoesNotIncludeOrganizationCreate() {
	owner := s.createUser()
	guest := s.createUser()
	org := s.createOrg(owner.ID)

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

	_, err = s.PermissionRepo.Create(s.ctx, repository.CreateGrantOpts{
		Principal: guest.ID,
		Scope:     org.ID,
		RoleID:    &role.ID,
	})
	s.Require().NoError(err)

	s.Assert().True(s.has(guest.ID, org.ID, model.ActionOrganizationRead))
	s.Assert().False(s.has(guest.ID, model.InstallationID(), model.ActionOrganizationCreate))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestAncestorGrantAppliesToDescendantsNotSiblings() {
	owner := s.createUser()
	actor := s.createUser()
	org := s.createOrg(owner.ID)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)
	siblingNS, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	siblingProject, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(siblingNS.ID, owner.ID))
	s.Require().NoError(err)

	s.grant(actor.ID, ns.ID, model.ActionProjectRead)
	s.Assert().True(s.has(actor.ID, project.ID, model.ActionProjectRead))
	s.Assert().False(s.has(actor.ID, siblingProject.ID, model.ActionProjectRead))
	s.Assert().False(s.has(actor.ID, org.ID, model.ActionOrganizationRead))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestChildGrantDoesNotGoUp() {
	owner := s.createUser()
	actor := s.createUser()
	org := s.createOrg(owner.ID)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)

	s.grant(actor.ID, project.ID, model.ActionProjectRead)
	s.Assert().True(s.has(actor.ID, project.ID, model.ActionProjectRead))
	s.Assert().False(s.has(actor.ID, ns.ID, model.ActionNamespaceRead))
	s.Assert().False(s.has(actor.ID, org.ID, model.ActionOrganizationRead))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestTeamGrantMemberJoinLeave() {
	owner := s.createUser()
	member := s.createUser()
	org := s.createOrg(owner.ID)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)
	team, err := s.TeamRepo.Create(s.ctx, repository.CreateTeamOpts{
		Name:      "rebac-team",
		CreatedBy: owner.ID,
		BelongsTo: org.ID,
	})
	s.Require().NoError(err)

	s.grant(team.ID, project.ID, model.ActionProjectRead)
	s.Assert().False(s.has(member.ID, project.ID, model.ActionProjectRead))

	s.Require().NoError(s.TeamRepo.AddMember(s.ctx, team.ID, member.ID, org.ID))
	s.Assert().True(s.has(member.ID, project.ID, model.ActionProjectRead))

	s.Require().NoError(s.TeamRepo.RemoveMember(s.ctx, team.ID, member.ID, org.ID))
	s.Assert().False(s.has(member.ID, project.ID, model.ActionProjectRead))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestOrgAsPrincipalOnForeignProject() {
	ownerA := s.createUser()
	ownerB := s.createUser()
	memberA := s.createUser()
	orgA := s.createOrg(ownerA.ID)
	orgB := s.createOrg(ownerB.ID)
	s.Require().NoError(s.OrganizationRepo.AddMember(s.ctx, orgA.ID, memberA.ID))

	nsB, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(ownerB.ID, orgB.ID))
	s.Require().NoError(err)
	projectB, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(nsB.ID, ownerB.ID))
	s.Require().NoError(err)

	grant := s.grant(orgA.ID, projectB.ID, model.ActionProjectRead)
	s.Assert().True(s.has(memberA.ID, projectB.ID, model.ActionProjectRead))
	s.Assert().False(s.has(memberA.ID, orgB.ID, model.ActionOrganizationRead))

	members, err := s.OrganizationRepo.ListMembers(s.ctx, orgB.ID, repository.CursorPage{Size: 20})
	s.Require().NoError(err)
	for _, m := range members.Items {
		s.Assert().NotEqual(memberA.ID, m.ID)
	}

	s.Require().NoError(s.PermissionRepo.Delete(s.ctx, grant.ID))
	s.Assert().False(s.has(memberA.ID, projectB.ID, model.ActionProjectRead))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestCreatorWithoutGrantIsDenied() {
	owner := s.createUser()
	org := s.createOrg(owner.ID)
	s.Assert().False(s.has(owner.ID, org.ID, model.ActionOrganizationRead))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestNonAuthzEdgesDoNotAuthorize() {
	owner := s.createUser()
	actor := s.createUser()
	org := s.createOrg(owner.ID)

	s.exec(`
		MATCH (u:`+actor.ID.Label()+` {id: $actor_id})
		MATCH (o:`+org.ID.Label()+` {id: $org_id})
		CREATE (u)-[:`+repository.EdgeKindAssignedTo.String()+` {id: $e1}]->(o)
		CREATE (u)-[:`+repository.EdgeKindRelatedTo.String()+` {id: $e2}]->(o)
		CREATE (u)-[:`+repository.EdgeKindHasLabel.String()+` {id: $e3}]->(o)
	`, map[string]any{
		"actor_id": actor.ID.String(),
		"org_id":   org.ID.String(),
		"e1":       model.NewRawID(),
		"e2":       model.NewRawID(),
		"e3":       model.NewRawID(),
	})

	s.Assert().False(s.has(actor.ID, org.ID, model.ActionOrganizationRead))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestRevokeGrantDenies() {
	owner := s.createUser()
	actor := s.createUser()
	org := s.createOrg(owner.ID)
	grant := s.grant(actor.ID, org.ID, model.ActionOrganizationRead)
	s.Assert().True(s.has(actor.ID, org.ID, model.ActionOrganizationRead))

	s.Require().NoError(s.PermissionRepo.Delete(s.ctx, grant.ID))
	s.Assert().False(s.has(actor.ID, org.ID, model.ActionOrganizationRead))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestDisableUserDenies() {
	owner := s.createUser()
	actor := s.createUser()
	org := s.createOrg(owner.ID)
	s.grant(actor.ID, org.ID, model.ActionOrganizationRead)
	s.Assert().True(s.has(actor.ID, org.ID, model.ActionOrganizationRead))

	_, err := s.UserRepo.Update(s.ctx, actor.ID, repository.UpdateUserOpts{
		Status: optional.Some(model.UserStatusInactive),
	})
	s.Require().NoError(err)
	s.Assert().False(s.has(actor.ID, org.ID, model.ActionOrganizationRead))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestListDoesNotLeakUngrantedOrgs() {
	owner := s.createUser()
	actor := s.createUser()
	visible := s.createOrg(owner.ID)
	hidden := s.createOrg(owner.ID)
	s.grant(actor.ID, visible.ID, model.ActionOrganizationRead)

	orgs, err := s.OrganizationRepo.List(s.ctx, actor.ID, repository.CursorPage{Size: 20}, repository.OrganizationListProjection())
	s.Require().NoError(err)
	ids := make([]model.ID, 0, len(orgs.Items))
	for _, org := range orgs.Items {
		ids = append(ids, org.ID)
	}
	s.Assert().Contains(ids, visible.ID)
	s.Assert().NotContains(ids, hidden.ID)

	visibleIDs, err := s.PermissionRepo.ListVisible(s.ctx, actor.ID, model.ActionOrganizationRead, model.InstallationID(), model.ResourceTypeOrganization)
	s.Require().NoError(err)
	s.Assert().Contains(visibleIDs, visible.ID)
	s.Assert().NotContains(visibleIDs, hidden.ID)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestIDORHasWithoutGrant() {
	owner := s.createUser()
	stranger := s.createUser()
	org := s.createOrg(owner.ID)
	s.grant(owner.ID, org.ID, model.ActionOrganizationRead)
	s.Assert().False(s.has(stranger.ID, org.ID, model.ActionOrganizationRead))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestGrantCRUD() {
	owner := s.createUser()
	actor := s.createUser()
	other := s.createUser()
	org := s.createOrg(owner.ID)
	otherOrg := s.createOrg(owner.ID)

	_, err := s.PermissionRepo.Get(s.ctx, model.MustNewID(model.ResourceTypePermission))
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	empty, err := s.PermissionRepo.ListByPrincipal(s.ctx, actor.ID)
	s.Require().NoError(err)
	s.Assert().Empty(empty)

	grant := s.grant(actor.ID, org.ID, model.ActionOrganizationRead, model.ActionOrganizationUpdate)
	got, err := s.PermissionRepo.Get(s.ctx, grant.ID)
	s.Require().NoError(err)
	s.Assert().Equal(grant.ID, got.ID)
	s.Assert().Equal(actor.ID, got.Principal)
	s.Assert().Equal(org.ID, got.Scope)

	s.grant(actor.ID, otherOrg.ID, model.ActionOrganizationRead)
	s.grant(other.ID, org.ID, model.ActionOrganizationRead)

	byPrincipal, err := s.PermissionRepo.ListByPrincipal(s.ctx, actor.ID)
	s.Require().NoError(err)
	s.Assert().Len(byPrincipal, 2)

	byScope, err := s.PermissionRepo.ListByScope(s.ctx, org.ID)
	s.Require().NoError(err)
	s.Assert().Len(byScope, 2)

	s.Require().NoError(s.PermissionRepo.Delete(s.ctx, grant.ID))
	s.Assert().False(s.has(actor.ID, org.ID, model.ActionOrganizationRead))
	_, err = s.PermissionRepo.Get(s.ctx, grant.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestEffectiveActions() {
	owner := s.createUser()
	actor := s.createUser()
	inactive := s.createUser()
	org := s.createOrg(owner.ID)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)

	empty, err := s.PermissionRepo.EffectiveActions(s.ctx, actor.ID, org.ID)
	s.Require().NoError(err)
	s.Assert().Empty(empty)

	s.grant(actor.ID, org.ID, model.ActionOrganizationRead, model.ActionOrganizationUpdate)
	direct, err := s.PermissionRepo.EffectiveActions(s.ctx, actor.ID, org.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.Action{model.ActionOrganizationRead, model.ActionOrganizationUpdate}, direct)

	s.grant(actor.ID, ns.ID, model.ActionProjectRead)
	inherited, err := s.PermissionRepo.EffectiveActions(s.ctx, actor.ID, project.ID)
	s.Require().NoError(err)
	s.Assert().Contains(inherited, model.ActionProjectRead)

	team, err := s.TeamRepo.Create(s.ctx, repository.CreateTeamOpts{Name: "actions-team", CreatedBy: owner.ID, BelongsTo: org.ID})
	s.Require().NoError(err)
	s.grant(team.ID, project.ID, model.ActionProjectUpdate)
	s.Require().NoError(s.TeamRepo.AddMember(s.ctx, team.ID, actor.ID, org.ID))
	withTeam, err := s.PermissionRepo.EffectiveActions(s.ctx, actor.ID, project.ID)
	s.Require().NoError(err)
	s.Assert().Contains(withTeam, model.ActionProjectRead)
	s.Assert().Contains(withTeam, model.ActionProjectUpdate)

	ownerB := s.createUser()
	orgB := s.createOrg(ownerB.ID)
	nsB, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(ownerB.ID, orgB.ID))
	s.Require().NoError(err)
	projectB, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(nsB.ID, ownerB.ID))
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(s.ctx, org.ID, actor.ID))
	s.grant(org.ID, projectB.ID, model.ActionProjectRead)
	orgPrincipal, err := s.PermissionRepo.EffectiveActions(s.ctx, actor.ID, projectB.ID)
	s.Require().NoError(err)
	s.Assert().Contains(orgPrincipal, model.ActionProjectRead)

	tmpl, err := model.RoleTemplateByKey(model.RoleKeyOrgMember)
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
	guest := s.createUser()
	_, err = s.PermissionRepo.Create(s.ctx, repository.CreateGrantOpts{
		Principal: guest.ID,
		Scope:     org.ID,
		RoleID:    &role.ID,
		Actions:   []model.Action{model.ActionOrganizationDelete},
	})
	s.Require().NoError(err)
	union, err := s.PermissionRepo.EffectiveActions(s.ctx, guest.ID, org.ID)
	s.Require().NoError(err)
	s.Assert().Contains(union, model.ActionOrganizationRead)
	s.Assert().Contains(union, model.ActionOrganizationDelete)

	s.grant(inactive.ID, org.ID, model.ActionOrganizationRead)
	_, err = s.UserRepo.Update(s.ctx, inactive.ID, repository.UpdateUserOpts{Status: optional.Some(model.UserStatusInactive)})
	s.Require().NoError(err)
	inactiveActions, err := s.PermissionRepo.EffectiveActions(s.ctx, inactive.ID, org.ID)
	s.Require().NoError(err)
	s.Assert().Empty(inactiveActions)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestExplain() {
	owner := s.createUser()
	actor := s.createUser()
	org := s.createOrg(owner.ID)

	denied, err := s.PermissionRepo.Explain(s.ctx, actor.ID, org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().False(denied.Allowed)
	s.Assert().Nil(denied.Principal)
	s.Assert().Nil(denied.Scope)
	s.Assert().Nil(denied.GrantID)

	grant := s.grant(actor.ID, org.ID, model.ActionOrganizationRead)
	allowed, err := s.PermissionRepo.Explain(s.ctx, actor.ID, org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().True(allowed.Allowed)
	s.Require().NotNil(allowed.Principal)
	s.Assert().Equal(actor.ID, *allowed.Principal)
	s.Require().NotNil(allowed.Scope)
	s.Assert().Equal(org.ID, *allowed.Scope)
	s.Require().NotNil(allowed.GrantID)
	s.Assert().Equal(grant.ID, *allowed.GrantID)
	s.Assert().Nil(allowed.RoleID)

	tmpl, err := model.RoleTemplateByKey(model.RoleKeyOrgMember)
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
	rolePrincipal := s.createUser()
	roleGrant, err := s.PermissionRepo.Create(s.ctx, repository.CreateGrantOpts{
		Principal: rolePrincipal.ID,
		Scope:     org.ID,
		RoleID:    &role.ID,
	})
	s.Require().NoError(err)
	roleAllowed, err := s.PermissionRepo.Explain(s.ctx, rolePrincipal.ID, org.ID, model.ActionOrganizationRead)
	s.Require().NoError(err)
	s.Assert().True(roleAllowed.Allowed)
	s.Require().NotNil(roleAllowed.RoleID)
	s.Assert().Equal(role.ID, *roleAllowed.RoleID)
	s.Require().NotNil(roleAllowed.GrantID)
	s.Assert().Equal(roleGrant.ID, *roleAllowed.GrantID)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestListVisibleUnderParent() {
	owner := s.createUser()
	actor := s.createUser()
	org := s.createOrg(owner.ID)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	visible, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)
	hidden, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)

	empty, err := s.PermissionRepo.ListVisible(s.ctx, actor.ID, model.ActionProjectRead, ns.ID, model.ResourceTypeProject)
	s.Require().NoError(err)
	s.Assert().Empty(empty)

	s.grant(actor.ID, visible.ID, model.ActionProjectRead)
	ids, err := s.PermissionRepo.ListVisible(s.ctx, actor.ID, model.ActionProjectRead, ns.ID, model.ResourceTypeProject)
	s.Require().NoError(err)
	s.Assert().Contains(ids, visible.ID)
	s.Assert().NotContains(ids, hidden.ID)

	updateIDs, err := s.PermissionRepo.ListVisible(s.ctx, actor.ID, model.ActionProjectUpdate, ns.ID, model.ResourceTypeProject)
	s.Require().NoError(err)
	s.Assert().NotContains(updateIDs, visible.ID)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestLinkInScopeOfAndCycle() {
	owner := s.createUser()
	actor := s.createUser()
	parent := s.createOrg(owner.ID)
	child := s.createOrg(owner.ID)

	s.Require().ErrorIs(s.PermissionRepo.LinkInScopeOf(s.ctx, parent.ID, parent.ID), model.ErrGrantCycle)

	s.Require().NoError(s.PermissionRepo.LinkInScopeOf(s.ctx, child.ID, parent.ID))
	s.grant(actor.ID, parent.ID, model.ActionOrganizationRead)
	s.Assert().True(s.has(actor.ID, child.ID, model.ActionOrganizationRead))

	s.Require().ErrorIs(s.PermissionRepo.LinkInScopeOf(s.ctx, parent.ID, child.ID), model.ErrGrantCycle)
}

func (s *PermissionRepositoryIntegrationTestSuite) TestMemberOfDepthDoesNotAuthorize() {
	owner := s.createUser()
	actor := s.createUser()
	org := s.createOrg(owner.ID)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)
	inner, err := s.TeamRepo.Create(s.ctx, repository.CreateTeamOpts{Name: "inner", CreatedBy: owner.ID, BelongsTo: org.ID})
	s.Require().NoError(err)
	outer, err := s.TeamRepo.Create(s.ctx, repository.CreateTeamOpts{Name: "outer", CreatedBy: owner.ID, BelongsTo: org.ID})
	s.Require().NoError(err)

	s.Require().NoError(s.TeamRepo.AddMember(s.ctx, inner.ID, actor.ID, org.ID))
	s.exec(`
		MATCH (inner:`+inner.ID.Label()+` {id: $inner_id})
		MATCH (outer:`+outer.ID.Label()+` {id: $outer_id})
		CREATE (inner)-[:`+repository.EdgeKindMemberOf.String()+` {id: $rel_id}]->(outer)
	`, map[string]any{
		"inner_id": inner.ID.String(),
		"outer_id": outer.ID.String(),
		"rel_id":   model.NewRawID(),
	})
	s.grant(outer.ID, project.ID, model.ActionProjectRead)
	s.Assert().False(s.has(actor.ID, project.ID, model.ActionProjectRead))
	s.Assert().True(s.has(outer.ID, project.ID, model.ActionProjectRead))
}

func (s *PermissionRepositoryIntegrationTestSuite) TestListGrantScopesAndAncestry() {
	owner := s.createUser()
	actor := s.createUser()
	org := s.createOrg(owner.ID)
	ns, err := s.NamespaceRepo.Create(s.ctx, testModel.NewCreateNamespaceOpts(owner.ID, org.ID))
	s.Require().NoError(err)
	project, err := s.ProjectRepo.Create(s.ctx, testModel.NewCreateProjectOpts(ns.ID, owner.ID))
	s.Require().NoError(err)

	s.grant(actor.ID, project.ID, model.ActionIssueRead)
	scopes, err := s.PermissionRepo.ListGrantScopes(s.ctx, actor.ID, model.ActionIssueRead)
	s.Require().NoError(err)
	s.Require().Len(scopes, 1)
	s.Assert().Equal(project.ID, scopes[0])

	ancestry, err := s.PermissionRepo.ListScopeAncestry(s.ctx, project.ID)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(ancestry), 1)
	s.Assert().Equal(project.ID, ancestry[0])
}

func TestPermissionRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionRepositoryIntegrationTestSuite))
}
