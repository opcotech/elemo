package service_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type NotificationServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	notificationService service.NotificationService

	recipient    *model.User
	notification *model.Notification

	ctx context.Context
}

func (s *NotificationServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupPg(&s.ContainerIntegrationTestSuite, container)

	var err error
	s.notificationService, err = service.NewNotificationService(s.NotificationRepo)
	s.Require().NoError(err)
}

func (s *NotificationServiceIntegrationTestSuite) SetupTest() {
	s.recipient = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.recipient))

	s.ctx = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.recipient.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.recipient.ID, s.Neo4jDB))

	s.notification = testModel.NewNotification(s.recipient.ID)
}

func (s *NotificationServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *NotificationServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *NotificationServiceIntegrationTestSuite) TestCreate() {
	err := s.notificationService.Create(s.ctx, s.notification)
	s.Require().NoError(err)
	s.Require().NotEmpty(s.notification.ID)
	s.Require().NotEmpty(s.recipient.ID)
	s.Assert().NotNil(s.notification.CreatedAt)
	s.Assert().Nil(s.notification.UpdatedAt)
}

func (s *NotificationServiceIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.notificationService.Create(s.ctx, s.notification))

	notification, err := s.notificationService.Get(s.ctx, s.notification.ID, s.notification.Recipient)
	s.Require().NoError(err)
	s.Assert().Equal(s.notification.ID, notification.ID)
	s.Assert().Equal(s.notification.Title, notification.Title)
	s.Assert().Equal(s.notification.Description, notification.Description)
	s.Assert().WithinDuration(*s.notification.CreatedAt, *notification.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(s.notification.UpdatedAt)
	s.Assert().Nil(notification.UpdatedAt)
}

func (s *NotificationServiceIntegrationTestSuite) TestGetAllByRecipient() {
	s.Require().NoError(s.notificationService.Create(s.ctx, testModel.NewNotification(s.recipient.ID)))
	s.Require().NoError(s.notificationService.Create(s.ctx, testModel.NewNotification(s.recipient.ID)))
	s.Require().NoError(s.notificationService.Create(s.ctx, testModel.NewNotification(s.recipient.ID)))

	notifications, err := s.notificationService.GetAllByRecipient(s.ctx, s.recipient.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(notifications, 3)

	notifications, err = s.notificationService.GetAllByRecipient(s.ctx, s.recipient.ID, 0, 2)
	s.Require().NoError(err)
	s.Assert().Len(notifications, 2)

	notifications, err = s.notificationService.GetAllByRecipient(s.ctx, s.recipient.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(notifications, 2)

	notifications, err = s.notificationService.GetAllByRecipient(s.ctx, s.recipient.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(notifications, 1)

	notifications, err = s.notificationService.GetAllByRecipient(s.ctx, s.recipient.ID, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(notifications, 0)
}

func (s *NotificationServiceIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.notificationService.Create(s.ctx, s.notification))

	notification, err := s.notificationService.Update(s.ctx, s.notification.ID, s.notification.Recipient, true)
	s.Require().NoError(err)

	s.Assert().True(notification.Read)
	s.Assert().Equal(s.notification.CreatedAt, notification.CreatedAt)
	s.Assert().NotNil(notification.UpdatedAt)

	notification, err = s.notificationService.Update(s.ctx, s.notification.ID, s.notification.Recipient, false)
	s.Require().NoError(err)

	s.Assert().False(notification.Read)
	s.Assert().WithinDuration(*s.notification.CreatedAt, *notification.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(notification.UpdatedAt)
}

func (s *NotificationServiceIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.notificationService.Create(s.ctx, s.notification))

	err := s.notificationService.Delete(s.ctx, s.notification.ID, s.notification.Recipient)
	s.Require().NoError(err)

	_, err = s.notificationService.Get(s.ctx, s.notification.ID, s.notification.Recipient)
	s.Require().ErrorIs(err, repository.ErrNotFound)
}

func TestNotificationServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceIntegrationTestSuite))
}
