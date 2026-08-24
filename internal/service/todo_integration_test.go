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
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type TodoServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	todoService service.TodoService

	testUser        *repository.User
	testUserContext context.Context
}

func (s *TodoServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)

	permissionService, err := service.NewPermissionService(s.PermissionRepo, s.RoleRepo)
	s.Require().NoError(err)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		permissionService,
	)
	s.Require().NoError(err)

	s.todoService, err = service.NewTodoService(
		s.TodoRepo,
		licenseService,
	)
	s.Require().NoError(err)
}

func (s *TodoServiceIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testUserContext = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.testUser.ID)
}

func newTestCreateTodoOpts(ownedBy, createdBy model.ID) service.CreateTodoOpts {
	return service.CreateTodoOpts{
		Title:       "test todo title",
		Description: "test todo description text",
		Priority:    model.TodoPriorityNormal,
		Completed:   false,
		OwnedBy:     ownedBy,
		CreatedBy:   createdBy,
		DueDate:     convert.ToPointer(time.Now().UTC().Add(24 * time.Hour)),
	}
}

func (s *TodoServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *TodoServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *TodoServiceIntegrationTestSuite) TestCreate() {
	opts := newTestCreateTodoOpts(s.testUser.ID, s.testUser.ID)
	todo, err := s.todoService.Create(s.testUserContext, opts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeTodo), todo.ID)
	s.Assert().NotNil(todo.CreatedAt)
	s.Assert().Nil(todo.UpdatedAt)
}

func (s *TodoServiceIntegrationTestSuite) TestCreateForOtherUser() {
	otherUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	opts := newTestCreateTodoOpts(otherUser.ID, s.testUser.ID)
	_, err = s.todoService.Create(s.testUserContext, opts)
	s.Require().ErrorIs(err, service.ErrNoPermission)
}

func (s *TodoServiceIntegrationTestSuite) TestGet() {
	opts := newTestCreateTodoOpts(s.testUser.ID, s.testUser.ID)
	createdTodo, err := s.todoService.Create(s.testUserContext, opts)
	s.Require().NoError(err)

	todo, err := s.todoService.Get(s.testUserContext, createdTodo.ID)
	s.Require().NoError(err)

	s.Assert().Equal(createdTodo.ID, todo.ID)
	s.Assert().Equal(opts.Title, todo.Title)
	s.Assert().Equal(opts.Description, todo.Description)
	s.Assert().Equal(opts.Priority, todo.Priority)
	s.Assert().Equal(opts.Completed, todo.Completed)
	s.Assert().Equal(opts.OwnedBy, todo.OwnedBy)
	s.Assert().Equal(opts.CreatedBy, todo.CreatedBy)
	s.Assert().WithinDuration(*opts.DueDate, *todo.DueDate, 100*time.Millisecond)
	s.Assert().Nil(todo.UpdatedAt)
}

func (s *TodoServiceIntegrationTestSuite) TestList() {
	opts1 := newTestCreateTodoOpts(s.testUser.ID, s.testUser.ID)
	opts1.Completed = true
	_, err := s.todoService.Create(s.testUserContext, opts1)
	s.Require().NoError(err)

	_, err = s.todoService.Create(s.testUserContext, newTestCreateTodoOpts(s.testUser.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.todoService.Create(s.testUserContext, newTestCreateTodoOpts(s.testUser.ID, s.testUser.ID))
	s.Require().NoError(err)

	todos, err := s.todoService.List(s.testUserContext, service.CursorPage{Size: 10}, nil)
	s.Require().NoError(err)
	s.Assert().Len(todos.Items, 3)

	todos, err = s.todoService.List(s.testUserContext, service.CursorPage{Size: 10}, convert.ToPointer(true))
	s.Require().NoError(err)
	s.Assert().Len(todos.Items, 1)

	todos, err = s.todoService.List(s.testUserContext, service.CursorPage{Size: 10}, convert.ToPointer(false))
	s.Require().NoError(err)
	s.Assert().Len(todos.Items, 2)
}

func (s *TodoServiceIntegrationTestSuite) TestUpdate() {
	opts := newTestCreateTodoOpts(s.testUser.ID, s.testUser.ID)
	createdTodo, err := s.todoService.Create(s.testUserContext, opts)
	s.Require().NoError(err)

	updateOpts := service.UpdateTodoOpts{
		Title:       optional.Some("new title"),
		Description: optional.Some("new description text"),
		Priority:    optional.Some(model.TodoPriorityCritical),
	}

	todo, err := s.todoService.Update(s.testUserContext, createdTodo.ID, updateOpts)
	s.Require().NoError(err)

	s.Assert().Equal(createdTodo.ID, todo.ID)
	s.Assert().Equal("new title", todo.Title)
	s.Assert().Equal("new description text", todo.Description)
	s.Assert().Equal(model.TodoPriorityCritical, todo.Priority)
	s.Assert().NotNil(todo.UpdatedAt)
}

func (s *TodoServiceIntegrationTestSuite) TestDelete() {
	opts := newTestCreateTodoOpts(s.testUser.ID, s.testUser.ID)
	createdTodo, err := s.todoService.Create(s.testUserContext, opts)
	s.Require().NoError(err)

	s.Require().NoError(s.todoService.Delete(s.testUserContext, createdTodo.ID))

	_, err = s.todoService.Get(s.testUserContext, createdTodo.ID)
	s.Assert().ErrorIs(err, service.ErrTodoGet)
}

func TestTodoServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TodoServiceIntegrationTestSuite))
}
