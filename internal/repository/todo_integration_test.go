package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/suite"
)

type TodoRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	createOpts repository.CreateTodoOpts
}

func (s *TodoRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *TodoRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreateTodoOpts(s.testUser.ID, s.testUser.ID)
}

func (s *TodoRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *TodoRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *TodoRepositoryIntegrationTestSuite) TestCreate() {
	todo, err := s.TodoRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeTodo), todo.ID)
	s.Assert().NotNil(todo.CreatedAt)
	s.Assert().Nil(todo.UpdatedAt)
}

func (s *TodoRepositoryIntegrationTestSuite) TestGet() {
	createdTodo, err := s.TodoRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	todo, err := s.TodoRepo.Get(context.Background(), createdTodo.ID)
	s.Require().NoError(err)

	s.Assert().Equal(createdTodo.ID, todo.ID)
	s.Assert().Equal(s.createOpts.Title, todo.Title)
	s.Assert().Equal(s.createOpts.Description, todo.Description)
	s.Assert().Equal(s.createOpts.CreatedBy, todo.CreatedBy)
	s.Assert().Equal(s.createOpts.OwnedBy, todo.OwnedBy)
	s.Assert().Equal(s.createOpts.Completed, todo.Completed)
	s.Assert().WithinDuration(*s.createOpts.DueDate, *todo.DueDate, 100*time.Millisecond)
	s.Assert().WithinDuration(*createdTodo.CreatedAt, *todo.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(todo.UpdatedAt)
}

func (s *TodoRepositoryIntegrationTestSuite) TestGetByOwner() {
	completedOpts := s.createOpts
	completedOpts.Completed = true

	_, err := s.TodoRepo.Create(context.Background(), completedOpts)
	s.Require().NoError(err)
	_, err = s.TodoRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	todos, err := s.TodoRepo.GetByOwner(context.Background(), s.testUser.ID, 0, 10, nil)
	s.Require().NoError(err)
	s.Assert().Len(todos, 2)

	todos, err = s.TodoRepo.GetByOwner(context.Background(), s.testUser.ID, 0, 10, convert.ToPointer(false))
	s.Require().NoError(err)
	s.Assert().Len(todos, 1)

	todos, err = s.TodoRepo.GetByOwner(context.Background(), s.testUser.ID, 0, 10, convert.ToPointer(true))
	s.Require().NoError(err)
	s.Assert().Len(todos, 1)
}

func (s *TodoRepositoryIntegrationTestSuite) TestUpdate() {
	createdTodo, err := s.TodoRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	dueDate := time.Now().UTC().Add(1 * time.Hour)
	updateOpts := repository.UpdateTodoOpts{
		Title:       optional.Some("New title"),
		Description: optional.Some("New description"),
		Completed:   optional.Some(true),
		DueDate:     optional.Some(dueDate),
	}

	todo, err := s.TodoRepo.Update(context.Background(), createdTodo.ID, updateOpts)
	s.Require().NoError(err)

	s.Assert().Equal(createdTodo.ID, todo.ID)
	s.Assert().Equal("New title", todo.Title)
	s.Assert().Equal("New description", todo.Description)
	s.Assert().Equal(createdTodo.CreatedBy, todo.CreatedBy)
	s.Assert().Equal(createdTodo.OwnedBy, todo.OwnedBy)
	s.Assert().True(todo.Completed)
	s.Assert().WithinDuration(dueDate, *todo.DueDate, 100*time.Millisecond)
	s.Assert().WithinDuration(*createdTodo.CreatedAt, *todo.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(todo.UpdatedAt)
}

func (s *TodoRepositoryIntegrationTestSuite) TestDelete() {
	createdTodo, err := s.TodoRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.TodoRepo.Delete(context.Background(), createdTodo.ID))

	_, err = s.TodoRepo.Get(context.Background(), createdTodo.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestTodoRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TodoRepositoryIntegrationTestSuite))
}

type CachedTodoRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser   *repository.User
	createOpts repository.CreateTodoOpts
	todoRepo   *repository.RedisCachedTodoRepository
}

func (s *CachedTodoRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.todoRepo, _ = repository.NewCachedTodoRepository(s.TodoRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedTodoRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreateTodoOpts(s.testUser.ID, s.testUser.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TestCreate() {
	todo, err := s.todoRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeTodo), todo.ID)
	s.Assert().NotNil(todo.CreatedAt)
	s.Assert().Nil(todo.UpdatedAt)

	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TestGet() {
	createdTodo, err := s.todoRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	original, err := s.TodoRepo.Get(context.Background(), createdTodo.ID)
	s.Require().NoError(err)

	usingCache, err := s.todoRepo.Get(context.Background(), createdTodo.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.todoRepo.Get(context.Background(), createdTodo.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache.ID, cached.ID)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TestGetByOwner() {
	completedOpts := s.createOpts
	completedOpts.Completed = true

	_, err := s.todoRepo.Create(context.Background(), completedOpts)
	s.Require().NoError(err)
	_, err = s.todoRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	originalTodos, err := s.TodoRepo.GetByOwner(context.Background(), s.testUser.ID, 0, 10, nil)
	s.Require().NoError(err)

	usingCacheTodos, err := s.todoRepo.GetByOwner(context.Background(), s.testUser.ID, 0, 10, nil)
	s.Require().NoError(err)

	s.Assert().Equal(originalTodos, usingCacheTodos)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedTodos, err := s.todoRepo.GetByOwner(context.Background(), s.testUser.ID, 0, 10, nil)
	s.Require().NoError(err)
	s.Assert().Equal(len(usingCacheTodos), len(cachedTodos))
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TestUpdate() {
	createdTodo, err := s.todoRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	dueDate := time.Now().UTC().Add(1 * time.Hour)
	updateOpts := repository.UpdateTodoOpts{
		Title:       optional.Some("New title"),
		Description: optional.Some("New description"),
		Completed:   optional.Some(true),
		DueDate:     optional.Some(dueDate),
	}

	todo, err := s.todoRepo.Update(context.Background(), createdTodo.ID, updateOpts)
	s.Require().NoError(err)

	s.Assert().Equal(createdTodo.ID, todo.ID)
	s.Assert().Equal("New title", todo.Title)
	s.Assert().Equal("New description", todo.Description)
	s.Assert().Equal(createdTodo.CreatedBy, todo.CreatedBy)
	s.Assert().Equal(createdTodo.OwnedBy, todo.OwnedBy)
	s.Assert().True(todo.Completed)
	s.Assert().WithinDuration(dueDate, *todo.DueDate, 100*time.Millisecond)
	s.Assert().WithinDuration(*createdTodo.CreatedAt, *todo.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(todo.UpdatedAt)

	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TestDelete() {
	createdTodo, err := s.todoRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	_, err = s.todoRepo.Get(context.Background(), createdTodo.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.todoRepo.Delete(context.Background(), createdTodo.ID))

	_, err = s.todoRepo.Get(context.Background(), createdTodo.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedTodoRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedTodoRepositoryIntegrationTestSuite))
}
