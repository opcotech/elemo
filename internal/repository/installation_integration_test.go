package repository_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
)

type InstallationRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	repo repository.InstallationRepository
}

func (s *InstallationRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, "InstallationRepositoryIntegrationTestSuite")
	var err error
	s.repo, err = repository.NewNeo4jInstallationRepository(repository.WithNeo4jDatabase(s.Neo4jDB))
	s.Require().NoError(err)
}

func (s *InstallationRepositoryIntegrationTestSuite) SetupTest() {
	s.BootstrapNeo4jDatabase(&s.ContainerIntegrationTestSuite)
}

func (s *InstallationRepositoryIntegrationTestSuite) TearDownTest() {
	s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *InstallationRepositoryIntegrationTestSuite) TearDownSuite() {
	s.CleanupContainers()
}

func (s *InstallationRepositoryIntegrationTestSuite) TestBackfillIsStable() {
	ctx := context.Background()
	s.Require().NoError(repository.Neo4jExecuteWriteAndConsume(ctx, s.Neo4jDB, `
		MATCH (i:Installation {id: $id})
		REMOVE i.logical_id
	`, map[string]any{"id": model.InstallationID().String()}))

	first, err := s.repo.EnsureLogicalID(ctx)
	s.Require().NoError(err)
	_, err = uuid.Parse(first)
	s.Require().NoError(err)
	second, err := s.repo.EnsureLogicalID(ctx)
	s.Require().NoError(err)
	s.Equal(first, second)
}

func (s *InstallationRepositoryIntegrationTestSuite) TestConcurrentFirstBootReturnsOneID() {
	ctx := context.Background()
	s.Require().NoError(repository.Neo4jExecuteWriteAndConsume(ctx, s.Neo4jDB, `
		MATCH (i:Installation {id: $id})
		REMOVE i.logical_id
	`, map[string]any{"id": model.InstallationID().String()}))

	const workers = 8
	ids := make(chan string, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			id, err := s.repo.EnsureLogicalID(ctx)
			ids <- id
			errs <- err
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		s.Require().NoError(err)
	}
	var expected string
	for id := range ids {
		if expected == "" {
			expected = id
		}
		s.Equal(expected, id)
	}
}

func (s *InstallationRepositoryIntegrationTestSuite) TestMissingInstallation() {
	ctx := context.Background()
	s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	_, err := s.repo.EnsureLogicalID(ctx)
	s.Require().ErrorIs(err, repository.ErrNotFound)
}

func TestInstallationRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(InstallationRepositoryIntegrationTestSuite))
}
