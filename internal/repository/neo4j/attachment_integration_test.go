package neo4j_test

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

type AttachmentRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *model.User
	testOrg    *model.Organization
	testDoc    *model.Document
	attachment *model.Attachment
}

func (s *AttachmentRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *AttachmentRepositoryIntegrationTestSuite) SetupTest() {
	s.testUser = testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), s.testUser))

	s.testOrg = testModel.NewOrganization()
	s.Require().NoError(s.OrganizationRepo.Create(context.Background(), s.testUser.ID, s.testOrg))

	s.testDoc = testModel.NewDocument(s.testUser.ID)
	s.Require().NoError(s.DocumentRepo.Create(context.Background(), s.testUser.ID, s.testDoc))

	s.attachment = testModel.NewAttachment(s.testUser.ID)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *AttachmentRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.AttachmentRepo.Create(context.Background(), s.testDoc.ID, s.attachment))
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeAttachment), s.attachment.ID)
	s.Assert().NotNil(s.attachment.CreatedAt)
	s.Assert().Nil(s.attachment.UpdatedAt)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.AttachmentRepo.Create(context.Background(), s.testDoc.ID, s.attachment))

	attachment, err := s.AttachmentRepo.Get(context.Background(), s.attachment.ID)
	s.Require().NoError(err)

	s.Assert().Equal(s.attachment.ID, attachment.ID)
	s.Assert().Equal(s.attachment.Name, attachment.Name)
	s.Assert().Equal(s.attachment.FileID, attachment.FileID)
	s.Assert().Equal(s.attachment.CreatedBy, attachment.CreatedBy)
	s.Assert().WithinDuration(*s.attachment.CreatedAt, *attachment.CreatedAt, 100*time.Millisecond)
	s.Assert().Nil(attachment.UpdatedAt)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TestGetAllBelongsTo() {
	s.Require().NoError(s.AttachmentRepo.Create(context.Background(), s.testDoc.ID, s.attachment))
	s.Require().NoError(s.AttachmentRepo.Create(context.Background(), s.testDoc.ID, testModel.NewAttachment(s.testUser.ID)))
	s.Require().NoError(s.AttachmentRepo.Create(context.Background(), s.testDoc.ID, testModel.NewAttachment(s.testUser.ID)))

	attachments, err := s.AttachmentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 0, 10)
	s.Require().NoError(err)
	s.Assert().Len(attachments, 3)

	attachments, err = s.AttachmentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 1, 2)
	s.Require().NoError(err)
	s.Assert().Len(attachments, 2)

	attachments, err = s.AttachmentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 2, 2)
	s.Require().NoError(err)
	s.Assert().Len(attachments, 1)

	attachments, err = s.AttachmentRepo.GetAllBelongsTo(context.Background(), s.testDoc.ID, 3, 2)
	s.Require().NoError(err)
	s.Assert().Len(attachments, 0)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.AttachmentRepo.Create(context.Background(), s.testDoc.ID, s.attachment))

	newName := "new name"
	attachment, err := s.AttachmentRepo.Update(context.Background(), s.attachment.ID, newName)
	s.Require().NoError(err)

	s.Assert().Equal(s.attachment.ID, attachment.ID)
	s.Assert().Equal(newName, attachment.Name)
	s.Assert().Equal(s.attachment.FileID, attachment.FileID)
	s.Assert().Equal(s.attachment.CreatedBy, attachment.CreatedBy)
	s.Assert().WithinDuration(*s.attachment.CreatedAt, *attachment.CreatedAt, 100*time.Millisecond)
	s.Assert().NotNil(attachment.UpdatedAt)
}

func (s *AttachmentRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.AttachmentRepo.Create(context.Background(), s.testDoc.ID, s.attachment))

	s.Require().NoError(s.AttachmentRepo.Delete(context.Background(), s.attachment.ID))

	_, err := s.AttachmentRepo.Get(context.Background(), s.attachment.ID)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestAttachmentRepositoryIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(AttachmentRepositoryIntegrationTestSuite))
}
