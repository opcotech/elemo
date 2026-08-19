package repository_test

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

	testUser   *repository.User
	createOpts repository.CreateNotificationOpts
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
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.createOpts = testModel.NewCreateNotificationOpts(s.testUser.ID)
}

func (s *NotificationRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *NotificationRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *NotificationRepositoryIntegrationTestSuite) TestCreate() {
	notification, err := s.NotificationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeNotification), notification.ID)
	s.Assert().NotNil(notification.CreatedAt)
}

func (s *NotificationRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.NotificationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	notification, err := s.NotificationRepo.Get(context.Background(), created.ID, created.Recipient, repository.NotificationDetailProjection())
	s.Require().NoError(err)

	s.Assert().Equal(created.ID, notification.ID)
	s.Assert().WithinDuration(*created.CreatedAt, *notification.CreatedAt, 100*time.Millisecond)
}

func (s *NotificationRepositoryIntegrationTestSuite) TestGetAllByRecipient() {
	_, err := s.NotificationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.NotificationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	notifications, err := s.NotificationRepo.ListByRecipient(context.Background(), s.createOpts.Recipient, repository.CursorPage{Size: 10}, repository.NotificationListProjection())
	s.Require().NoError(err)
	s.Assert().Len(notifications.Items, 2)

	notifications, err = s.NotificationRepo.ListByRecipient(context.Background(), s.createOpts.Recipient, repository.CursorPage{Size: 1}, repository.NotificationListProjection())
	s.Require().NoError(err)
	s.Assert().Len(notifications.Items, 1)
	s.Assert().True(notifications.PageInfo.HasMore)

	notifications, err = s.NotificationRepo.ListByRecipient(context.Background(), s.createOpts.Recipient, repository.CursorPage{Size: 1, Token: notifications.PageInfo.NextPageToken}, repository.NotificationListProjection())
	s.Require().NoError(err)
	s.Assert().Len(notifications.Items, 1)
	s.Assert().False(notifications.PageInfo.HasMore)
}

func (s *NotificationRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.NotificationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	notification, err := s.NotificationRepo.Update(context.Background(), created.ID, created.Recipient, repository.UpdateNotificationOpts{Read: true})
	s.Require().NoError(err)
	s.Require().True(notification.Read)

	notification, err = s.NotificationRepo.Update(context.Background(), created.ID, created.Recipient, repository.UpdateNotificationOpts{Read: false})
	s.Require().NoError(err)
	s.Require().False(notification.Read)
}

func (s *NotificationRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.NotificationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	s.Require().NoError(s.NotificationRepo.Delete(context.Background(), created.ID, created.Recipient))

	_, err = s.NotificationRepo.Get(context.Background(), created.ID, created.Recipient, repository.NotificationDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestNotificationRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationRepositoryIntegrationTestSuite))
}
