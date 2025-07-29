package s3_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/opcotech/elemo/internal/pkg"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
)

type StaticFileRepositoryIntegrationTestSuite struct {
	testutil.ConfigurationTestSuite
	testutil.ContainerIntegrationTestSuite
	testutil.LocalStackContainerIntegrationTestSuite

	staticFilePath string
	staticFile     []byte
}

func (s *StaticFileRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupLocalStack(&s.ContainerIntegrationTestSuite, container)
}

func (s *StaticFileRepositoryIntegrationTestSuite) SetupTest() {
	s.staticFilePath = pkg.GenerateRandomString(10) + ".png"

	qrCode, err := qrcode.New("https://elemo.click")
	s.Require().NoError(err)

	buf := new(pkg.WriteCloserBuffer)
	writer := standard.NewWithWriter(buf)
	s.Require().NoError(qrCode.Save(writer))

	s.staticFile = buf.Bytes()
}

func (s *StaticFileRepositoryIntegrationTestSuite) TearDownTest() {
	defer s.CleanupLocalStack(&s.ContainerIntegrationTestSuite)
}

func (s *StaticFileRepositoryIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *StaticFileRepositoryIntegrationTestSuite) TestCreate() {
	s.Require().NoError(s.StaticFileRepository.Create(context.Background(), s.staticFilePath, s.staticFile))
}

func (s *StaticFileRepositoryIntegrationTestSuite) TestGet() {
	s.Require().NoError(s.StaticFileRepository.Create(context.Background(), s.staticFilePath, s.staticFile))

	data, err := s.StaticFileRepository.Get(context.Background(), s.staticFilePath)
	s.Require().NoError(err)

	s.Assert().ElementsMatch(s.staticFile, data)
}

func (s *StaticFileRepositoryIntegrationTestSuite) TestUpdate() {
	// First create a file
	s.Require().NoError(s.StaticFileRepository.Create(context.Background(), s.staticFilePath, s.staticFile))

	// Create new data for update
	newData := []byte("updated file content")

	// Update the file
	s.Require().NoError(s.StaticFileRepository.Update(context.Background(), s.staticFilePath, newData))

	// Verify the file was updated
	data, err := s.StaticFileRepository.Get(context.Background(), s.staticFilePath)
	s.Require().NoError(err)
	s.Assert().ElementsMatch(newData, data)
}

func (s *StaticFileRepositoryIntegrationTestSuite) TestUpdateNonExistentFile() {
	// Try to update a file that doesn't exist
	newData := []byte("updated file content")

	// Update should succeed (S3 PutObject will create the file if it doesn't exist)
	s.Require().NoError(s.StaticFileRepository.Update(context.Background(), "non-existent-file.txt", newData))

	// Verify the file was created
	data, err := s.StaticFileRepository.Get(context.Background(), "non-existent-file.txt")
	s.Require().NoError(err)
	s.Assert().ElementsMatch(newData, data)
}

func (s *StaticFileRepositoryIntegrationTestSuite) TestUpdateWithEmptyData() {
	// First create a file
	s.Require().NoError(s.StaticFileRepository.Create(context.Background(), s.staticFilePath, s.staticFile))

	// Update with empty data
	emptyData := []byte{}
	s.Require().NoError(s.StaticFileRepository.Update(context.Background(), s.staticFilePath, emptyData))

	// Verify the file was updated with empty data
	data, err := s.StaticFileRepository.Get(context.Background(), s.staticFilePath)
	s.Require().NoError(err)
	s.Assert().ElementsMatch(emptyData, data)
}

func (s *StaticFileRepositoryIntegrationTestSuite) TestDelete() {
	s.Require().NoError(s.StaticFileRepository.Create(context.Background(), s.staticFilePath, s.staticFile))

	s.Require().NoError(s.StaticFileRepository.Delete(context.Background(), s.staticFilePath))

	_, err := s.StaticFileRepository.Get(context.Background(), s.staticFilePath)
	s.Assert().ErrorIs(err, repository.ErrNotFound)
}

func TestStaticFileRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(StaticFileRepositoryIntegrationTestSuite))
}
