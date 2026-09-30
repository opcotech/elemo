package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
)

type denyMutations struct{}

func (denyMutations) AllowsMutation(_ context.Context) error {
	return entitlement.ErrMutationDenied
}

func deniedOpts() service.Option {
	return service.WithMutationPolicy(denyMutations{})
}

func requireMutationDenied(t *testing.T, err error) {
	t.Helper()
	require.ErrorIs(t, err, entitlement.ErrMutationDenied)
}

func TestDomainMutationsDeniedBeforeSideEffects(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	projectID := model.MustNewID(model.ResourceTypeProject)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	memberID := model.MustNewID(model.ResourceTypeUser)
	roleID := model.MustNewID(model.ResourceTypeRole)
	teamID := model.MustNewID(model.ResourceTypeTeam)
	folderID := model.MustNewID(model.ResourceTypeFolder)
	definitionID := model.MustNewID(model.ResourceTypeCustomFieldDefinition)
	grantID := model.MustNewID(model.ResourceTypePermission)
	relationID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("document", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewDocumentService(
			mockrepo.NewMockDocumentRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			mocksvc.NewMockStaticFileService(ctrl),
			mocksvc.NewMockSearchService(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, orgID, service.CreateDocumentOpts{})
		requireMutationDenied(t, err)
		_, err = svc.Update(ctx, id, service.UpdateDocumentOpts{})
		requireMutationDenied(t, err)
		_, err = svc.MoveLibrary(ctx, id, orgID)
		requireMutationDenied(t, err)
		_, err = svc.MoveToFolder(ctx, id, &folderID)
		requireMutationDenied(t, err)
		err = svc.Relate(ctx, id, issueID)
		requireMutationDenied(t, err)
		err = svc.Unrelate(ctx, id, issueID)
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, id)
		requireMutationDenied(t, err)
	})

	t.Run("issue", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewIssueService(
			mockrepo.NewMockIssueRepository(ctrl),
			mockrepo.NewMockAssignmentRepository(ctrl),
			mockrepo.NewMockLabelRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			mocksvc.NewMockSearchService(ctrl),
			mocksvc.NewMockCustomFieldService(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, projectID, service.CreateIssueOpts{})
		requireMutationDenied(t, err)
		_, err = svc.Update(ctx, issueID, service.UpdateIssueOpts{})
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, issueID)
		requireMutationDenied(t, err)
		_, err = svc.AddRelation(ctx, issueID, relationID, model.IssueRelationKindBlocks)
		requireMutationDenied(t, err)
		_, err = svc.UpdateRelation(ctx, issueID, relationID, model.IssueRelationKindBlocks)
		requireMutationDenied(t, err)
		err = svc.RemoveRelation(ctx, issueID, relationID)
		requireMutationDenied(t, err)
	})

	t.Run("organization", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewOrganizationService(
			mockrepo.NewMockOrganizationRepository(ctrl),
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			mockrepo.NewMockRoleRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			entitlement.Unrestricted(),
			mocksvc.NewMockEmailService(ctrl),
			mocksvc.NewMockNotificationService(ctrl),
			mocksvc.NewMockSearchService(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, id, service.CreateOrganizationOpts{})
		requireMutationDenied(t, err)
		_, err = svc.Update(ctx, orgID, service.UpdateOrganizationOpts{})
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, orgID, false)
		requireMutationDenied(t, err)
		err = svc.AddMember(ctx, orgID, memberID)
		requireMutationDenied(t, err)
		err = svc.RemoveMember(ctx, orgID, memberID)
		requireMutationDenied(t, err)
		err = svc.InviteMember(ctx, orgID, service.InviteOrganizationMemberOpts{})
		requireMutationDenied(t, err)
		err = svc.RevokeInvitation(ctx, orgID, memberID)
		requireMutationDenied(t, err)
		err = svc.AcceptInvitation(ctx, orgID, service.AcceptOrganizationInvitationOpts{})
		requireMutationDenied(t, err)
	})

	t.Run("namespace", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewNamespaceService(
			mockrepo.NewMockNamespaceRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			mocksvc.NewMockSearchService(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, orgID, service.CreateNamespaceOpts{})
		requireMutationDenied(t, err)
		_, err = svc.Update(ctx, namespaceID, service.UpdateNamespaceOpts{})
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, namespaceID)
		requireMutationDenied(t, err)
	})

	t.Run("project", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewProjectService(
			mockrepo.NewMockProjectRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			mocksvc.NewMockSearchService(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, namespaceID, service.CreateProjectOpts{})
		requireMutationDenied(t, err)
		_, err = svc.Update(ctx, projectID, service.UpdateProjectOpts{})
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, projectID)
		requireMutationDenied(t, err)
	})

	t.Run("folder", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewFolderService(
			mockrepo.NewMockFolderRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, orgID, service.CreateFolderOpts{})
		requireMutationDenied(t, err)
		_, err = svc.Update(ctx, folderID, service.UpdateFolderOpts{})
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, folderID)
		requireMutationDenied(t, err)
	})

	t.Run("role", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewRoleService(
			mockrepo.NewMockRoleRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			mockrepo.NewMockOrganizationRepository(ctrl),
			mocksvc.NewMockNotificationService(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, id, orgID, service.CreateRoleOpts{})
		requireMutationDenied(t, err)
		_, err = svc.Update(ctx, roleID, orgID, service.UpdateRoleOpts{})
		requireMutationDenied(t, err)
		err = svc.AddMember(ctx, roleID, memberID, orgID)
		requireMutationDenied(t, err)
		err = svc.RemoveMember(ctx, roleID, memberID, orgID)
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, roleID, orgID)
		requireMutationDenied(t, err)
	})

	t.Run("team", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewTeamService(
			mockrepo.NewMockTeamRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, orgID, service.CreateTeamOpts{})
		requireMutationDenied(t, err)
		_, err = svc.Update(ctx, teamID, orgID, service.UpdateTeamOpts{})
		requireMutationDenied(t, err)
		err = svc.AddMember(ctx, teamID, memberID, orgID)
		requireMutationDenied(t, err)
		err = svc.RemoveMember(ctx, teamID, memberID, orgID)
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, teamID, orgID)
		requireMutationDenied(t, err)
	})

	t.Run("user create", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewUserService(
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			entitlement.Unrestricted(),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, service.CreateUserOpts{})
		requireMutationDenied(t, err)
	})

	t.Run("todo", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewTodoService(mockrepo.NewMockTodoRepository(ctrl), deniedOpts())
		require.NoError(t, err)
		_, err = svc.Create(ctx, service.CreateTodoOpts{})
		requireMutationDenied(t, err)
		_, err = svc.Update(ctx, id, service.UpdateTodoOpts{})
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, id)
		requireMutationDenied(t, err)
	})

	t.Run("notification", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewNotificationService(mockrepo.NewMockNotificationRepository(ctrl), deniedOpts())
		require.NoError(t, err)
		_, err = svc.Create(ctx, service.CreateNotificationOpts{})
		requireMutationDenied(t, err)
		_, err = svc.Update(ctx, id, memberID, service.UpdateNotificationOpts{})
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, id, memberID)
		requireMutationDenied(t, err)
	})

	t.Run("custom field", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewCustomFieldService(
			mockrepo.NewMockCustomFieldRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.CreateDefinition(ctx, service.CreateCustomFieldOpts{})
		requireMutationDenied(t, err)
		_, err = svc.UpdateDefinition(ctx, definitionID, service.UpdateCustomFieldOpts{})
		requireMutationDenied(t, err)
		_, err = svc.ArchiveDefinition(ctx, definitionID)
		requireMutationDenied(t, err)
		err = svc.DeleteDefinition(ctx, definitionID)
		requireMutationDenied(t, err)
		err = svc.SetValue(ctx, issueID, definitionID, model.CustomFieldTypedValue{})
		requireMutationDenied(t, err)
		err = svc.DeleteValue(ctx, issueID, definitionID)
		requireMutationDenied(t, err)
		err = svc.StageForResource(ctx, orgID, issueID, nil)
		requireMutationDenied(t, err)
		err = svc.CommitForResource(ctx, issueID)
		requireMutationDenied(t, err)
		err = svc.AbortForResource(ctx, issueID)
		requireMutationDenied(t, err)
		err = svc.DeleteForResource(ctx, issueID)
		requireMutationDenied(t, err)
	})

	t.Run("permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewPermissionService(
			mockrepo.NewMockPermissionRepository(ctrl),
			mockrepo.NewMockRoleRepository(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, service.CreateGrantOpts{})
		requireMutationDenied(t, err)
		_, err = svc.CtxUserCreate(ctx, service.CreateGrantOpts{})
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, grantID)
		requireMutationDenied(t, err)
		err = svc.CtxUserDelete(ctx, grantID)
		requireMutationDenied(t, err)
		err = svc.LinkInScopeOf(ctx, projectID, namespaceID)
		requireMutationDenied(t, err)
		err = svc.BootstrapCreator(ctx, id, orgID, nil)
		requireMutationDenied(t, err)
		err = svc.GrantRole(ctx, id, orgID, roleID)
		requireMutationDenied(t, err)
	})

	t.Run("static file", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewStaticFileService(mockrepo.NewMockStaticFileRepository(ctrl), deniedOpts())
		require.NoError(t, err)
		err = svc.Create(ctx, "file.txt", []byte("x"))
		requireMutationDenied(t, err)
		err = svc.Update(ctx, "file.txt", []byte("y"))
		requireMutationDenied(t, err)
		err = svc.Delete(ctx, "file.txt")
		requireMutationDenied(t, err)
	})

	t.Run("plugin", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewPluginService(
			config.PluginConfig{Directory: t.TempDir()},
			mockrepo.NewMockPluginRepository(ctrl),
			mockrepo.NewMockExtensionRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			mocksvc.NewMockIssueService(ctrl),
			mocksvc.NewMockProjectService(ctrl),
			mocksvc.NewMockUserService(ctrl),
			nil,
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Install(ctx, []byte("zip"))
		requireMutationDenied(t, err)
		_, err = svc.Upgrade(ctx, "plugin", []byte("zip"))
		requireMutationDenied(t, err)
		err = svc.Uninstall(ctx, "plugin")
		requireMutationDenied(t, err)
		err = svc.Enable(ctx, "plugin", orgID, nil)
		requireMutationDenied(t, err)
		err = svc.Disable(ctx, "plugin", orgID)
		requireMutationDenied(t, err)
		err = svc.SetConfig(ctx, "plugin", orgID, nil)
		requireMutationDenied(t, err)
		_, err = svc.Invoke(ctx, "plugin", elemoplugin.InvokeRequest{})
		requireMutationDenied(t, err)
		_, err = svc.CreateNode(ctx, "plugin", service.CreateExtensionNodeOpts{Parent: issueID})
		requireMutationDenied(t, err)
		_, err = svc.UpdateNode(ctx, "plugin", issueID, nil)
		requireMutationDenied(t, err)
		err = svc.DeleteNode(ctx, "plugin", issueID)
		requireMutationDenied(t, err)
		_, err = svc.MoveNode(ctx, "plugin", issueID, projectID)
		requireMutationDenied(t, err)
		_, err = svc.CreateRelation(ctx, "plugin", service.CreateExtensionRelationOpts{
			From: issueID,
			To:   projectID,
		})
		requireMutationDenied(t, err)
		err = svc.DeleteRelation(ctx, "plugin", "rel")
		requireMutationDenied(t, err)
	})
}

func TestUserUpdateReadOnlyExceptions(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	t.Run("profile update is denied", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewUserService(
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			entitlement.Unrestricted(),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Update(ctx, userID, service.UpdateUserOpts{FirstName: optional.Some("Ada")})
		require.ErrorIs(t, err, entitlement.ErrMutationDenied)
	})

	t.Run("mixed password and profile update is denied", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewUserService(
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			entitlement.Unrestricted(),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Update(ctx, userID, service.UpdateUserOpts{
			Password: optional.Some("secret"),
			Bio:      optional.Some("nope"),
		})
		require.ErrorIs(t, err, entitlement.ErrMutationDenied)
	})

	t.Run("password-only update is allowed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockUserRepository(ctrl)
		repo.EXPECT().Update(gomock.Any(), userID, gomock.Any()).Return(&repository.User{ID: userID}, nil)
		svc, err := service.NewUserService(
			repo,
			mockrepo.NewMockUserTokenRepository(ctrl),
			entitlement.Unrestricted(),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Update(ctx, userID, service.UpdateUserOpts{Password: optional.Some("secret")})
		require.NoError(t, err)
	})

	t.Run("deactivation-only update is allowed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockUserRepository(ctrl)
		repo.EXPECT().Update(gomock.Any(), userID, gomock.Any()).Return(&repository.User{ID: userID, Status: model.UserStatusInactive}, nil)
		svc, err := service.NewUserService(
			repo,
			mockrepo.NewMockUserTokenRepository(ctrl),
			entitlement.Unrestricted(),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Update(ctx, userID, service.UpdateUserOpts{Status: optional.Some(model.UserStatusInactive)})
		require.NoError(t, err)
	})

	t.Run("user delete is allowed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockUserRepository(ctrl)
		repo.EXPECT().Update(gomock.Any(), userID, gomock.Any()).Return(&repository.User{ID: userID, Status: model.UserStatusDeleted}, nil)
		svc, err := service.NewUserService(
			repo,
			mockrepo.NewMockUserTokenRepository(ctrl),
			entitlement.Unrestricted(),
			deniedOpts(),
		)
		require.NoError(t, err)
		require.NoError(t, svc.Delete(ctx, userID, false))
	})

	t.Run("forced user delete is denied", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewUserService(
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			entitlement.Unrestricted(),
			deniedOpts(),
		)
		require.NoError(t, err)
		require.ErrorIs(t, svc.Delete(ctx, userID, true), entitlement.ErrMutationDenied)
	})

	t.Run("pending status update is denied", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewUserService(
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			entitlement.Unrestricted(),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Update(ctx, userID, service.UpdateUserOpts{Status: optional.Some(model.UserStatusPending)})
		require.ErrorIs(t, err, entitlement.ErrMutationDenied)
	})

	t.Run("active status update is denied", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		svc, err := service.NewUserService(
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			entitlement.Unrestricted(),
			deniedOpts(),
		)
		require.NoError(t, err)
		_, err = svc.Update(ctx, userID, service.UpdateUserOpts{Status: optional.Some(model.UserStatusActive)})
		require.ErrorIs(t, err, entitlement.ErrMutationDenied)
	})
}

func TestUserUpdateIsReadOnlyExempt(t *testing.T) {
	t.Parallel()
	require.False(t, service.UserUpdateIsReadOnlyExempt(service.UpdateUserOpts{}))
	require.True(t, service.UserUpdateIsReadOnlyExempt(service.UpdateUserOpts{Password: optional.Some("x")}))
	require.True(t, service.UserUpdateIsReadOnlyExempt(service.UpdateUserOpts{Status: optional.Some(model.UserStatusDeleted)}))
	require.True(t, service.UserUpdateIsReadOnlyExempt(service.UpdateUserOpts{Status: optional.Some(model.UserStatusInactive)}))
	require.False(t, service.UserUpdateIsReadOnlyExempt(service.UpdateUserOpts{Status: optional.Some(model.UserStatusActive)}))
	require.False(t, service.UserUpdateIsReadOnlyExempt(service.UpdateUserOpts{Status: optional.Some(model.UserStatusPending)}))
	require.False(t, service.UserUpdateIsReadOnlyExempt(service.UpdateUserOpts{
		Password: optional.Some("x"),
		Email:    optional.Some("a@example.com"),
	}))
}

func TestReadOnlyOperationalExceptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("custom field reconcile remains available", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListPendingOperations(gomock.Any(), gomock.Any(), 100).Return(nil, nil)
		svc, err := service.NewCustomFieldService(
			repo,
			mocksvc.NewMockPermissionService(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		require.NoError(t, svc.ReconcilePending(ctx))
	})

	t.Run("plugin restore remains available", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockPluginRepository(ctrl)
		repo.EXPECT().ListInstallations(gomock.Any()).Return(nil, nil)
		svc, err := service.NewPluginService(
			config.PluginConfig{Directory: t.TempDir()},
			repo,
			mockrepo.NewMockExtensionRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			mocksvc.NewMockIssueService(ctrl),
			mocksvc.NewMockProjectService(ctrl),
			mocksvc.NewMockUserService(ctrl),
			nil,
			deniedOpts(),
		)
		require.NoError(t, err)
		require.NoError(t, svc.Restore(ctx))
	})

	t.Run("search index remains available", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		id := model.MustNewID(model.ResourceTypeIssue)
		perm := mocksvc.NewMockPermissionService(ctrl)
		perm.EXPECT().ListScopeAncestry(gomock.Any(), id).Return([]model.ID{id}, nil)
		searchRepo := mockrepo.NewMockSearchRepository(ctrl)
		searchRepo.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil)
		svc, err := service.NewSearchService(
			searchRepo,
			perm,
			nil,
			deniedOpts(),
		)
		require.NoError(t, err)
		require.NoError(t, svc.Index(ctx, service.IndexInput{ID: id, Title: "issue"}))
	})

	t.Run("permission generation bump remains available", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		repo := mockrepo.NewMockPermissionRepository(ctrl)
		principal := model.MustNewID(model.ResourceTypeUser)
		repo.EXPECT().BumpGeneration(gomock.Any(), principal).Return(nil)
		svc, err := service.NewPermissionService(
			repo,
			mockrepo.NewMockRoleRepository(ctrl),
			deniedOpts(),
		)
		require.NoError(t, err)
		require.NoError(t, svc.BumpGeneration(ctx, principal))
	})
}

type oneShotMutations struct {
	calls int
}

func (c *oneShotMutations) AllowsMutation(_ context.Context) error {
	c.calls++
	if c.calls > 1 {
		return entitlement.ErrMutationDenied
	}
	return nil
}

func TestNestedMutationReusesOuterDecision(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	userID := model.MustNewID(model.ResourceTypeUser)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	projectID := model.MustNewID(model.ResourceTypeProject)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
	policy := &oneShotMutations{}

	permRepo := mockrepo.NewMockPermissionRepository(ctrl)
	permRepo.EXPECT().Has(gomock.Any(), userID, namespaceID, model.ActionProjectCreate).Return(true, nil)
	permRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&repository.Grant{
		ID:        model.MustNewID(model.ResourceTypePermission),
		Principal: userID,
		Scope:     projectID,
	}, nil)
	permRepo.EXPECT().BumpGeneration(gomock.Any(), userID).Return(nil)

	perm, err := service.NewPermissionService(
		permRepo,
		mockrepo.NewMockRoleRepository(ctrl),
		service.WithMutationPolicy(policy),
	)
	require.NoError(t, err)

	projectRepo := mockrepo.NewMockProjectRepository(ctrl)
	projectRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&repository.Project{
		ID:   projectID,
		Key:  "ABC",
		Name: "Alpha Project",
	}, nil)

	search := mocksvc.NewMockSearchService(ctrl)
	search.EXPECT().EnqueueIndex(gomock.Any(), projectID).Return(nil)

	svc, err := service.NewProjectService(
		projectRepo,
		perm,
		search,
		service.WithMutationPolicy(policy),
	)
	require.NoError(t, err)

	got, err := svc.Create(ctx, namespaceID, service.CreateProjectOpts{Key: "ABC", Name: "Alpha Project"})
	require.NoError(t, err)
	require.Equal(t, projectID, got.ID)
	require.Equal(t, 1, policy.calls)
}
