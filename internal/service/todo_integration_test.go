package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type TodoServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	todoService service.TodoService

	testUser        *model.User
	testUserContext context.Context
}

func (s *TodoServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		service.WithPermissionRepository(s.PermissionRepo),
	)
	s.Require().NoError(err)

	s.todoService, err = service.NewTodoService(
		service.WithTodoRepository(s.TodoRepo),
		service.WithPermissionRepository(s.PermissionRepo),
		service.WithLicenseService(licenseService),
	)
	s.Require().NoError(err)
}

func (s *TodoServiceIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))
	s.testUserContext = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.testUser.ID)
}

func (s *TodoServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *TodoServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *TodoServiceIntegrationTestSuite) Create() {
	s.T().Skip("implement me")
}

func (s *TodoServiceIntegrationTestSuite) Get() {
	s.T().Skip("implement me")
}

func (s *TodoServiceIntegrationTestSuite) GetAll() {
	s.T().Skip("implement me")
}

func (s *TodoServiceIntegrationTestSuite) Update() {
	s.T().Skip("implement me")
}

func (s *TodoServiceIntegrationTestSuite) Delete() {
	s.T().Skip("implement me")
}

func TestTodoServiceIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(TodoServiceIntegrationTestSuite))
}
