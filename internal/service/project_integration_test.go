package service_test

import (
	"context"
	"reflect"
	"strings"
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

type ProjectServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	projectService service.ProjectService

	owner     *repository.User
	org       *repository.Organization
	namespace *repository.Namespace

	ctx context.Context
}

func (s *ProjectServiceIntegrationTestSuite) SetupSuite() {
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

	s.projectService, err = service.NewProjectService(
		service.WithProjectRepository(s.ProjectRepo),
		service.WithPermissionService(permissionService),
		service.WithLicenseService(licenseService),
	)
	s.Require().NoError(err)
}

func (s *ProjectServiceIntegrationTestSuite) SetupTest() {
	var err error
	s.owner, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.owner.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.owner.ID, s.Neo4jDB))

	s.org, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.owner.ID))
	s.Require().NoError(err)

	s.namespace, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.owner.ID, s.org.ID))
	s.Require().NoError(err)

	_, err = s.PermissionRepo.Create(context.Background(), repository.CreatePermissionOpts{
		Subject: s.owner.ID,
		Target:  s.namespace.ID,
		Kind:    model.PermissionKindWrite,
	})
	s.Require().NoError(err)
}

func (s *ProjectServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *ProjectServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *ProjectServiceIntegrationTestSuite) newCreateOpts() service.CreateProjectOpts {
	return service.CreateProjectOpts{
		Key:         strings.ToUpper(pkg.GenerateRandomStringAlpha(3)),
		Name:        pkg.GenerateRandomString(10),
		Description: pkg.GenerateRandomString(20),
		Logo:        "https://www.gravatar.com/avatar",
		Status:      model.ProjectStatusActive,
	}
}

func (s *ProjectServiceIntegrationTestSuite) TestCreate() {
	opts := s.newCreateOpts()
	project, err := s.projectService.Create(s.ctx, s.namespace.ID, opts)
	s.Require().NoError(err)
	s.Require().NotEmpty(project.ID)
	s.Assert().Equal(opts.Key, project.Key)
	s.Assert().Equal(opts.Name, project.Name)
	s.Assert().Equal(opts.Description, project.Description)
	s.Assert().Equal(opts.Status, project.Status)
	s.Assert().NotNil(project.CreatedAt)

	hasPermission, err := s.PermissionRepo.HasPermission(
		context.Background(),
		s.owner.ID,
		project.ID,
		model.PermissionKindAll,
	)
	s.Require().NoError(err)
	s.Assert().True(hasPermission)
}

func (s *ProjectServiceIntegrationTestSuite) TestCreateNormalizesKeyToUppercase() {
	opts := s.newCreateOpts()
	opts.Key = strings.ToLower(opts.Key)

	project, err := s.projectService.Create(s.ctx, s.namespace.ID, opts)
	s.Require().NoError(err)
	s.Assert().Equal(strings.ToUpper(opts.Key), project.Key)
}

func (s *ProjectServiceIntegrationTestSuite) TestCreateWithoutPermission() {
	otherUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	otherCtx := context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUser.ID)

	_, err = s.projectService.Create(otherCtx, s.namespace.ID, s.newCreateOpts())
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func (s *ProjectServiceIntegrationTestSuite) TestGet() {
	created, err := s.projectService.Create(s.ctx, s.namespace.ID, s.newCreateOpts())
	s.Require().NoError(err)

	project, err := s.projectService.Get(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, project.ID)
	s.Assert().Equal(created.Key, project.Key)
	s.Assert().Equal(created.Name, project.Name)
}

func (s *ProjectServiceIntegrationTestSuite) TestGetByKey() {
	created, err := s.projectService.Create(s.ctx, s.namespace.ID, s.newCreateOpts())
	s.Require().NoError(err)

	project, err := s.projectService.GetByKey(s.ctx, created.Key)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, project.ID)
	s.Assert().Equal(created.Key, project.Key)
	s.Assert().Equal(created.Name, project.Name)
}

func (s *ProjectServiceIntegrationTestSuite) TestGetAll() {
	_, err := s.projectService.Create(s.ctx, s.namespace.ID, s.newCreateOpts())
	s.Require().NoError(err)
	_, err = s.projectService.Create(s.ctx, s.namespace.ID, s.newCreateOpts())
	s.Require().NoError(err)

	projects, err := s.projectService.GetAll(s.ctx, s.namespace.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(projects, 2)
}

func (s *ProjectServiceIntegrationTestSuite) TestUpdate() {
	created, err := s.projectService.Create(s.ctx, s.namespace.ID, s.newCreateOpts())
	s.Require().NoError(err)

	project, err := s.projectService.Update(s.ctx, created.ID, service.UpdateProjectOpts{
		Name: optional.Some("updated-project"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("updated-project", project.Name)
	s.Assert().NotNil(project.UpdatedAt)
}

func (s *ProjectServiceIntegrationTestSuite) TestUpdateNormalizesKeyToUppercase() {
	created, err := s.projectService.Create(s.ctx, s.namespace.ID, s.newCreateOpts())
	s.Require().NoError(err)

	newKey := strings.ToLower(pkg.GenerateRandomStringAlpha(4))
	project, err := s.projectService.Update(s.ctx, created.ID, service.UpdateProjectOpts{
		Key: optional.Some(newKey),
	})
	s.Require().NoError(err)
	s.Assert().Equal(strings.ToUpper(newKey), project.Key)
}

func (s *ProjectServiceIntegrationTestSuite) TestDelete() {
	created, err := s.projectService.Create(s.ctx, s.namespace.ID, s.newCreateOpts())
	s.Require().NoError(err)

	s.Require().NoError(s.projectService.Delete(s.ctx, created.ID))
	_, err = s.projectService.Get(s.ctx, created.ID)
	s.Assert().Error(err)
}

func TestProjectServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectServiceIntegrationTestSuite))
}
