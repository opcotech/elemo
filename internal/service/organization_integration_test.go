package service_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type OrganizationServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	organizationService service.OrganizationService
	emailService        service.EmailService
	emailSender         *mock.EmailSender

	owner        *model.User
	organization *model.Organization

	ctx            context.Context
	capturedTokens map[string]string
}

func (s *OrganizationServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)

	permissionService, err := service.NewPermissionService(s.PermissionRepo)
	s.Require().NoError(err)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		service.WithPermissionService(permissionService),
	)
	s.Require().NoError(err)

	// Create a mock email sender for integration tests
	ctrl := gomock.NewController(s.T())
	s.emailSender = mock.NewEmailSender(ctrl)
	s.capturedTokens = make(map[string]string)
	// Capture tokens from invitation emails
	s.emailSender.EXPECT().SendEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Do(func(_ context.Context, _, to string, template any) {
			// Extract token from the template data if it's an organization invitation
			if tmpl, ok := template.(*email.Template); ok && tmpl != nil {
				if data, ok := tmpl.Data.Get().(*email.OrganizationInviteTemplateData); ok && data != nil {
					// Extract token from invitation URL
					// URL format: http://localhost:3000/organizations/join?organization=ORG_ID&token=TOKEN
					if strings.Contains(data.InvitationURL, "token=") {
						parts := strings.Split(data.InvitationURL, "token=")
						if len(parts) > 1 {
							s.capturedTokens[to] = parts[1]
						}
					}
				}
			}
		}).
		Return(nil)

	// Create a real EmailService with mock sender
	smtpConf := &config.SMTPConfig{
		ClientURL:      "http://localhost:3000",
		SupportAddress: "support@example.com",
	}
	s.emailService, err = service.NewEmailService(s.emailSender, "templates", smtpConf)
	s.Require().NoError(err)

	s.organizationService, err = service.NewOrganizationService(
		service.WithUserRepository(s.UserRepo),
		service.WithOrganizationRepository(s.OrganizationRepo),
		service.WithRoleRepository(s.RoleRepo),
		service.WithPermissionService(permissionService),
		service.WithLicenseService(licenseService),
		service.WithUserTokenRepository(s.UserTokenRepository),
		service.WithEmailService(s.emailService),
	)
	s.Require().NoError(err)
}

func (s *OrganizationServiceIntegrationTestSuite) SetupTest() {
	s.owner = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.owner))

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.owner.ID, s.Neo4jDB))

	s.organization = testModel.NewOrganization()
	s.capturedTokens = make(map[string]string)
}

func (s *OrganizationServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *OrganizationServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *OrganizationServiceIntegrationTestSuite) TestCreate() {
	err := s.organizationService.Create(s.ctx, s.owner.ID, s.organization)
	s.Require().NoError(err)
	s.Require().NotEmpty(s.organization.ID)
	s.Assert().NotNil(s.organization.CreatedAt)
	s.Assert().Nil(s.organization.UpdatedAt)
}

func (s *OrganizationServiceIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Equal(s.organization.ID, organization.ID)
	s.Assert().Equal(s.organization.Name, organization.Name)
	s.Assert().Equal(s.organization.Logo, organization.Logo)
	s.Assert().Equal(s.organization.Website, organization.Website)
	s.Assert().Equal(s.organization.Status, organization.Status)
	s.Assert().ElementsMatch(s.organization.Namespaces, organization.Namespaces)
	s.Assert().ElementsMatch(s.organization.Teams, organization.Teams)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, organization.Members)
	s.Assert().Equal(s.organization.CreatedAt, organization.CreatedAt)
	s.Assert().Equal(s.organization.UpdatedAt, organization.UpdatedAt)
}

func (s *OrganizationServiceIntegrationTestSuite) TestGetAll() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, testModel.NewOrganization()))
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, testModel.NewOrganization()))
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, testModel.NewOrganization()))

	organizations, err := s.organizationService.GetAll(s.ctx, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(organizations, 3)

	organizations, err = s.organizationService.GetAll(s.ctx, 0, 2)
	s.Require().NoError(err)
	s.Assert().Len(organizations, 2)

	organizations, err = s.organizationService.GetAll(s.ctx, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(organizations, 2)

	organizations, err = s.organizationService.GetAll(s.ctx, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(organizations, 1)

	organizations, err = s.organizationService.GetAll(s.ctx, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(organizations, 0)
}

func (s *OrganizationServiceIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	patch := map[string]any{
		"name": "new name",
		"logo": "https://example.com/static/new-logo.png",
	}

	organization, err := s.organizationService.Update(s.ctx, s.organization.ID, patch)
	s.Require().NoError(err)
	s.Assert().Equal(patch["name"], organization.Name)
	s.Assert().Equal(patch["logo"], organization.Logo)
	s.Assert().Equal(s.organization.Website, organization.Website)
	s.Assert().Equal(s.organization.Status, organization.Status)
	s.Assert().ElementsMatch(s.organization.Namespaces, organization.Namespaces)
	s.Assert().ElementsMatch(s.organization.Teams, organization.Teams)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, organization.Members)
	s.Assert().Equal(s.organization.CreatedAt, organization.CreatedAt)
	s.Assert().NotNil(organization.UpdatedAt)
}

func (s *OrganizationServiceIntegrationTestSuite) TestAddMember() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, organization.Members)

	member := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), member))

	err = s.organizationService.AddMember(s.ctx, s.organization.ID, member.ID)
	s.Require().NoError(err)

	organization, err = s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID, member.ID}, organization.Members)
}

func (s *OrganizationServiceIntegrationTestSuite) TestGetMembers() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	members, err := s.organizationService.GetMembers(s.ctx, s.organization.ID)
	s.Require().NoError(err)

	memberIDs := make([]model.ID, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}

	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, memberIDs)
	s.Assert().Len(members, 1)

	// Owner should have roles (includes virtual roles based on permissions)
	s.Assert().NotNil(members[0].Roles)
	s.Assert().NotEmpty(members[0].Roles)

	// Owner should have "owner" role (virtual role based on permissions)
	s.Assert().Contains(members[0].Roles, "Owner")
}

func (s *OrganizationServiceIntegrationTestSuite) TestRemoveMember() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	member := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), member))

	err := s.organizationService.AddMember(s.ctx, s.organization.ID, member.ID)
	s.Require().NoError(err)

	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID, member.ID}, organization.Members)

	err = s.organizationService.RemoveMember(s.ctx, s.organization.ID, member.ID)
	s.Require().NoError(err)

	organization, err = s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().ElementsMatch([]model.ID{s.owner.ID}, organization.Members)
}

func (s *OrganizationServiceIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	err := s.organizationService.Delete(s.ctx, s.organization.ID, false)
	s.Require().NoError(err)

	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Equal(model.OrganizationStatusDeleted, organization.Status)

	err = s.organizationService.Delete(s.ctx, s.organization.ID, true)
	s.Require().NoError(err)

	_, err = s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().ErrorIs(err, repository.ErrNotFound)
}

func (s *OrganizationServiceIntegrationTestSuite) TestInviteMember() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	email := "new.user@example.com"

	// Invite a new user
	err := s.organizationService.InviteMember(s.ctx, s.organization.ID, email)
	s.Require().NoError(err)

	// Verify invitation exists
	invitations, err := s.OrganizationRepo.GetInvitations(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)
	s.Assert().Equal(email, invitations[0].Email)

	// Verify user was created with pending status
	user, err := s.UserRepo.GetByEmail(context.Background(), email)
	s.Require().NoError(err)
	s.Assert().Equal(model.UserStatusPending, user.Status)

	// Verify user is not a member yet
	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().NotContains(organization.Members, user.ID)

	// Verify invitation token exists
	token, err := s.UserTokenRepository.Get(context.Background(), user.ID, model.UserTokenContextInvite)
	s.Require().NoError(err)
	s.Require().NotNil(token)
}

func (s *OrganizationServiceIntegrationTestSuite) TestInviteMemberWithExistingUser() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	// Create an existing active user
	existingUser := testModel.NewUser()
	existingUser.Email = "existing@example.com"
	existingUser.Status = model.UserStatusActive
	s.Require().NoError(s.UserRepo.Create(context.Background(), existingUser))

	// Invite the existing user
	err := s.organizationService.InviteMember(s.ctx, s.organization.ID, existingUser.Email)
	s.Require().NoError(err)

	// Verify invitation exists
	invitations, err := s.OrganizationRepo.GetInvitations(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)
	s.Assert().Equal(existingUser.ID, invitations[0].ID)

	// Verify user is still active (not changed)
	user, err := s.UserRepo.Get(context.Background(), existingUser.ID)
	s.Require().NoError(err)
	s.Assert().Equal(model.UserStatusActive, user.Status)
}

func (s *OrganizationServiceIntegrationTestSuite) TestInviteMemberWithRoleID() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	// Create a role
	role := testModel.NewRole()
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.owner.ID, s.organization.ID, role))

	email := "role.user@example.com"

	// Invite with role ID
	err := s.organizationService.InviteMember(s.ctx, s.organization.ID, email, role.ID)
	s.Require().NoError(err)

	// Verify invitation exists
	invitations, err := s.OrganizationRepo.GetInvitations(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)

	// Verify invitation token contains role_id
	user, err := s.UserRepo.GetByEmail(context.Background(), email)
	s.Require().NoError(err)

	token, err := s.UserTokenRepository.Get(context.Background(), user.ID, model.UserTokenContextInvite)
	s.Require().NoError(err)
	s.Require().NotNil(token)
}

func (s *OrganizationServiceIntegrationTestSuite) TestInviteMemberAlreadyMember() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	member := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), member))

	// Add user as member first
	err := s.organizationService.AddMember(s.ctx, s.organization.ID, member.ID)
	s.Require().NoError(err)

	// Try to invite already-member user
	err = s.organizationService.InviteMember(s.ctx, s.organization.ID, member.Email)
	s.Require().Error(err)
	s.Require().ErrorIs(err, service.ErrOrganizationMemberAlreadyExists)
}

func (s *OrganizationServiceIntegrationTestSuite) TestRevokeInvitation() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	email := "to.revoke@example.com"

	// Invite user
	err := s.organizationService.InviteMember(s.ctx, s.organization.ID, email)
	s.Require().NoError(err)

	// Get invited user
	user, err := s.UserRepo.GetByEmail(context.Background(), email)
	s.Require().NoError(err)

	// Verify invitation exists
	invitations, err := s.OrganizationRepo.GetInvitations(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)

	// Revoke invitation
	err = s.organizationService.RevokeInvitation(s.ctx, s.organization.ID, user.ID)
	s.Require().NoError(err)

	// Verify invitation is removed
	invitations, err = s.OrganizationRepo.GetInvitations(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(invitations, 0)

	// Verify invitation token is deleted
	_, err = s.UserTokenRepository.Get(context.Background(), user.ID, model.UserTokenContextInvite)
	s.Require().Error(err)
	s.Require().ErrorIs(err, repository.ErrNotFound)

	// Verify pending user is deleted (if not member of any org)
	if user.Status == model.UserStatusPending {
		_, err = s.UserRepo.Get(context.Background(), user.ID)
		s.Require().Error(err)
		s.Require().ErrorIs(err, repository.ErrNotFound)
	}
}

func (s *OrganizationServiceIntegrationTestSuite) TestRevokeInvitationWithActiveUser() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	// Create an active user
	activeUser := testModel.NewUser()
	activeUser.Status = model.UserStatusActive
	s.Require().NoError(s.UserRepo.Create(context.Background(), activeUser))

	// Invite user
	err := s.organizationService.InviteMember(s.ctx, s.organization.ID, activeUser.Email)
	s.Require().NoError(err)

	// Revoke invitation
	err = s.organizationService.RevokeInvitation(s.ctx, s.organization.ID, activeUser.ID)
	s.Require().NoError(err)

	// Verify user still exists (active users are not deleted)
	user, err := s.UserRepo.Get(context.Background(), activeUser.ID)
	s.Require().NoError(err)
	s.Assert().Equal(activeUser.ID, user.ID)
}

func (s *OrganizationServiceIntegrationTestSuite) TestAcceptInvitation() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	email := "to.accept@example.com"
	password := "securepassword123"

	// Invite user
	err := s.organizationService.InviteMember(s.ctx, s.organization.ID, email)
	s.Require().NoError(err)

	// Get invited user
	user, err := s.UserRepo.GetByEmail(context.Background(), email)
	s.Require().NoError(err)

	// Get the token that was captured from the email
	token, ok := s.capturedTokens[email]
	s.Require().True(ok, "token should have been captured from email")
	s.Require().NotEmpty(token, "token should not be empty")

	// Accept invitation
	err = s.organizationService.AcceptInvitation(context.Background(), s.organization.ID, token, password)
	s.Require().NoError(err)

	// Verify user is now a member
	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Contains(organization.Members, user.ID)

	// Verify invitation is removed
	invitations, err := s.OrganizationRepo.GetInvitations(context.Background(), s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Len(invitations, 0)

	// Verify invitation token is deleted
	_, err = s.UserTokenRepository.Get(context.Background(), user.ID, model.UserTokenContextInvite)
	s.Require().Error(err)
	s.Require().ErrorIs(err, repository.ErrNotFound)

	// Verify user is activated and password is set
	user, err = s.UserRepo.Get(context.Background(), user.ID)
	s.Require().NoError(err)
	s.Assert().Equal(model.UserStatusActive, user.Status)
	s.Assert().NotEmpty(user.Password)
}

func (s *OrganizationServiceIntegrationTestSuite) TestAcceptInvitationWithRoleID() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	// Create a role
	role := testModel.NewRole()
	s.Require().NoError(s.RoleRepo.Create(context.Background(), s.owner.ID, s.organization.ID, role))

	email := "role.user@example.com"
	password := "securepassword123"

	// Invite with role ID
	err := s.organizationService.InviteMember(s.ctx, s.organization.ID, email, role.ID)
	s.Require().NoError(err)

	// Get invited user
	user, err := s.UserRepo.GetByEmail(context.Background(), email)
	s.Require().NoError(err)

	// Get the token that was captured from the email
	token, ok := s.capturedTokens[email]
	s.Require().True(ok, "token should have been captured from email")
	s.Require().NotEmpty(token, "token should not be empty")

	// Accept invitation
	err = s.organizationService.AcceptInvitation(context.Background(), s.organization.ID, token, password)
	s.Require().NoError(err)

	// Verify user is now a member
	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Contains(organization.Members, user.ID)

	// Verify user is assigned to role by getting the role and checking members
	roleWithMembers, err := s.RoleRepo.Get(context.Background(), role.ID, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Contains(roleWithMembers.Members, user.ID)
}

func (s *OrganizationServiceIntegrationTestSuite) TestAcceptInvitationWithActiveUser() {
	s.Require().NoError(s.organizationService.Create(s.ctx, s.owner.ID, s.organization))

	// Create an active user
	activeUser := testModel.NewUser()
	activeUser.Status = model.UserStatusActive
	s.Require().NoError(s.UserRepo.Create(context.Background(), activeUser))

	// Invite user
	err := s.organizationService.InviteMember(s.ctx, s.organization.ID, activeUser.Email)
	s.Require().NoError(err)

	// Get the token that was captured from the email
	token, ok := s.capturedTokens[activeUser.Email]
	s.Require().True(ok, "token should have been captured from email")
	s.Require().NotEmpty(token, "token should not be empty")

	// Accept invitation (no password needed for active user)
	err = s.organizationService.AcceptInvitation(context.Background(), s.organization.ID, token, "")
	s.Require().NoError(err)

	// Verify user is now a member
	organization, err := s.organizationService.Get(s.ctx, s.organization.ID)
	s.Require().NoError(err)
	s.Assert().Contains(organization.Members, activeUser.ID)
}

func TestOrganizationServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationServiceIntegrationTestSuite))
}
