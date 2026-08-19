//go:build integration

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

	testUser        *repository.User
	testUserContext context.Context
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
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testUserContext = context.WithValue(context.Background(), pkg.CtxKeyUserID, s.testUser.ID)
	s.Require().NoError(testRepo.MakeUserSystemOwner(s.testUser.ID, s.Neo4jDB))
}

func (s *NotificationServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *NotificationServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *NotificationServiceIntegrationTestSuite) TestCreate() {
	notification, err := s.notificationService.Create(s.testUserContext, service.CreateNotificationOpts{
		Title:       "test notification",
		Description: "test description text",
		Recipient:   s.testUser.ID,
	})
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeNotification), notification.ID)
	s.Assert().NotNil(notification.CreatedAt)
}

func (s *NotificationServiceIntegrationTestSuite) TestGet() {
	created, err := s.notificationService.Create(s.testUserContext, service.CreateNotificationOpts{
		Title:       "test notification",
		Description: "test description text",
		Recipient:   s.testUser.ID,
	})
	s.Require().NoError(err)

	notification, err := s.notificationService.Get(s.testUserContext, created.ID, s.testUser.ID)
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, notification.ID)
	s.Assert().WithinDuration(*created.CreatedAt, *notification.CreatedAt, 100*time.Millisecond)
}

func (s *NotificationServiceIntegrationTestSuite) TestGetAllByRecipient() {
	_, err := s.notificationService.Create(s.testUserContext, service.CreateNotificationOpts{
		Title: "n1 title", Description: "n1 description text", Recipient: s.testUser.ID,
	})
	s.Require().NoError(err)
	_, err = s.notificationService.Create(s.testUserContext, service.CreateNotificationOpts{
		Title: "n2 title", Description: "n2 description text", Recipient: s.testUser.ID,
	})
	s.Require().NoError(err)

	notifications, err := s.notificationService.ListByRecipient(s.testUserContext, s.testUser.ID, service.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Assert().Len(notifications.Items, 2)
}

func (s *NotificationServiceIntegrationTestSuite) TestUpdate() {
	created, err := s.notificationService.Create(s.testUserContext, service.CreateNotificationOpts{
		Title: "test notification", Description: "test description text", Recipient: s.testUser.ID,
	})
	s.Require().NoError(err)

	notification, err := s.notificationService.Update(s.testUserContext, created.ID, s.testUser.ID, service.UpdateNotificationOpts{Read: true})
	s.Require().NoError(err)
	s.Require().True(notification.Read)
}

func (s *NotificationServiceIntegrationTestSuite) TestDelete() {
	created, err := s.notificationService.Create(s.testUserContext, service.CreateNotificationOpts{
		Title: "test notification", Description: "test description text", Recipient: s.testUser.ID,
	})
	s.Require().NoError(err)

	s.Require().NoError(s.notificationService.Delete(s.testUserContext, created.ID, s.testUser.ID))
	_, err = s.notificationService.Get(s.testUserContext, created.ID, s.testUser.ID)
	s.Assert().Error(err)
}

func TestNotificationServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceIntegrationTestSuite))
}
