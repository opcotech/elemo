package redis_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/repository/redis"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type CachedTodoRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser *model.User
	todo     *model.Todo
	todoRepo *redis.CachedTodoRepository
}

func (s *CachedTodoRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}

	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())

	s.todoRepo, _ = redis.NewCachedTodoRepository(s.TodoRepo, redis.WithDatabase(s.RedisDB))
}

func (s *CachedTodoRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.todo = testModel.NewTodo(s.testUser.ID, s.testUser.ID)
	s.Require().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.todoRepo.Create(context.Background(), s.todo))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeTodo), s.todo.ID)
	s.Assert().NotNil(s.todo.CreatedAt)
	s.Assert().Nil(s.todo.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.todoRepo.Create(context.Background(), s.todo))

	original, err := s.TodoRepo.Get(context.Background(), s.todo.ID)
	s.Require().NoError(err)

	usingCache, err := s.todoRepo.Get(context.Background(), s.todo.ID)
	s.Require().NoError(err)

	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cached, err := s.todoRepo.Get(context.Background(), s.todo.ID)
	s.Require().NoError(err)

	s.Assert().Equal(usingCache, cached)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TestGetByOwner() {
	completedTodo := testModel.NewTodo(s.testUser.ID, s.testUser.ID)
	completedTodo.Completed = true

	s.Require().NoError(s.todoRepo.Create(context.Background(), completedTodo))
	s.Require().NoError(s.todoRepo.Create(context.Background(), s.todo))

	originalTodos, err := s.TodoRepo.GetByOwner(context.Background(), s.todo.OwnedBy, 0, 10, nil)
	s.Require().NoError(err)

	usingCacheTodos, err := s.todoRepo.GetByOwner(context.Background(), s.todo.OwnedBy, 0, 10, nil)
	s.Require().NoError(err)

	s.Assert().Equal(originalTodos, usingCacheTodos)
	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	cachedTodos, err := s.todoRepo.GetByOwner(context.Background(), s.todo.OwnedBy, 0, 10, nil)
	s.Require().NoError(err)
	s.Assert().Equal(usingCacheTodos, cachedTodos)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.todoRepo.Create(context.Background(), s.todo))

	dueDate := time.Now().Add(1 * time.Hour)
	patch := map[string]any{
		"title":       "New title",
		"description": "New description",
		"due_date":    dueDate.Format(time.RFC3339Nano),
		"completed":   true,
	}

	todo, err := s.todoRepo.Update(context.Background(), s.todo.ID, patch)
	s.Require().NoError(err)

	s.Assert().Equal(s.todo.ID, todo.ID)
	s.Assert().Equal(patch["title"], todo.Title)
	s.Assert().Equal(patch["description"], todo.Description)
	s.Assert().Equal(s.todo.CreatedBy, todo.CreatedBy)
	s.Assert().Equal(s.todo.OwnedBy, todo.OwnedBy)
	s.Assert().True(todo.Completed)
	s.Assert().WithinDuration(dueDate, *todo.DueDate, 100*time.Millisecond)
	s.Assert().WithinDuration(*s.todo.CreatedAt, *todo.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(todo.UpdatedAt)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedTodoRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.todoRepo.Create(context.Background(), s.todo))

	_, err := s.todoRepo.Get(context.Background(), s.todo.ID)
	s.Require().NoError(err)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 1)

	s.Require().NoError(s.todoRepo.Delete(context.Background(), s.todo.ID))

	_, err = s.todoRepo.Get(context.Background(), s.todo.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)

	s.Assert().Len(s.GetKeys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func TestCachedTodoRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CachedTodoRepositoryIntegrationTestSuite))
}
