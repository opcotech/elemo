package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
)

type StaticFileServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite
	testutil.LocalStackContainerIntegrationTestSuite

	staticFileService service.StaticFileService
	staticFilePath    string
	staticFile        []byte
}

func (s *StaticFileServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)
	s.SetupLocalStack(&s.ContainerIntegrationTestSuite, container)

	permissionService, err := service.NewPermissionService(s.PermissionRepo)
	s.Require().NoError(err)

	licenseService, err := service.NewLicenseService(
		testutil.ParseLicense(s.T()),
		s.LicenseRepo,
		service.WithPermissionService(permissionService),
	)
	s.Require().NoError(err)

	s.staticFileService, err = service.NewStaticFileService(
		s.StaticFileRepository,
		service.WithLicenseService(licenseService),
	)
	s.Require().NoError(err)
}

func (s *StaticFileServiceIntegrationTestSuite) SetupTest() {
	s.staticFilePath = pkg.GenerateRandomString(10) + ".txt"
	s.staticFile = []byte("This is a test file content for static file service testing.")
}

func (s *StaticFileServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
	defer s.CleanupLocalStack(&s.ContainerIntegrationTestSuite)
}

func (s *StaticFileServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *StaticFileServiceIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.staticFileService.Create(context.Background(), s.staticFilePath, s.staticFile))
}

func (s *StaticFileServiceIntegrationTestSuite) TestCreateEmptyData() {
	s.Require().NoError(s.staticFileService.Create(context.Background(), s.staticFilePath, []byte{}))

	data, err := s.staticFileService.Get(context.Background(), s.staticFilePath)
	s.Require().NoError(err)
	s.Assert().Empty(data)
}

func (s *StaticFileServiceIntegrationTestSuite) TestCreateEmptyPath() {
	err := s.staticFileService.Create(context.Background(), "", s.staticFile)
	s.Require().ErrorIs(err, service.ErrStaticFileInvalidPath)
	s.Require().ErrorIs(err, service.ErrStaticFileCreate)
}

func (s *StaticFileServiceIntegrationTestSuite) TestCreateAbsolutePath() {
	err := s.staticFileService.Create(context.Background(), "/etc/passwd", s.staticFile)
	s.Require().ErrorIs(err, service.ErrStaticFileInvalidPath)
	s.Require().ErrorIs(err, service.ErrStaticFileCreate)
}

func (s *StaticFileServiceIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.staticFileService.Create(context.Background(), s.staticFilePath, s.staticFile))

	data, err := s.staticFileService.Get(context.Background(), s.staticFilePath)
	s.Require().NoError(err)
	s.Assert().Equal(s.staticFile, data)
}

func (s *StaticFileServiceIntegrationTestSuite) TestGetNormalizedPath() {
	s.Require().NoError(s.staticFileService.Create(context.Background(), "dir/file.txt", s.staticFile))

	data, err := s.staticFileService.Get(context.Background(), "dir/../dir/file.txt")
	s.Require().NoError(err)
	s.Assert().Equal(s.staticFile, data)
}

func (s *StaticFileServiceIntegrationTestSuite) TestGetMissingFile() {
	_, err := s.staticFileService.Get(context.Background(), "missing-file.txt")
	s.Require().ErrorIs(err, service.ErrStaticFileGet)
	s.Require().ErrorIs(err, repository.ErrNotFound)
}

func (s *StaticFileServiceIntegrationTestSuite) TestGetEmptyPath() {
	_, err := s.staticFileService.Get(context.Background(), "")
	s.Require().ErrorIs(err, service.ErrStaticFileInvalidPath)
	s.Require().ErrorIs(err, service.ErrStaticFileGet)
}

func (s *StaticFileServiceIntegrationTestSuite) TestGetAbsolutePath() {
	_, err := s.staticFileService.Get(context.Background(), "/etc/passwd")
	s.Require().ErrorIs(err, service.ErrStaticFileInvalidPath)
	s.Require().ErrorIs(err, service.ErrStaticFileGet)
}

func (s *StaticFileServiceIntegrationTestSuite) TestUpdate() {
	s.Require().NoError(s.staticFileService.Create(context.Background(), s.staticFilePath, s.staticFile))

	newData := []byte("updated file content")
	s.Require().NoError(s.staticFileService.Update(context.Background(), s.staticFilePath, newData))

	data, err := s.staticFileService.Get(context.Background(), s.staticFilePath)
	s.Require().NoError(err)
	s.Assert().Equal(newData, data)
}

func (s *StaticFileServiceIntegrationTestSuite) TestUpdateEmptyData() {
	s.Require().NoError(s.staticFileService.Create(context.Background(), s.staticFilePath, s.staticFile))
	s.Require().NoError(s.staticFileService.Update(context.Background(), s.staticFilePath, []byte{}))

	data, err := s.staticFileService.Get(context.Background(), s.staticFilePath)
	s.Require().NoError(err)
	s.Assert().Empty(data)
}

func (s *StaticFileServiceIntegrationTestSuite) TestUpdateEmptyPath() {
	err := s.staticFileService.Update(context.Background(), "", s.staticFile)
	s.Require().ErrorIs(err, service.ErrStaticFileInvalidPath)
	s.Require().ErrorIs(err, service.ErrStaticFileUpdate)
}

func (s *StaticFileServiceIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.staticFileService.Create(context.Background(), s.staticFilePath, s.staticFile))

	s.Require().NoError(s.staticFileService.Delete(context.Background(), s.staticFilePath))

	_, err := s.staticFileService.Get(context.Background(), s.staticFilePath)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func (s *StaticFileServiceIntegrationTestSuite) TestDeleteEmptyPath() {
	err := s.staticFileService.Delete(context.Background(), "")
	s.Require().ErrorIs(err, service.ErrStaticFileInvalidPath)
	s.Require().ErrorIs(err, service.ErrStaticFileDelete)
}

func TestStaticFileServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(StaticFileServiceIntegrationTestSuite))
}
