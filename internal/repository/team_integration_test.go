package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type TeamRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	createOpts repository.CreateTeamOpts
}

func (s *TeamRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *TeamRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateTeamOpts(s.testUser.ID, s.testOrg.ID)
}

func (s *TeamRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *TeamRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *TeamRepositoryIntegrationTestSuite) TestCreate() {
	team, err := s.TeamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeTeam), team.ID)
	s.Assert().NotNil(team.CreatedAt)
	s.Assert().Nil(team.UpdatedAt)
	s.Require().NotNil(team.MemberCount)
	s.Assert().Equal(int64(0), *team.MemberCount)
}

func (s *TeamRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.TeamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	team, err := s.TeamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, team.ID)
	s.Assert().Equal(s.createOpts.Name, team.Name)
	s.Assert().Equal(s.createOpts.Description, team.Description)
	s.Require().NotNil(team.MemberCount)
	s.Assert().Equal(int64(0), *team.MemberCount)
	s.Assert().WithinDuration(*created.CreatedAt, *team.CreatedAt, 100*time.Millisecond)
}

func (s *TeamRepositoryIntegrationTestSuite) TestListBelongsTo() {
	_, err := s.TeamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.TeamRepo.Create(context.Background(), testModel.NewCreateTeamOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)
	_, err = s.TeamRepo.Create(context.Background(), testModel.NewCreateTeamOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)

	otherOrg, err := s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.TeamRepo.Create(context.Background(), testModel.NewCreateTeamOpts(s.testUser.ID, otherOrg.ID))
	s.Require().NoError(err)

	teams, err := s.TeamRepo.ListBelongsTo(context.Background(), s.testOrg.ID, repository.CursorPage{Size: 10}, repository.TeamListProjection())
	s.Require().NoError(err)
	s.Assert().Len(teams.Items, 3)

	teams, err = s.TeamRepo.ListBelongsTo(context.Background(), s.testOrg.ID, repository.CursorPage{Size: 2}, repository.TeamListProjection())
	s.Require().NoError(err)
	s.Assert().Len(teams.Items, 2)

	otherTeams, err := s.TeamRepo.ListBelongsTo(context.Background(), otherOrg.ID, repository.CursorPage{Size: 10}, repository.TeamListProjection())
	s.Require().NoError(err)
	s.Assert().Len(otherTeams.Items, 1)
}

func (s *TeamRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.TeamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	team, err := s.TeamRepo.Update(context.Background(), created.ID, s.testOrg.ID, repository.UpdateTeamOpts{
		Name:        optional.Some("new name"),
		Description: optional.Some("new description"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", team.Name)
	s.Assert().Equal("new description", team.Description)
	s.Assert().NotNil(team.UpdatedAt)
}

func (s *TeamRepositoryIntegrationTestSuite) TestAddMember() {
	created, err := s.TeamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	newUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.Require().NoError(s.TeamRepo.AddMember(context.Background(), created.ID, newUser.ID, s.testOrg.ID))

	team, err := s.TeamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	s.Require().NotNil(team.MemberCount)
	s.Assert().Equal(int64(1), *team.MemberCount)

	members, err := s.TeamRepo.ListMembers(context.Background(), created.ID, s.testOrg.ID, repository.CursorPage{Size: 10})
	s.Require().NoError(err)
	s.Require().Len(members.Items, 1)
	s.Assert().Equal(newUser.ID, members.Items[0].ID)
}

func (s *TeamRepositoryIntegrationTestSuite) TestRemoveMember() {
	created, err := s.TeamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	other, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)

	s.Require().NoError(s.TeamRepo.AddMember(context.Background(), created.ID, member.ID, s.testOrg.ID))
	s.Require().NoError(s.TeamRepo.AddMember(context.Background(), created.ID, other.ID, s.testOrg.ID))
	s.Require().NoError(s.TeamRepo.RemoveMember(context.Background(), created.ID, member.ID, s.testOrg.ID))

	team, err := s.TeamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	s.Require().NotNil(team.MemberCount)
	s.Assert().Equal(int64(1), *team.MemberCount)
}

func (s *TeamRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.TeamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.TeamRepo.Delete(context.Background(), created.ID, s.testOrg.ID))
	_, err = s.TeamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestTeamRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TeamRepositoryIntegrationTestSuite))
}

type CachedTeamRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	createOpts repository.CreateTeamOpts
	teamRepo   *repository.RedisCachedTeamRepository
}

func (s *CachedTeamRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.teamRepo, _ = repository.NewCachedTeamRepository(s.TeamRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedTeamRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateTeamOpts(s.testUser.ID, s.testOrg.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedTeamRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedTeamRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedTeamRepositoryIntegrationTestSuite) TestCreate() {
	team, err := s.teamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotNil(team.CreatedAt)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 0)
}

func (s *CachedTeamRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.teamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.TeamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	usingCache, err := s.teamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
	cached, err := s.teamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original.ID, cached.ID)
	s.Assert().Equal(original.Name, cached.Name)
	s.Assert().Equal(original.Description, cached.Description)
}

func (s *CachedTeamRepositoryIntegrationTestSuite) TestListBelongsTo() {
	_, err := s.teamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.TeamRepo.ListBelongsTo(context.Background(), s.testOrg.ID, repository.CursorPage{Size: 10}, repository.TeamListProjection())
	s.Require().NoError(err)
	usingCache, err := s.teamRepo.ListBelongsTo(context.Background(), s.testOrg.ID, repository.CursorPage{Size: 10}, repository.TeamListProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
}

func (s *CachedTeamRepositoryIntegrationTestSuite) TestAddMember() {
	created, err := s.teamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	newUser, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.teamRepo.AddMember(context.Background(), created.ID, newUser.ID, s.testOrg.ID))
	team, err := s.teamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	s.Require().NotNil(team.MemberCount)
	s.Assert().Equal(int64(1), *team.MemberCount)
}

func (s *CachedTeamRepositoryIntegrationTestSuite) TestRemoveMember() {
	created, err := s.teamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	member, err := s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.Require().NoError(s.teamRepo.AddMember(context.Background(), created.ID, member.ID, s.testOrg.ID))
	s.Require().NoError(s.teamRepo.RemoveMember(context.Background(), created.ID, member.ID, s.testOrg.ID))
	team, err := s.teamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	s.Require().NotNil(team.MemberCount)
	s.Assert().Equal(int64(0), *team.MemberCount)
}

func (s *CachedTeamRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.teamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.teamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
	team, err := s.teamRepo.Update(context.Background(), created.ID, s.testOrg.ID, repository.UpdateTeamOpts{
		Name: optional.Some("new name"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", team.Name)
	updated, err := s.teamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal("new name", updated.Name)
}

func (s *CachedTeamRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.teamRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.teamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Require().NoError(err)
	s.Require().NoError(s.teamRepo.Delete(context.Background(), created.ID, s.testOrg.ID))
	_, err = s.teamRepo.Get(context.Background(), created.ID, s.testOrg.ID, repository.TeamDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestCachedTeamRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedTeamRepositoryIntegrationTestSuite))
}
