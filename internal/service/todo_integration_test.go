package service_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
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

	todo *model.Todo
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

	s.todo = testModel.NewTodo(s.testUser.ID, s.testUser.ID)
}

func (s *TodoServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *TodoServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *TodoServiceIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.todoService.Create(s.testUserContext, s.todo))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeTodo), s.todo.ID)
	s.Assert().NotNil(s.todo.CreatedAt)
	s.Assert().Nil(s.todo.UpdatedAt)
}

func (s *TodoServiceIntegrationTestSuite) TestCreateForOtherUser() {
	otherUser := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), otherUser))

	org := testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, org))
	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), org.ID, otherUser.ID))

	s.todo.CreatedBy = s.testUser.ID
	s.todo.OwnedBy = otherUser.ID

	s.Require().NoError(s.todoService.Create(s.testUserContext, s.todo))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeTodo), s.todo.ID)
	s.Assert().NotNil(s.todo.CreatedAt)
	s.Assert().Nil(s.todo.UpdatedAt)
}

func (s *TodoServiceIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.todoService.Create(s.testUserContext, s.todo))

	todo, err := s.todoService.Get(s.testUserContext, s.todo.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.todo.ID, todo.ID)
	s.Assert().Equal(s.todo.Title, todo.Title)
	s.Assert().Equal(s.todo.Description, todo.Description)
	s.Assert().Equal(s.todo.Priority, todo.Priority)
	s.Assert().Equal(s.todo.Completed, todo.Completed)
	s.Assert().Equal(s.todo.OwnedBy, todo.OwnedBy)
	s.Assert().Equal(s.todo.CreatedBy, todo.CreatedBy)
	s.Assert().WithinDuration(*s.todo.DueDate, *todo.DueDate, 100*time.Millisecond)
	s.Assert().WithinDuration(*s.todo.CreatedAt, *todo.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(todo.UpdatedAt)
}

func (s *TodoServiceIntegrationTestSuite) TestGetAll() {
	todo1 := testModel.NewTodo(s.testUser.ID, s.testUser.ID)
	todo1.Completed = true

	s.Require().NoError(s.todoService.Create(s.testUserContext, todo1))
	s.Require().NoError(s.todoService.Create(s.testUserContext, testModel.NewTodo(s.testUser.ID, s.testUser.ID)))
	s.Require().NoError(s.todoService.Create(s.testUserContext, testModel.NewTodo(s.testUser.ID, s.testUser.ID)))

	todos, err := s.todoService.GetAll(s.testUserContext, 0, 10, nil)
	s.Require().NoError(err)
	s.Assert().Len(todos, 3)

	todos, err = s.todoService.GetAll(s.testUserContext, 0, 10, convert.ToPointer(true))
	s.Require().NoError(err)
	s.Assert().Len(todos, 1)

	todos, err = s.todoService.GetAll(s.testUserContext, 0, 10, convert.ToPointer(false))
	s.Require().NoError(err)
	s.Assert().Len(todos, 2)
}

func (s *TodoServiceIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.todoService.Create(s.testUserContext, s.todo))

	priority := model.TodoPriorityCritical
	patch := map[string]any{
		"title":       "new title",
		"description": "new description",
		"priority":    priority.String(),
	}

	todo, err := s.todoService.Update(s.testUserContext, s.todo.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.todo.ID, todo.ID)
	s.Assert().Equal(patch["title"], todo.Title)
	s.Assert().Equal(patch["description"], todo.Description)
	s.Assert().Equal(priority, todo.Priority)
	s.Assert().Equal(s.todo.Completed, todo.Completed)
	s.Assert().Equal(s.todo.OwnedBy, todo.OwnedBy)
	s.Assert().Equal(s.todo.CreatedBy, todo.CreatedBy)
	s.Assert().WithinDuration(*s.todo.DueDate, *todo.DueDate, 100*time.Millisecond)
	s.Assert().WithinDuration(*s.todo.CreatedAt, *todo.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(todo.UpdatedAt)
}

func (s *TodoServiceIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.todoService.Create(s.testUserContext, s.todo))

	s.Require().NoError(s.todoService.Delete(s.testUserContext, s.todo.ID))

	_, err := s.todoService.Get(s.testUserContext, s.todo.ID)
	s.Assert().ErrorIs(err, service.ErrNoPermission)
}

func TestTodoServiceIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(TodoServiceIntegrationTestSuite))
}
