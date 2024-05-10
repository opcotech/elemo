package pg_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type NotificationRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	testUser     *model.User
	notification *model.Notification
}

func (s *NotificationRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)
}

func (s *NotificationRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.notification = testModel.NewNotification(s.testUser.ID)
}

func (s *NotificationRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *NotificationRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *NotificationRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.NotificationRepo.Create(context.Background(), s.notification))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeNotification), s.notification.ID)
	s.Assert().NotNil(s.notification.CreatedAt)
}

func (s *NotificationRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.NotificationRepo.Create(context.Background(), s.notification))

	notification, err := s.NotificationRepo.Get(context.Background(), s.notification.ID, s.notification.Recipient)
	s.Require().NoError(err)

	s.Assert().Equal(s.notification.ID, notification.ID)
	s.Assert().WithinDuration(*s.notification.CreatedAt, *notification.CreatedAt, 100*time.Millisecond)
}

func (s *NotificationRepositoryIntegrationTestSuite) TestGetAllByRecipient() {
	s.Require().NoError(s.NotificationRepo.Create(context.Background(), s.notification))
	s.Require().NoError(s.NotificationRepo.Create(context.Background(), s.notification))

	notifications, err := s.NotificationRepo.GetAllByRecipient(context.Background(), s.notification.Recipient, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(notifications, 2)

	notifications, err = s.NotificationRepo.GetAllByRecipient(context.Background(), s.notification.Recipient, 0, 1)
	s.Require().NoError(err)
	s.Assert().Len(notifications, 1)

	notifications, err = s.NotificationRepo.GetAllByRecipient(context.Background(), s.notification.Recipient, 1, 1)
	s.Require().NoError(err)
	s.Assert().Len(notifications, 1)

	notifications, err = s.NotificationRepo.GetAllByRecipient(context.Background(), s.notification.Recipient, 2, 1)
	s.Require().NoError(err)
	s.Assert().Len(notifications, 0)
}

func (s *NotificationRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.NotificationRepo.Create(context.Background(), s.notification))

	notification, err := s.NotificationRepo.Update(context.Background(), s.notification.ID, s.notification.Recipient, true)
	s.Require().NoError(err)
	s.Require().True(notification.Read)

	notification, err = s.NotificationRepo.Update(context.Background(), s.notification.ID, s.notification.Recipient, false)
	s.Require().NoError(err)
	s.Require().False(notification.Read)
}

func (s *NotificationRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.NotificationRepo.Create(context.Background(), s.notification))

	s.Require().NoError(s.NotificationRepo.Delete(context.Background(), s.notification.ID, s.notification.Recipient))

	_, err := s.NotificationRepo.Get(context.Background(), s.notification.ID, s.notification.Recipient)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestNotificationRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationRepositoryIntegrationTestSuite))
}
