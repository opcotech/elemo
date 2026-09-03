package service_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
)

type pluginLifecycleHarness struct {
	ctx     context.Context
	orgID   model.ID
	lic     *mocksvc.MockLicenseService
	perm    *mocksvc.MockPermissionService
	repo    *mockrepo.MockPluginRepository
	extRepo *mockrepo.MockExtensionRepository
	svc     service.PluginService
}

func newPluginLifecycleHarness(t *testing.T) pluginLifecycleHarness {
	t.Helper()
	ctrl := gomock.NewController(t)
	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0)).AnyTimes()
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Len(0)).Return(ctx, span).AnyTimes()

	logger := mocklog.NewMockLogger(ctrl)
	logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	lic := mocksvc.NewMockLicenseService(ctrl)
	perm := mocksvc.NewMockPermissionService(ctrl)
	repo := mockrepo.NewMockPluginRepository(ctrl)
	extRepo := mockrepo.NewMockExtensionRepository(ctrl)

	svc, err := service.NewPluginService(
		config.PluginConfig{Directory: t.TempDir()},
		repo,
		extRepo,
		perm,
		lic,
		nil, nil, nil, nil,
		service.WithLogger(logger),
		service.WithTracer(tracer),
	)
	require.NoError(t, err)
	service.SetPluginRegistry(svc, elemoplugin.NewRegistry(elemoplugin.NewHandlerRuntime()))
	return pluginLifecycleHarness{
		ctx:     ctx,
		orgID:   orgID,
		lic:     lic,
		perm:    perm,
		repo:    repo,
		extRepo: extRepo,
		svc:     svc,
	}
}

func frontendPluginZip(t *testing.T, manifest model.PluginManifest) []byte {
	t.Helper()
	require.NoError(t, manifest.Validate())
	yamlBytes, err := yaml.Marshal(manifest)
	require.NoError(t, err)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	mw, err := zw.Create("plugin.yaml")
	require.NoError(t, err)
	_, err = mw.Write(yamlBytes)
	require.NoError(t, err)
	fw, err := zw.Create("frontend/index.js")
	require.NoError(t, err)
	_, err = fw.Write([]byte("export default {}"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func TestPluginService_Install(t *testing.T) {
	t.Parallel()
	manifest := frontendPluginManifest()
	zipBytes := frontendPluginZip(t, manifest)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(true, nil)
		h.repo.EXPECT().GetInstallation(h.ctx, manifest.ID).Return(nil, repository.ErrNotFound)
		saved := &model.PluginInstallation{
			ID:       "inst-1",
			PluginID: manifest.ID,
			Version:  manifest.Version,
			Status:   model.PluginStatusInstalled,
			Manifest: manifest,
		}
		h.repo.EXPECT().UpsertInstallation(h.ctx, gomock.Any()).Return(saved, nil)

		got, err := h.svc.Install(h.ctx, zipBytes)
		require.NoError(t, err)
		assert.Equal(t, manifest.ID, got.PluginID)
	})

	t.Run("already installed", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(true, nil)
		h.repo.EXPECT().GetInstallation(h.ctx, manifest.ID).Return(&model.PluginInstallation{PluginID: manifest.ID}, nil)

		_, err := h.svc.Install(h.ctx, zipBytes)
		require.ErrorIs(t, err, repository.ErrPluginConflict)
	})

	t.Run("bad zip", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(true, nil)

		_, err := h.svc.Install(h.ctx, []byte("not-a-zip"))
		require.ErrorIs(t, err, service.ErrPluginInstall)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(false, nil)

		_, err := h.svc.Install(h.ctx, zipBytes)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestPluginService_Upgrade(t *testing.T) {
	t.Parallel()
	manifest := frontendPluginManifest()
	upgraded := manifest
	upgraded.Version = "1.1.0"
	zipBytes := frontendPluginZip(t, upgraded)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(true, nil)
		current := &model.PluginInstallation{
			PluginID: manifest.ID,
			Version:  manifest.Version,
			Status:   model.PluginStatusInstalled,
			Manifest: manifest,
		}
		h.repo.EXPECT().GetInstallation(h.ctx, manifest.ID).Return(current, nil)
		saved := *current
		saved.Version = upgraded.Version
		saved.Manifest = upgraded
		h.repo.EXPECT().UpsertInstallation(h.ctx, gomock.Any()).Return(&saved, nil)
		h.repo.EXPECT().ListActivations(h.ctx, manifest.ID).Return([]*model.PluginActivation{
			{PluginID: manifest.ID, ScopeID: h.orgID, Enabled: true},
		}, nil)

		got, err := h.svc.Upgrade(h.ctx, manifest.ID, zipBytes)
		require.NoError(t, err)
		assert.Equal(t, upgraded.Version, got.Version)
	})

	t.Run("id mismatch", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(true, nil)
		h.repo.EXPECT().GetInstallation(h.ctx, "com.elemo.other").Return(&model.PluginInstallation{
			PluginID: "com.elemo.other",
			Version:  "1.0.0",
			Manifest: model.PluginManifest{
				SchemaVersion: 1,
				ID:            "com.elemo.other",
				Name:          "Other",
				Version:       "1.0.0",
				Requires:      model.PluginRequires{PluginAPI: "^1"},
				Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
			},
		}, nil)

		_, err := h.svc.Upgrade(h.ctx, "com.elemo.other", zipBytes)
		require.ErrorIs(t, err, model.ErrInvalidPluginID)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(true, nil)
		h.repo.EXPECT().GetInstallation(h.ctx, manifest.ID).Return(nil, repository.ErrNotFound)

		_, err := h.svc.Upgrade(h.ctx, manifest.ID, zipBytes)
		require.ErrorIs(t, err, repository.ErrNotFound)
	})
}

func TestPluginService_Uninstall(t *testing.T) {
	t.Parallel()
	manifest := frontendPluginManifest()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(true, nil)
		h.repo.EXPECT().GetInstallation(h.ctx, manifest.ID).Return(&model.PluginInstallation{
			PluginID: manifest.ID,
			Version:  manifest.Version,
			Manifest: manifest,
		}, nil)
		h.extRepo.EXPECT().DeleteByPlugin(h.ctx, manifest.ID).Return(nil)
		h.repo.EXPECT().DeleteActivations(h.ctx, manifest.ID).Return(nil)
		h.repo.EXPECT().DeleteStorageForPlugin(h.ctx, manifest.ID).Return(nil)
		h.repo.EXPECT().DeleteInstallation(h.ctx, manifest.ID).Return(nil)

		require.NoError(t, h.svc.Uninstall(h.ctx, manifest.ID))
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(true, nil)
		h.repo.EXPECT().GetInstallation(h.ctx, manifest.ID).Return(nil, repository.ErrNotFound)

		err := h.svc.Uninstall(h.ctx, manifest.ID)
		require.ErrorIs(t, err, repository.ErrNotFound)
	})
}

func TestPluginService_Disable(t *testing.T) {
	t.Parallel()
	manifest := frontendPluginManifest()

	t.Run("stops when last activation", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, h.orgID, model.ActionPluginManage).Return(true, nil)
		h.repo.EXPECT().GetActivation(h.ctx, manifest.ID, h.orgID).Return(&model.PluginActivation{
			PluginID: manifest.ID,
			ScopeID:  h.orgID,
			Enabled:  true,
			Config:   json.RawMessage(`{}`),
		}, nil)
		h.repo.EXPECT().UpsertActivation(h.ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, act *model.PluginActivation) (*model.PluginActivation, error) {
				assert.False(t, act.Enabled)
				return act, nil
			},
		)
		h.repo.EXPECT().ListActivations(h.ctx, manifest.ID).Return([]*model.PluginActivation{
			{PluginID: manifest.ID, ScopeID: h.orgID, Enabled: false},
		}, nil)

		require.NoError(t, h.svc.Disable(h.ctx, manifest.ID, h.orgID))
	})
}

func TestPluginService_GetAndList(t *testing.T) {
	t.Parallel()
	manifest := frontendPluginManifest()
	inst := &model.PluginInstallation{
		PluginID: manifest.ID,
		Version:  manifest.Version,
		Status:   model.PluginStatusInstalled,
		Manifest: manifest,
	}

	t.Run("get", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(true, nil)
		h.repo.EXPECT().GetInstallation(h.ctx, manifest.ID).Return(inst, nil)

		got, err := h.svc.Get(h.ctx, manifest.ID)
		require.NoError(t, err)
		assert.Equal(t, manifest.ID, got.PluginID)
	})

	t.Run("list", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, model.InstallationID(), model.ActionPluginInstall).Return(true, nil)
		h.repo.EXPECT().ListInstallations(h.ctx).Return([]*model.PluginInstallation{inst}, nil)

		got, err := h.svc.List(h.ctx)
		require.NoError(t, err)
		require.Len(t, got, 1)
	})
}

func TestPluginService_Config(t *testing.T) {
	t.Parallel()
	manifest := frontendPluginManifest()
	cfg := json.RawMessage(`{"k":"v"}`)

	t.Run("get managed missing activation", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, h.orgID, model.ActionPluginManage).Return(true, nil)
		h.repo.EXPECT().GetActivation(h.ctx, manifest.ID, h.orgID).Return(nil, repository.ErrNotFound)

		got, err := h.svc.GetManagedConfig(h.ctx, manifest.ID, h.orgID)
		require.NoError(t, err)
		assert.JSONEq(t, `{}`, string(got))
	})

	t.Run("get managed", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, h.orgID, model.ActionPluginManage).Return(true, nil)
		h.repo.EXPECT().GetActivation(h.ctx, manifest.ID, h.orgID).Return(&model.PluginActivation{
			PluginID: manifest.ID,
			ScopeID:  h.orgID,
			Config:   cfg,
		}, nil)

		got, err := h.svc.GetManagedConfig(h.ctx, manifest.ID, h.orgID)
		require.NoError(t, err)
		assert.JSONEq(t, `{"k":"v"}`, string(got))
	})

	t.Run("get nearest", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().ListScopeAncestry(h.ctx, h.orgID).Return([]model.ID{h.orgID}, nil)
		h.repo.EXPECT().ListActivationsByScope(h.ctx, []model.ID{h.orgID}).Return([]*model.PluginActivation{
			{PluginID: manifest.ID, ScopeID: h.orgID, Enabled: true, Config: cfg},
		}, nil)

		got, err := h.svc.GetConfig(h.ctx, manifest.ID, h.orgID)
		require.NoError(t, err)
		assert.JSONEq(t, `{"k":"v"}`, string(got))
	})

	t.Run("set", func(t *testing.T) {
		t.Parallel()
		h := newPluginLifecycleHarness(t)
		h.lic.EXPECT().HasFeature(h.ctx, license.FeaturePlugins).Return(true, nil)
		h.perm.EXPECT().CtxUserHas(h.ctx, h.orgID, model.ActionPluginManage).Return(true, nil)
		h.repo.EXPECT().GetInstallation(h.ctx, manifest.ID).Return(&model.PluginInstallation{
			PluginID: manifest.ID,
			Manifest: manifest,
		}, nil)
		h.repo.EXPECT().GetActivation(h.ctx, manifest.ID, h.orgID).Return(&model.PluginActivation{
			PluginID: manifest.ID,
			ScopeID:  h.orgID,
			Enabled:  true,
		}, nil)
		h.repo.EXPECT().UpsertActivation(h.ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, act *model.PluginActivation) (*model.PluginActivation, error) {
				assert.JSONEq(t, `{"k":"v"}`, string(act.Config))
				return act, nil
			},
		)

		require.NoError(t, h.svc.SetConfig(h.ctx, manifest.ID, h.orgID, cfg))
	})
}
