package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

type LicenseServiceIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.Neo4jContainerIntegrationTestSuite

	license        *license.License
	licenseService service.LicenseService
}

func (s *LicenseServiceIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	container := reflect.TypeOf(s).Elem().String()
	s.SetupNeo4j(&s.ContainerIntegrationTestSuite, container)

	s.license = testutil.ParseLicense(s.T())

	permissionService, err := service.NewPermissionService(s.PermissionRepo)
	s.Require().NoError(err)

	s.licenseService, err = service.NewLicenseService(
		s.license,
		s.LicenseRepo,
		service.WithPermissionService(permissionService),
	)
	s.Require().NoError(err)
}

func (s *LicenseServiceIntegrationTestSuite) TearDownTest() {
	defer s.CleanupNeo4j(&s.ContainerIntegrationTestSuite)
}

func (s *LicenseServiceIntegrationTestSuite) TearDownSuite() {
	defer s.CleanupContainers()
}

func (s *LicenseServiceIntegrationTestSuite) TestExpired() {
	expired, err := s.licenseService.Expired(context.Background())
	s.Require().NoError(err)
	s.Require().False(expired)
}

func (s *LicenseServiceIntegrationTestSuite) TestHasFeature() {
	hasFeature, err := s.licenseService.HasFeature(context.Background(), license.FeatureReleases)
	s.Require().NoError(err)
	s.Require().True(hasFeature)
}

func (s *LicenseServiceIntegrationTestSuite) TestWithinThreshold() {
	withinThreshold, err := s.licenseService.WithinThreshold(context.Background(), license.QuotaUsers)
	s.Require().NoError(err)
	s.Require().True(withinThreshold)
}

func (s *LicenseServiceIntegrationTestSuite) TestGetLicense() {
	user := testModel.NewUser()
	s.Require().NoError(s.UserRepo.Create(context.Background(), user))
	s.Require().NoError(testRepo.MakeUserSystemOwner(user.ID, s.Neo4jDB))

	retrievedLicense, err := s.licenseService.GetLicense(context.WithValue(context.Background(), pkg.CtxKeyUserID, user.ID))
	s.Require().NoError(err)
	s.Require().Equal(s.license, &retrievedLicense)
}

func (s *LicenseServiceIntegrationTestSuite) TestPing() {
	s.Require().NoError(s.licenseService.Ping(context.Background()))
}

func TestLicenseServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LicenseServiceIntegrationTestSuite))
}
