//go:build integration

package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	"github.com/stretchr/testify/suite"
)

type OrganizationRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	createOpts repository.CreateOrganizationOpts
}

func (s *OrganizationRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *OrganizationRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateOrganizationOpts(s.testUser.ID)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestCreate() {
	org, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeOrganization), org.ID)
	s.Assert().NotNil(org.CreatedAt)
	s.Assert().Nil(org.UpdatedAt)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	org, err := s.OrganizationRepo.Get(context.Background(), created.ID, repository.OrganizationDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, org.ID)
	s.Assert().Equal(s.createOpts.Name, org.Name)
	s.Assert().Equal(s.createOpts.Email, org.Email)
	s.Assert().Equal(s.createOpts.Logo, org.Logo)
	s.Assert().Equal(s.createOpts.Website, org.Website)
	s.Assert().Equal(s.createOpts.Status, org.Status)
	s.Require().NotNil(org.MemberCount)
	s.Assert().Equal(int64(1), *org.MemberCount)
	s.Assert().WithinDuration(*created.CreatedAt, *org.CreatedAt, 100*time.Millisecond)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestGetAll() {
	_, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)

	orgs, err := s.OrganizationRepo.List(context.Background(), s.testUser.ID, repository.CursorPage{Size: 10}, repository.OrganizationListProjection())
	s.Require().NoError(err)
	s.Require().Len(orgs.Items, 3)

	orgs, err = s.OrganizationRepo.List(context.Background(), s.testUser.ID, repository.CursorPage{Size: 2}, repository.OrganizationListProjection())
	s.Require().NoError(err)
	s.Require().Len(orgs.Items, 2)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	newEmail := testutil.GenerateEmail(10)
	org, err := s.OrganizationRepo.Update(context.Background(), created.ID, repository.UpdateOrganizationOpts{
		Name:  optional.Some("new name"),
		Email: optional.Some(newEmail),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", org.Name)
	s.Assert().Equal(newEmail, org.Email)
	s.Require().NotNil(org.MemberCount)
	s.Assert().Equal(int64(1), *org.MemberCount)
	s.Assert().NotNil(org.UpdatedAt)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestAddMember() {
	created, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), created.ID, member.ID))
	org, err := s.OrganizationRepo.Get(context.Background(), created.ID, repository.OrganizationDetailProjection())
	s.Require().NoError(err)
	s.Require().NotNil(org.MemberCount)
	s.Assert().Equal(int64(2), *org.MemberCount)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestRemoveMember() {
	created, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddMember(context.Background(), created.ID, member.ID))
	s.Require().NoError(s.OrganizationRepo.RemoveMember(context.Background(), created.ID, s.testUser.ID))
	org, err := s.OrganizationRepo.Get(context.Background(), created.ID, repository.OrganizationDetailProjection())
	s.Require().NoError(err)
	s.Require().NotNil(org.MemberCount)
	s.Assert().Equal(int64(1), *org.MemberCount)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.Delete(context.Background(), created.ID))
	_, err = s.OrganizationRepo.Get(context.Background(), created.ID, repository.OrganizationDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestAddInvitation() {
	created, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	invitedUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddInvitation(context.Background(), created.ID, invitedUser.ID))
	invitations, err := s.OrganizationRepo.GetInvitations(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)
	s.Assert().Equal(invitedUser.ID, invitations[0].ID)
	org, err := s.OrganizationRepo.Get(context.Background(), created.ID, repository.OrganizationDetailProjection())
	s.Require().NoError(err)
	s.Require().NotNil(org.MemberCount)
	s.Assert().Equal(int64(1), *org.MemberCount)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestAddInvitationWithInvalidOrgID() {
	invalidID := model.MustNewNilID(model.ResourceTypeOrganization)
	invitedUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	err = s.OrganizationRepo.AddInvitation(context.Background(), invalidID, invitedUser.ID)
	s.Assert().Error(err)
	s.Assert().ErrorIs(err, repository.ErrOrganizationAddMember)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestAddInvitationWithInvalidUserID() {
	created, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	invalidID := model.MustNewNilID(model.ResourceTypeUser)
	err = s.OrganizationRepo.AddInvitation(context.Background(), created.ID, invalidID)
	s.Assert().Error(err)
	s.Assert().ErrorIs(err, repository.ErrOrganizationAddMember)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestRemoveInvitation() {
	created, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	invitedUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddInvitation(context.Background(), created.ID, invitedUser.ID))
	s.Require().NoError(s.OrganizationRepo.RemoveInvitation(context.Background(), created.ID, invitedUser.ID))
	invitations, err := s.OrganizationRepo.GetInvitations(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 0)
}

func (s *OrganizationRepositoryIntegrationTestSuite) TestGetInvitations() {
	created, err := s.OrganizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	u1, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	u2, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.OrganizationRepo.AddInvitation(context.Background(), created.ID, u1.ID))
	s.Require().NoError(s.OrganizationRepo.AddInvitation(context.Background(), created.ID, u2.ID))
	invitations, err := s.OrganizationRepo.GetInvitations(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 2)
}

func TestOrganizationRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationRepositoryIntegrationTestSuite))
}

type CachedOrganizationRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser         *repository.User
	createOpts       repository.CreateOrganizationOpts
	organizationRepo *repository.RedisCachedOrganizationRepository
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.organizationRepo, _ = repository.NewCachedOrganizationRepository(s.OrganizationRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateOrganizationOpts(s.testUser.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestCreate() {
	org, err := s.organizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotNil(org.CreatedAt)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.organizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.OrganizationRepo.Get(context.Background(), created.ID, repository.OrganizationDetailProjection())
	s.Require().NoError(err)
	usingCache, err := s.organizationRepo.Get(context.Background(), created.ID, repository.OrganizationDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestGetAll() {
	_, err := s.organizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.OrganizationRepo.List(context.Background(), s.testUser.ID, repository.CursorPage{Size: 10}, repository.OrganizationListProjection())
	s.Require().NoError(err)
	usingCache, err := s.organizationRepo.List(context.Background(), s.testUser.ID, repository.CursorPage{Size: 10}, repository.OrganizationListProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.organizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	org, err := s.organizationRepo.Update(context.Background(), created.ID, repository.UpdateOrganizationOpts{
		Name: optional.Some("new name"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", org.Name)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 1)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.organizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.organizationRepo.Get(context.Background(), created.ID, repository.OrganizationDetailProjection())
	s.Require().NoError(err)
	s.Require().NoError(s.organizationRepo.Delete(context.Background(), created.ID))
	_, err = s.organizationRepo.Get(context.Background(), created.ID, repository.OrganizationDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestAddInvitation() {
	created, err := s.organizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	invitedUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.organizationRepo.AddInvitation(context.Background(), created.ID, invitedUser.ID))
	invitations, err := s.organizationRepo.GetInvitations(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestRemoveInvitation() {
	created, err := s.organizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	invitedUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.organizationRepo.AddInvitation(context.Background(), created.ID, invitedUser.ID))
	s.Require().NoError(s.organizationRepo.RemoveInvitation(context.Background(), created.ID, invitedUser.ID))
	invitations, err := s.organizationRepo.GetInvitations(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 0)
}

func (s *CachedOrganizationRepositoryIntegrationTestSuite) TestGetInvitations() {
	created, err := s.organizationRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	u1, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.organizationRepo.AddInvitation(context.Background(), created.ID, u1.ID))
	invitations, err := s.organizationRepo.GetInvitations(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Require().Len(invitations, 1)
}

func TestCachedOrganizationRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedOrganizationRepositoryIntegrationTestSuite))
}
