package repository_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil"
)

type PluginRepositoryIntegrationTestSuite struct {
	testutil.ContainerIntegrationTestSuite
	testutil.PgContainerIntegrationTestSuite

	pluginID string
	scope    model.ID
}

func (s *PluginRepositoryIntegrationTestSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping integration test")
	}
	s.SetupPg(&s.ContainerIntegrationTestSuite, reflect.TypeOf(s).Elem().String())
}

func (s *PluginRepositoryIntegrationTestSuite) SetupTest() {
	s.pluginID = "com.elemo.timetracking"
	s.scope = model.MustNewID(model.ResourceTypeOrganization)
}

func (s *PluginRepositoryIntegrationTestSuite) TearDownTest() {
	s.CleanupPg(&s.ContainerIntegrationTestSuite)
}

func (s *PluginRepositoryIntegrationTestSuite) TearDownSuite() {
	s.CleanupContainers()
}

func testPluginManifest() model.PluginManifest {
	return model.PluginManifest{
		SchemaVersion: model.PluginSchemaVersionV1,
		ID:            "com.elemo.timetracking",
		Name:          "Time Tracking",
		Version:       "1.0.0",
		Requires:      model.PluginRequires{PluginAPI: "^1"},
		Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
		Capabilities:  []model.PluginCapability{model.CapabilityIssuesRead},
	}
}

func (s *PluginRepositoryIntegrationTestSuite) TestInstallationCRUD() {
	ctx := context.Background()
	manifest := testPluginManifest()

	created, err := s.PluginRepo.UpsertInstallation(ctx, &model.PluginInstallation{
		PluginID: manifest.ID,
		Version:  manifest.Version,
		Manifest: manifest,
	})
	s.Require().NoError(err)
	s.NotEmpty(created.ID)
	s.Equal(model.PluginStatusInstalled, created.Status)
	s.NotNil(created.CreatedAt)

	got, err := s.PluginRepo.GetInstallation(ctx, manifest.ID)
	s.Require().NoError(err)
	s.Equal(created.ID, got.ID)
	s.Equal(manifest.Version, got.Version)
	s.Equal(manifest.Name, got.Manifest.Name)

	listed, err := s.PluginRepo.ListInstallations(ctx)
	s.Require().NoError(err)
	s.Len(listed, 1)

	updated, err := s.PluginRepo.UpsertInstallation(ctx, &model.PluginInstallation{
		ID:       created.ID,
		PluginID: manifest.ID,
		Version:  "1.1.0",
		Status:   model.PluginStatusActive,
		Manifest: manifest,
	})
	s.Require().NoError(err)
	s.Equal("1.1.0", updated.Version)

	s.Require().NoError(s.PluginRepo.DeleteInstallation(ctx, manifest.ID))
	_, err = s.PluginRepo.GetInstallation(ctx, manifest.ID)
	s.ErrorIs(err, repository.ErrNotFound)
}

func (s *PluginRepositoryIntegrationTestSuite) TestInstallationNotFound() {
	ctx := context.Background()
	_, err := s.PluginRepo.GetInstallation(ctx, "com.elemo.missing")
	s.ErrorIs(err, repository.ErrNotFound)

	err = s.PluginRepo.DeleteInstallation(ctx, "com.elemo.missing")
	s.ErrorIs(err, repository.ErrNotFound)
}

func (s *PluginRepositoryIntegrationTestSuite) TestActivationCRUD() {
	ctx := context.Background()
	manifest := testPluginManifest()
	_, err := s.PluginRepo.UpsertInstallation(ctx, &model.PluginInstallation{
		PluginID: manifest.ID,
		Version:  manifest.Version,
		Manifest: manifest,
	})
	s.Require().NoError(err)

	cfg := json.RawMessage(`{"k":"v"}`)
	act, err := s.PluginRepo.UpsertActivation(ctx, &model.PluginActivation{
		PluginID: manifest.ID,
		ScopeID:  s.scope,
		Enabled:  true,
		Config:   cfg,
	})
	s.Require().NoError(err)
	s.True(act.Enabled)

	got, err := s.PluginRepo.GetActivation(ctx, manifest.ID, s.scope)
	s.Require().NoError(err)
	s.Equal(s.scope, got.ScopeID)
	s.JSONEq(`{"k":"v"}`, string(got.Config))

	listed, err := s.PluginRepo.ListActivations(ctx, manifest.ID)
	s.Require().NoError(err)
	s.Len(listed, 1)

	byScope, err := s.PluginRepo.ListActivationsByScope(ctx, []model.ID{s.scope})
	s.Require().NoError(err)
	s.Len(byScope, 1)

	disabled, err := s.PluginRepo.UpsertActivation(ctx, &model.PluginActivation{
		PluginID: manifest.ID,
		ScopeID:  s.scope,
		Enabled:  false,
		Config:   cfg,
	})
	s.Require().NoError(err)
	s.False(disabled.Enabled)

	byScope, err = s.PluginRepo.ListActivationsByScope(ctx, []model.ID{s.scope})
	s.Require().NoError(err)
	s.Empty(byScope)

	s.Require().NoError(s.PluginRepo.DeleteActivations(ctx, manifest.ID))
	_, err = s.PluginRepo.GetActivation(ctx, manifest.ID, s.scope)
	s.ErrorIs(err, repository.ErrNotFound)
}

func (s *PluginRepositoryIntegrationTestSuite) TestStorageCRUD() {
	ctx := context.Background()
	manifest := testPluginManifest()
	_, err := s.PluginRepo.UpsertInstallation(ctx, &model.PluginInstallation{
		PluginID: manifest.ID,
		Version:  manifest.Version,
		Manifest: manifest,
	})
	s.Require().NoError(err)

	saved, err := s.PluginRepo.SetStorage(ctx, &model.PluginStorageEntry{
		PluginID: manifest.ID,
		ScopeID:  s.scope,
		Key:      "timer",
		Value:    []byte(`{"running":true}`),
	})
	s.Require().NoError(err)
	s.Equal("timer", saved.Key)

	got, err := s.PluginRepo.GetStorage(ctx, manifest.ID, s.scope, "timer")
	s.Require().NoError(err)
	s.JSONEq(`{"running":true}`, string(got.Value))

	listed, err := s.PluginRepo.ListStorage(ctx, manifest.ID, s.scope)
	s.Require().NoError(err)
	s.Len(listed, 1)

	s.Require().NoError(s.PluginRepo.DeleteStorage(ctx, manifest.ID, s.scope, "timer"))
	_, err = s.PluginRepo.GetStorage(ctx, manifest.ID, s.scope, "timer")
	s.ErrorIs(err, repository.ErrNotFound)

	_, err = s.PluginRepo.SetStorage(ctx, &model.PluginStorageEntry{
		PluginID: manifest.ID,
		ScopeID:  s.scope,
		Key:      "keep",
		Value:    []byte(`{"ok":1}`),
	})
	s.Require().NoError(err)
	s.Require().NoError(s.PluginRepo.DeleteStorageForPlugin(ctx, manifest.ID))
	listed, err = s.PluginRepo.ListStorage(ctx, manifest.ID, s.scope)
	s.Require().NoError(err)
	s.Empty(listed)
}

func (s *PluginRepositoryIntegrationTestSuite) TestStorageNotFound() {
	ctx := context.Background()
	_, err := s.PluginRepo.GetStorage(ctx, s.pluginID, s.scope, "missing")
	s.ErrorIs(err, repository.ErrNotFound)
	err = s.PluginRepo.DeleteStorage(ctx, s.pluginID, s.scope, "missing")
	s.ErrorIs(err, repository.ErrNotFound)
}

func TestPluginRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(PluginRepositoryIntegrationTestSuite))
}
