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
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type OrganizationServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite
	testutil.SearchContainerIntegrationTestSuite

	organizationService service.OrganizationService
	emailService        service.EmailService
	emailSender         *mocksvc.MockEmailSender
	capturedTokens      map[string]string

	owner *repository.User
	ctx   context.Context
}

func (s *OrganizationServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)
	s.SetupSearch(&s.ContainerIntegrationTestSuite, container)

	permissionService, err := service.NewPermissionService(s.PermissionRepo, s.RoleRepo)
	s.Require().NoError(err)

	searchService, err := service.NewSearchService(
		s.SearchRepo,
		permissionService,
		nil,
	)
	s.Require().NoError(err)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		permissionService,
	)
	s.Require().NoError(err)

	ctrl := gomock.NewController(s.T())
	s.emailSender = mocksvc.NewMockEmailSender(ctrl)
	s.capturedTokens = make(map[string]string)
	s.emailSender.EXPECT().SendEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		Do(func(_ context.Context, _, to string, template any) {
			if tmpl, ok := template.(*email.Template); ok && tmpl != nil {
				if data, ok := tmpl.Data.Get().(*email.OrganizationInviteTemplateData); ok && data != nil {
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

	smtpConf := &config.SMTPConfig{
		ClientURL:      "http://localhost:3000",
		SupportAddress: "support@example.com",
	}
	s.emailService, err = service.NewEmailService(s.emailSender, "templates", smtpConf)
	s.Require().NoError(err)

	notificationService, err := service.NewNotificationService(s.NotificationRepo)
	s.Require().NoError(err)

	s.organizationService, err = service.NewOrganizationService(
		s.OrganizationRepo,
		s.UserRepo,
		s.UserTokenRepository,
		s.RoleRepo,
		permissionService,
		licenseService,
		s.emailService,
		notificationService,
		searchService,
	)
	s.Require().NoError(err)
}

func (s *OrganizationServiceIntegrationTestSuite) SetupTest() {
	var err error
	s.owner, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.GrantOrganizationCreate(s.owner.ID, s.Neo4jDB))
	s.capturedTokens = make(map[string]string)
}

func (s *OrganizationServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupSearch(&s.ContainerIntegrationTestSuite)
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *OrganizationServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func serviceCreateOrgOpts() service.CreateOrganizationOpts {
	o := testModel.NewCreateOrganizationOpts(model.MustNewNilID(model.ResourceTypeUser))
	return service.CreateOrganizationOpts{
		Name:    o.Name,
		Slug:    o.Slug,
		Email:   o.Email,
		Logo:    o.Logo,
		Website: o.Website,
		Status:  o.Status,
	}
}

func (s *OrganizationServiceIntegrationTestSuite) TestCreate() {
	org, err := s.organizationService.Create(s.ctx, s.owner.ID, serviceCreateOrgOpts())
	s.Require().NoError(err)
	s.Require().NotEmpty(org.ID)
	s.Assert().NotNil(org.CreatedAt)
	s.Assert().Nil(org.UpdatedAt)
}

func (s *OrganizationServiceIntegrationTestSuite) TestCreateWithoutGrant() {
	user, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, user.ID)

	_, err = s.organizationService.Create(ctx, user.ID, serviceCreateOrgOpts())
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *OrganizationServiceIntegrationTestSuite) TestGet() {
	created, err := s.organizationService.Create(s.ctx, s.owner.ID, serviceCreateOrgOpts())
	s.Require().NoError(err)

	org, err := s.organizationService.Get(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, org.ID)
	s.Assert().Equal(created.Name, org.Name)
	s.Assert().Equal(created.Email, org.Email)
}

func (s *OrganizationServiceIntegrationTestSuite) TestList() {
	_, err := s.organizationService.Create(s.ctx, s.owner.ID, serviceCreateOrgOpts())
	s.Require().NoError(err)
	_, err = s.organizationService.Create(s.ctx, s.owner.ID, serviceCreateOrgOpts())
	s.Require().NoError(err)

	orgs, err := s.organizationService.List(s.ctx, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(len(orgs.Items), 2)
}

func (s *OrganizationServiceIntegrationTestSuite) TestUpdate() {
	created, err := s.organizationService.Create(s.ctx, s.owner.ID, serviceCreateOrgOpts())
	s.Require().NoError(err)

	org, err := s.organizationService.Update(s.ctx, created.ID, service.UpdateOrganizationOpts{
		Name: optional.Some("updated org name"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("updated org name", org.Name)
	s.Assert().NotNil(org.UpdatedAt)
}

func (s *OrganizationServiceIntegrationTestSuite) TestAddMember() {
	created, err := s.organizationService.Create(s.ctx, s.owner.ID, serviceCreateOrgOpts())
	s.Require().NoError(err)

	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.Require().NoError(s.organizationService.AddMember(s.ctx, created.ID, member.ID))

	members, err := s.organizationService.ListMembers(s.ctx, created.ID, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().GreaterOrEqual(len(members.Items), 2)
}

func (s *OrganizationServiceIntegrationTestSuite) TestRemoveMember() {
	created, err := s.organizationService.Create(s.ctx, s.owner.ID, serviceCreateOrgOpts())
	s.Require().NoError(err)

	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.organizationService.AddMember(s.ctx, created.ID, member.ID))
	s.Require().NoError(s.organizationService.RemoveMember(s.ctx, created.ID, member.ID))
}

func (s *OrganizationServiceIntegrationTestSuite) TestDelete() {
	created, err := s.organizationService.Create(s.ctx, s.owner.ID, serviceCreateOrgOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.organizationService.Delete(s.ctx, created.ID, false))
}

func (s *OrganizationServiceIntegrationTestSuite) TestInviteMember() {
	created, err := s.organizationService.Create(s.ctx, s.owner.ID, serviceCreateOrgOpts())
	s.Require().NoError(err)

	inviteEmail := testutil.GenerateEmail(10)
	s.Require().NoError(s.organizationService.InviteMember(s.ctx, created.ID, service.InviteOrganizationMemberOpts{
		Email: inviteEmail,
	}))
	s.Assert().Contains(s.capturedTokens, inviteEmail)
}

func TestOrganizationServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationServiceIntegrationTestSuite))
}
