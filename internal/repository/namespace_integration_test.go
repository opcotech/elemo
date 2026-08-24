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

type NamespaceRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	testUser   *repository.User
	testOrg    *repository.Organization
	createOpts repository.CreateNamespaceOpts
}

func (s *NamespaceRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *NamespaceRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.PermissionRepo.Create(context.Background(), testModel.NewCreateGrantOpts(
		s.testUser.ID,
		s.testOrg.ID,
		testModel.OrgAdminActions()...,
	))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestCreate() {
	ns, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotEqual(model.MustNewNilID(model.ResourceTypeNamespace), ns.ID)
	s.Assert().NotNil(ns.CreatedAt)
	s.Assert().Nil(ns.UpdatedAt)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(created.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(created.ID, s.testUser.ID))
	s.Require().NoError(err)

	ns, err := s.NamespaceRepo.Get(context.Background(), created.ID, repository.NamespaceDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(created.ID, ns.ID)
	s.Assert().Equal(s.createOpts.Name, ns.Name)
	s.Assert().Equal(s.createOpts.Description, ns.Description)
	s.Assert().WithinDuration(*created.CreatedAt, *ns.CreatedAt, 100*time.Millisecond)
	s.Require().NotNil(ns.ProjectCount)
	s.Assert().Equal(int64(1), *ns.ProjectCount)
	s.Require().NotNil(ns.DocumentCount)
	s.Assert().Equal(int64(1), *ns.DocumentCount)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestList() {
	created, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.ProjectRepo.Create(context.Background(), testModel.NewCreateProjectOpts(created.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.DocumentRepo.Create(context.Background(), testModel.NewCreateDocumentOpts(created.ID, s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)
	_, err = s.NamespaceRepo.Create(context.Background(), testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID))
	s.Require().NoError(err)

	namespaces, err := s.NamespaceRepo.ListForOrganization(context.Background(), repository.NamespaceListQuery{OrgID: s.testOrg.ID, ActorID: s.testUser.ID, Page: repository.CursorPage{Size: 10}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()})
	s.Require().NoError(err)
	s.Assert().Len(namespaces.Items, 3)

	var withRelated *repository.Namespace
	for _, ns := range namespaces.Items {
		if ns.ID == created.ID {
			withRelated = ns
			break
		}
	}
	s.Require().NotNil(withRelated)
	s.Require().NotNil(withRelated.ProjectCount)
	s.Assert().Equal(int64(1), *withRelated.ProjectCount)
	s.Require().NotNil(withRelated.DocumentCount)
	s.Assert().Equal(int64(1), *withRelated.DocumentCount)

	namespaces, err = s.NamespaceRepo.ListForOrganization(context.Background(), repository.NamespaceListQuery{OrgID: s.testOrg.ID, ActorID: s.testUser.ID, Page: repository.CursorPage{Size: 2}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()})
	s.Require().NoError(err)
	s.Assert().Len(namespaces.Items, 2)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestListAccessible() {
	created, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)

	accessible, err := s.NamespaceRepo.ListAccessible(context.Background(), repository.NamespaceListAccessibleQuery{ActorID: s.testUser.ID, Page: repository.CursorPage{Size: 10}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()})
	s.Require().NoError(err)
	s.Require().Len(accessible.Items, 1)
	s.Assert().Equal(created.ID, accessible.Items[0].ID)
	s.Assert().Equal(s.testOrg.ID, accessible.Items[0].Organization.ID)
	s.Assert().Equal(s.testOrg.Name, accessible.Items[0].Organization.Name)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	ns, err := s.NamespaceRepo.Update(context.Background(), created.ID, repository.UpdateNamespaceOpts{
		Name:        optional.Some("new name"),
		Description: optional.Some("new description"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", ns.Name)
	s.Assert().Equal("new description", ns.Description)
	s.Assert().NotNil(ns.UpdatedAt)
}

func (s *NamespaceRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.NamespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Require().NoError(s.NamespaceRepo.Delete(context.Background(), created.ID))
	_, err = s.NamespaceRepo.Get(context.Background(), created.ID, repository.NamespaceDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestNamespaceRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceRepositoryIntegrationTestSuite))
}

type CachedNamespaceRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.RedisContainerIntegrationTestSuite

	testUser      *repository.User
	testOrg       *repository.Organization
	createOpts    repository.CreateNamespaceOpts
	namespaceRepo *repository.RedisCachedNamespaceRepository
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.SetupRedis(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
	s.namespaceRepo, _ = repository.NewCachedNamespaceRepository(s.NamespaceRepo, repository.WithRedisDatabase(s.RedisDB))
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) SetupTest() {
	var err error
	s.testUser, err = s.UserRepo.Create(context.Background(), testModel.NewCreateUserOpts())
	s.Require().NoError(err)
	s.testOrg, err = s.OrganizationRepo.Create(context.Background(), testModel.NewCreateOrganizationOpts(s.testUser.ID))
	s.Require().NoError(err)
	_, err = s.PermissionRepo.Create(context.Background(), testModel.NewCreateGrantOpts(
		s.testUser.ID,
		s.testOrg.ID,
		testModel.OrgAdminActions()...,
	))
	s.Require().NoError(err)
	s.createOpts = testModel.NewCreateNamespaceOpts(s.testUser.ID, s.testOrg.ID)
	s.Require().Len(s.Keys(&s.ContainerIntegrationTestSuite, "*"), 0)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupRedis(&s.ContainerIntegrationTestSuite)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestCreate() {
	ns, err := s.namespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	s.Assert().NotNil(ns.CreatedAt)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 0)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestGet() {
	created, err := s.namespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.NamespaceRepo.Get(context.Background(), created.ID, repository.NamespaceDetailProjection())
	s.Require().NoError(err)
	usingCache, err := s.namespaceRepo.Get(context.Background(), created.ID, repository.NamespaceDetailProjection())
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestList() {
	_, err := s.namespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	original, err := s.NamespaceRepo.ListForOrganization(context.Background(), repository.NamespaceListQuery{OrgID: s.testOrg.ID, ActorID: s.testUser.ID, Page: repository.CursorPage{Size: 10}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()})
	s.Require().NoError(err)
	usingCache, err := s.namespaceRepo.ListForOrganization(context.Background(), repository.NamespaceListQuery{OrgID: s.testOrg.ID, ActorID: s.testUser.ID, Page: repository.CursorPage{Size: 10}, Order: repository.SortDirectionDesc, Projection: repository.NamespaceListProjection()})
	s.Require().NoError(err)
	s.Assert().Equal(original, usingCache)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestUpdate() {
	created, err := s.namespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	ns, err := s.namespaceRepo.Update(context.Background(), created.ID, repository.UpdateNamespaceOpts{
		Name: optional.Some("new name"),
	})
	s.Require().NoError(err)
	s.Assert().Equal("new name", ns.Name)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 1)
}

func (s *CachedNamespaceRepositoryIntegrationTestSuite) TestDelete() {
	created, err := s.namespaceRepo.Create(context.Background(), s.createOpts)
	s.Require().NoError(err)
	_, err = s.namespaceRepo.Get(context.Background(), created.ID, repository.NamespaceDetailProjection())
	s.Require().NoError(err)
	s.Require().NoError(s.namespaceRepo.Delete(context.Background(), created.ID))
	_, err = s.namespaceRepo.Get(context.Background(), created.ID, repository.NamespaceDetailProjection())
	s.Assert().ErrorIs(err, repository.ErrNotFound)
	s.Assert().Len(cacheKeysWithoutIssueListGeneration(s.Keys(&s.ContainerIntegrationTestSuite, "*")), 0)
}

func TestCachedNamespaceRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CachedNamespaceRepositoryIntegrationTestSuite))
}
