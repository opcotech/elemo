package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/event"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
)

type stubRuntime struct {
	loadErr  error
	startErr error
	starts   int
}

func (r *stubRuntime) Load(context.Context, elemoplugin.Plugin) error {
	return r.loadErr
}

func (r *stubRuntime) Start(context.Context, string) error {
	r.starts++
	return r.startErr
}

func (*stubRuntime) Stop(context.Context, string) error { return nil }

func (*stubRuntime) Call(context.Context, string, string, []byte) ([]byte, error) {
	return nil, elemoplugin.ErrPluginNotStarted
}

func frontendPluginManifest() model.PluginManifest {
	return model.PluginManifest{
		SchemaVersion: model.PluginSchemaVersionV1,
		ID:            "com.elemo.timetracking",
		Name:          "Time Tracking",
		Version:       "1.0.0",
		Requires:      model.PluginRequires{PluginAPI: "^1"},
		Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
		Capabilities:  []model.PluginCapability{model.CapabilityIssuesRead},
		Slots:         []model.PluginUISlot{model.PluginSlotIssueSidebar},
	}
}

func timeSourceOwnerManifest() model.PluginManifest {
	m := frontendPluginManifest()
	m.Graph = &model.PluginGraphSchema{
		Nodes: []model.PluginGraphNodeDecl{
			{
				Kind:  "TimeEntry",
				Scope: model.PluginGraphNodeScope{Parent: "Issue"},
				Properties: []model.PluginGraphPropertyDecl{
					{Name: "seconds", Type: model.PluginGraphPropertyTypeInteger, Required: true},
				},
			},
		},
	}
	return m
}

func accountingEnableManifest() model.PluginManifest {
	return model.PluginManifest{
		SchemaVersion: model.PluginSchemaVersionV1,
		ID:            "com.elemo.accounting",
		Name:          "Accounting",
		Version:       "1.0.0",
		Requires:      model.PluginRequires{PluginAPI: "^1"},
		Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
		Capabilities:  []model.PluginCapability{model.CapabilityIssuesRead, model.CapabilityGraphRead},
		Config: []model.PluginConfigFieldDecl{
			{Name: "time_source", Type: model.PluginConfigFieldTypeGraphBinding, Foreign: "LoggedTime"},
		},
		Graph: &model.PluginGraphSchema{
			Foreign: []model.PluginGraphForeignDecl{
				{
					Name:   "LoggedTime",
					Parent: "Issue",
					Properties: []model.PluginGraphPropertyDecl{
						{Name: "seconds", Type: model.PluginGraphPropertyTypeInteger, Required: true},
					},
				},
			},
			Nodes: []model.PluginGraphNodeDecl{
				{
					Kind:  "Budget",
					Scope: model.PluginGraphNodeScope{Parent: "Organization"},
					Properties: []model.PluginGraphPropertyDecl{
						{Name: "seconds", Type: model.PluginGraphPropertyTypeInteger, Required: true},
					},
				},
			},
			Relations: []model.PluginGraphRelationDecl{
				{
					Kind:        "COUNTED_AGAINST",
					From:        "LoggedTime",
					To:          "Budget",
					Cardinality: model.PluginGraphCardinalityManyToOne,
				},
			},
		},
	}
}

type recordingRuntime struct {
	mu    sync.Mutex
	calls []string
}

func (*recordingRuntime) Load(context.Context, elemoplugin.Plugin) error { return nil }

func (*recordingRuntime) Start(context.Context, string) error { return nil }

func (*recordingRuntime) Stop(context.Context, string) error { return nil }

func (r *recordingRuntime) Call(_ context.Context, _, function string, _ []byte) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, function)
	return []byte(`{"ok":true}`), nil
}

func (r *recordingRuntime) callCount(function string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, call := range r.calls {
		if call == function {
			n++
		}
	}
	return n
}

func newPluginServiceHarness(t *testing.T) (
	context.Context,
	model.ID,
	*mocksvc.MockLicenseService,
	*mocksvc.MockPermissionService,
	*mockrepo.MockPluginRepository,
	service.PluginService,
) {
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

	svc, err := service.NewPluginService(
		config.PluginConfig{Directory: t.TempDir()},
		repo,
		mockrepo.NewMockExtensionRepository(ctrl),
		perm,
		lic,
		nil, nil, nil, nil,
		service.WithLogger(logger),
		service.WithTracer(tracer),
	)
	require.NoError(t, err)
	return ctx, orgID, lic, perm, repo, svc
}

func putRegistryPlugin(
	t *testing.T,
	svc service.PluginService,
	runtime elemoplugin.Runtime,
	manifest model.PluginManifest,
	root string,
	status model.PluginStatus,
) {
	t.Helper()
	if runtime == nil {
		runtime = elemoplugin.NoopRuntime{}
	}
	reg := elemoplugin.NewRegistry(runtime)
	require.NoError(t, reg.Put(context.Background(), elemoplugin.LoadedPlugin{
		ID:       manifest.ID,
		Version:  manifest.Version,
		Manifest: manifest,
		Root:     root,
		Status:   model.PluginStatusInstalled,
	}, nil))
	switch status {
	case model.PluginStatusActive:
		require.NoError(t, reg.Start(context.Background(), manifest.ID))
	case model.PluginStatusFailed:
		err := reg.Start(context.Background(), manifest.ID)
		require.Error(t, err)
	case model.PluginStatusDisabled:
		require.NoError(t, reg.Start(context.Background(), manifest.ID))
		require.NoError(t, reg.Stop(context.Background(), manifest.ID, 0))
	case model.PluginStatusInstalled:
	default:
		t.Fatalf("unsupported registry status %s", status)
	}
	service.SetPluginRegistry(svc, reg)
}

func writeFrontendAsset(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "frontend", "index.js")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte("export default {}"), 0o600))
	return root
}

func assertInsideRoot(t *testing.T, root, path string) {
	t.Helper()
	absRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)
	absPath, err := filepath.EvalSymlinks(path)
	require.NoError(t, err)
	rel, err := filepath.Rel(absRoot, absPath)
	require.NoError(t, err)
	assert.NotEqual(t, "..", rel)
	assert.False(t, strings.HasPrefix(rel, ".."+string(os.PathSeparator)))
}

func TestNewPluginService(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, err := service.NewPluginService(
		config.PluginConfig{},
		nil,
		mockrepo.NewMockExtensionRepository(ctrl),
		mocksvc.NewMockPermissionService(ctrl),
		mocksvc.NewMockLicenseService(ctrl),
		nil, nil, nil, nil,
	)
	assert.ErrorIs(t, err, service.ErrNoPluginRepository)
}

func TestPluginService_InvokeDisabled(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0)).AnyTimes()
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Len(0)).Return(ctx, span).AnyTimes()

	lic := mocksvc.NewMockLicenseService(ctrl)
	lic.EXPECT().HasFeature(ctx, license.FeaturePlugins).Return(true, nil)

	perm := mocksvc.NewMockPermissionService(ctrl)
	perm.EXPECT().ListScopeAncestry(ctx, orgID).Return([]model.ID{orgID}, nil)

	repo := mockrepo.NewMockPluginRepository(ctrl)
	repo.EXPECT().ListActivationsByScope(ctx, []model.ID{orgID}).Return([]*model.PluginActivation{}, nil)

	svc, err := service.NewPluginService(
		config.PluginConfig{Directory: t.TempDir()},
		repo,
		mockrepo.NewMockExtensionRepository(ctrl),
		perm,
		lic,
		nil, nil, nil, nil,
		service.WithLogger(mocklog.NewMockLogger(ctrl)),
		service.WithTracer(tracer),
	)
	require.NoError(t, err)

	_, err = svc.Invoke(ctx, "com.elemo.timetracking", elemoplugin.InvokeRequest{
		Function: "timer.start",
		ScopeID:  orgID.Composite(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, service.ErrPluginInvoke)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestPluginService_InvokeLogsGuestError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0)).AnyTimes()
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Len(0)).Return(ctx, span).AnyTimes()

	logger := mocklog.NewMockLogger(ctrl)
	logger.EXPECT().Warn(gomock.Any(), "plugin invoke failed", gomock.Any())

	lic := mocksvc.NewMockLicenseService(ctrl)
	lic.EXPECT().HasFeature(ctx, license.FeaturePlugins).Return(true, nil)

	perm := mocksvc.NewMockPermissionService(ctrl)
	perm.EXPECT().ListScopeAncestry(ctx, orgID).Return([]model.ID{orgID}, nil)

	manifest := frontendPluginManifest()
	require.NoError(t, manifest.Validate())

	repo := mockrepo.NewMockPluginRepository(ctrl)
	repo.EXPECT().ListActivationsByScope(ctx, []model.ID{orgID}).Return([]*model.PluginActivation{
		{PluginID: manifest.ID, ScopeID: orgID, Enabled: true},
	}, nil)

	svc, err := service.NewPluginService(
		config.PluginConfig{Directory: t.TempDir()},
		repo,
		mockrepo.NewMockExtensionRepository(ctrl),
		perm,
		lic,
		nil, nil, nil, nil,
		service.WithLogger(logger),
		service.WithTracer(tracer),
	)
	require.NoError(t, err)

	rt := &guestErrorRuntime{message: "no permission"}
	putRegistryPlugin(t, svc, rt, manifest, t.TempDir(), model.PluginStatusActive)

	resp, err := svc.Invoke(ctx, manifest.ID, elemoplugin.InvokeRequest{
		Function: "account.create",
		ScopeID:  orgID.Composite(),
	})
	require.NoError(t, err)
	assert.False(t, resp.OK)
	assert.Equal(t, "no permission", resp.Error)
}

func TestParseTypedID(t *testing.T) {
	t.Parallel()
	orgID := model.MustNewID(model.ResourceTypeOrganization)

	got, err := service.ParseTypedID(orgID.String(), model.ResourceTypeOrganization.String())
	require.NoError(t, err)
	assert.Equal(t, orgID, got)

	got, err = service.ParseTypedID(orgID.Composite(), model.ResourceTypeOrganization.String())
	require.NoError(t, err)
	assert.Equal(t, orgID, got)

	_, err = service.ParseTypedID(orgID.Composite(), model.ResourceTypeProject.String())
	require.Error(t, err)
}

type guestErrorRuntime struct {
	message string
}

func (*guestErrorRuntime) Load(context.Context, elemoplugin.Plugin) error { return nil }

func (*guestErrorRuntime) Start(context.Context, string) error { return nil }

func (*guestErrorRuntime) Stop(context.Context, string) error { return nil }

func (r *guestErrorRuntime) Call(context.Context, string, string, []byte) ([]byte, error) {
	return []byte(`{"ok":false,"error":"` + r.message + `"}`), nil
}

func TestPluginService_CreateNodeRequiresCapability(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := model.MustNewID(model.ResourceTypeUser)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0)).AnyTimes()
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Len(0)).Return(ctx, span).AnyTimes()

	lic := mocksvc.NewMockLicenseService(ctrl)
	lic.EXPECT().HasFeature(ctx, license.FeaturePlugins).Return(true, nil)

	manifest := model.PluginManifest{
		SchemaVersion: 1,
		ID:            "com.elemo.timetracking",
		Name:          "Time Tracking",
		Version:       "1.0.0",
		Requires:      model.PluginRequires{PluginAPI: "^1"},
		Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
		Capabilities:  []model.PluginCapability{model.CapabilityIssuesRead},
	}
	require.NoError(t, manifest.Validate())

	repo := mockrepo.NewMockPluginRepository(ctrl)
	repo.EXPECT().GetInstallation(ctx, manifest.ID).Return(&model.PluginInstallation{
		PluginID: manifest.ID,
		Version:  manifest.Version,
		Status:   model.PluginStatusActive,
		Manifest: manifest,
	}, nil)

	svc, err := service.NewPluginService(
		config.PluginConfig{Directory: t.TempDir()},
		repo,
		mockrepo.NewMockExtensionRepository(ctrl),
		mocksvc.NewMockPermissionService(ctrl),
		lic,
		nil, nil, nil, nil,
		service.WithLogger(mocklog.NewMockLogger(ctrl)),
		service.WithTracer(tracer),
	)
	require.NoError(t, err)

	_, err = svc.CreateNode(ctx, manifest.ID, service.CreateExtensionNodeOpts{
		Kind:   "TimeEntry",
		Parent: issueID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, service.ErrNoPermission)
}

func TestPluginService_ListFrontend(t *testing.T) {
	t.Parallel()

	manifest := frontendPluginManifest()
	require.NoError(t, manifest.Validate())

	tests := []struct {
		name      string
		status    model.PluginStatus
		inReg     bool
		frontend  bool
		wantCount int
	}{
		{name: "installed", status: model.PluginStatusInstalled, inReg: true, frontend: true, wantCount: 1},
		{name: "active", status: model.PluginStatusActive, inReg: true, frontend: true, wantCount: 1},
		{name: "failed", status: model.PluginStatusFailed, inReg: true, frontend: true, wantCount: 1},
		{name: "disabled", status: model.PluginStatusDisabled, inReg: true, frontend: true, wantCount: 0},
		{name: "missing from registry", inReg: false, frontend: true, wantCount: 0},
		{name: "no frontend", status: model.PluginStatusActive, inReg: true, frontend: false, wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, orgID, lic, perm, repo, svc := newPluginServiceHarness(t)
			m := manifest
			if !tt.frontend {
				m.Frontend = nil
				m.Backend = &model.PluginBackendDecl{Entry: model.PluginBackendWASMPath}
			}

			lic.EXPECT().HasFeature(ctx, license.FeaturePlugins).Return(true, nil)
			perm.EXPECT().CtxUserHas(ctx, orgID, model.ActionOrganizationRead).Return(true, nil)
			perm.EXPECT().ListScopeAncestry(ctx, orgID).Return([]model.ID{orgID}, nil)
			repo.EXPECT().ListActivationsByScope(ctx, []model.ID{orgID}).Return([]*model.PluginActivation{
				{PluginID: m.ID, ScopeID: orgID, Enabled: true},
			}, nil)
			repo.EXPECT().GetInstallation(ctx, m.ID).Return(&model.PluginInstallation{
				PluginID: m.ID,
				Version:  m.Version,
				Status:   tt.status,
				Manifest: m,
			}, nil)

			if tt.inReg {
				rt := elemoplugin.Runtime(elemoplugin.NoopRuntime{})
				if tt.status == model.PluginStatusFailed {
					rt = &stubRuntime{startErr: errors.New("start failed")}
				}
				putRegistryPlugin(t, svc, rt, m, t.TempDir(), tt.status)
			}

			got, err := svc.ListFrontend(ctx, orgID)
			require.NoError(t, err)
			require.Len(t, got, tt.wantCount)
			if tt.wantCount == 1 {
				assert.Equal(t, m.ID, got[0].ID)
				assert.Equal(t, "frontend/index.js", got[0].Entrypoint)
			}
		})
	}
}

func TestPluginService_ListManaged(t *testing.T) {
	t.Parallel()

	ctx, orgID, lic, perm, repo, svc := newPluginServiceHarness(t)
	inst := &model.PluginInstallation{
		PluginID: "com.elemo.timetracking",
		Version:  "1.1.0",
		Status:   model.PluginStatusInstalled,
		Manifest: frontendPluginManifest(),
	}

	lic.EXPECT().HasFeature(ctx, license.FeaturePlugins).Return(true, nil)
	perm.EXPECT().CtxUserHas(ctx, orgID, model.ActionPluginManage).Return(true, nil)
	repo.EXPECT().ListInstallations(ctx).Return([]*model.PluginInstallation{inst}, nil)
	repo.EXPECT().GetActivation(ctx, inst.PluginID, orgID).Return(nil, repository.ErrNotFound)

	got, err := svc.ListManaged(ctx, orgID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, inst.PluginID, got[0].Installation.PluginID)
	assert.Nil(t, got[0].Enabled)
}

func TestPluginService_AssetPath(t *testing.T) {
	t.Parallel()

	manifest := frontendPluginManifest()
	require.NoError(t, manifest.Validate())

	tests := []struct {
		name    string
		status  model.PluginStatus
		version string
		rel     string
		wantErr error
	}{
		{name: "installed", status: model.PluginStatusInstalled, version: manifest.Version, rel: "frontend/index.js"},
		{name: "active", status: model.PluginStatusActive, version: manifest.Version, rel: "frontend/index.js"},
		{name: "failed", status: model.PluginStatusFailed, version: manifest.Version, rel: "frontend/index.js"},
		{name: "disabled", status: model.PluginStatusDisabled, version: manifest.Version, rel: "frontend/index.js", wantErr: repository.ErrNotFound},
		{name: "version mismatch", status: model.PluginStatusActive, version: "9.9.9", rel: "frontend/index.js", wantErr: repository.ErrNotFound},
		{name: "missing file", status: model.PluginStatusInstalled, version: manifest.Version, rel: "frontend/missing.js", wantErr: repository.ErrNotFound},
		{name: "directory", status: model.PluginStatusInstalled, version: manifest.Version, rel: "frontend", wantErr: repository.ErrNotFound},
		{name: "path traversal", status: model.PluginStatusInstalled, version: manifest.Version, rel: "../secret", wantErr: repository.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _, _, _, _, svc := newPluginServiceHarness(t)
			root := writeFrontendAsset(t)
			rt := elemoplugin.Runtime(elemoplugin.NoopRuntime{})
			if tt.status == model.PluginStatusFailed {
				rt = &stubRuntime{startErr: errors.New("start failed")}
			}
			putRegistryPlugin(t, svc, rt, manifest, root, tt.status)

			path, err := svc.AssetPath(ctx, manifest.ID, tt.version, tt.rel)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, service.ErrPluginAsset)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.FileExists(t, path)
			assertInsideRoot(t, root, path)
		})
	}
}

func TestPluginService_OpenAsset(t *testing.T) {
	t.Parallel()

	manifest := frontendPluginManifest()
	require.NoError(t, manifest.Validate())

	t.Run("reads confined file", func(t *testing.T) {
		t.Parallel()
		ctx, _, _, _, _, svc := newPluginServiceHarness(t)
		root := writeFrontendAsset(t)
		putRegistryPlugin(t, svc, nil, manifest, root, model.PluginStatusInstalled)

		f, err := svc.OpenAsset(ctx, manifest.ID, manifest.Version, "frontend/index.js")
		require.NoError(t, err)
		defer f.Close()

		data, err := io.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, "export default {}", string(data))

		assertInsideRoot(t, root, f.Name())
	})

	t.Run("rejects traversal", func(t *testing.T) {
		t.Parallel()
		ctx, _, _, _, _, svc := newPluginServiceHarness(t)
		root := writeFrontendAsset(t)
		putRegistryPlugin(t, svc, nil, manifest, root, model.PluginStatusInstalled)

		f, err := svc.OpenAsset(ctx, manifest.ID, manifest.Version, "../secret")
		require.Error(t, err)
		assert.Nil(t, f)
		assert.ErrorIs(t, err, service.ErrPluginAsset)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})
}

func TestPluginService_Enable(t *testing.T) {
	t.Parallel()

	t.Run("starts then persists", func(t *testing.T) {
		t.Parallel()
		ctx, orgID, lic, perm, repo, svc := newPluginServiceHarness(t)
		manifest := frontendPluginManifest()
		require.NoError(t, manifest.Validate())
		rt := &stubRuntime{}
		putRegistryPlugin(t, svc, rt, manifest, t.TempDir(), model.PluginStatusInstalled)

		lic.EXPECT().HasFeature(ctx, license.FeaturePlugins).Return(true, nil)
		perm.EXPECT().CtxUserHas(ctx, orgID, model.ActionPluginManage).Return(true, nil)
		repo.EXPECT().GetInstallation(ctx, manifest.ID).Return(&model.PluginInstallation{
			PluginID: manifest.ID,
			Version:  manifest.Version,
			Status:   model.PluginStatusInstalled,
			Manifest: manifest,
		}, nil)
		repo.EXPECT().GetActivation(ctx, manifest.ID, orgID).Return(nil, repository.ErrNotFound)
		repo.EXPECT().UpsertActivation(ctx, &model.PluginActivation{
			PluginID: manifest.ID, ScopeID: orgID, Enabled: true,
		}).Return(&model.PluginActivation{
			PluginID: manifest.ID, ScopeID: orgID, Enabled: true,
		}, nil)

		require.NoError(t, svc.Enable(ctx, manifest.ID, orgID, nil))
		assert.Equal(t, 1, rt.starts)
	})

	t.Run("does not persist when start fails", func(t *testing.T) {
		t.Parallel()
		ctx, orgID, lic, perm, repo, svc := newPluginServiceHarness(t)
		manifest := frontendPluginManifest()
		require.NoError(t, manifest.Validate())
		rt := &stubRuntime{startErr: errors.New("start failed")}
		putRegistryPlugin(t, svc, rt, manifest, t.TempDir(), model.PluginStatusInstalled)

		lic.EXPECT().HasFeature(ctx, license.FeaturePlugins).Return(true, nil)
		perm.EXPECT().CtxUserHas(ctx, orgID, model.ActionPluginManage).Return(true, nil)
		repo.EXPECT().GetInstallation(ctx, manifest.ID).Return(&model.PluginInstallation{
			PluginID: manifest.ID,
			Version:  manifest.Version,
			Status:   model.PluginStatusInstalled,
			Manifest: manifest,
		}, nil)
		repo.EXPECT().GetActivation(ctx, manifest.ID, orgID).Return(nil, repository.ErrNotFound)
		repo.EXPECT().UpsertActivation(gomock.Any(), gomock.Any()).Times(0)

		err := svc.Enable(ctx, manifest.ID, orgID, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, service.ErrPluginEnable)
		assert.Equal(t, 1, rt.starts)
	})

	t.Run("rejects incompatible graph binding", func(t *testing.T) {
		t.Parallel()
		ctx, orgID, lic, perm, repo, svc := newPluginServiceHarness(t)
		manifest := accountingEnableManifest()
		require.NoError(t, manifest.Validate())
		owner := timeSourceOwnerManifest()
		require.NoError(t, owner.Validate())
		rt := &stubRuntime{}
		putRegistryPlugin(t, svc, rt, manifest, t.TempDir(), model.PluginStatusInstalled)

		lic.EXPECT().HasFeature(ctx, license.FeaturePlugins).Return(true, nil)
		perm.EXPECT().CtxUserHas(ctx, orgID, model.ActionPluginManage).Return(true, nil)
		repo.EXPECT().GetInstallation(ctx, manifest.ID).Return(&model.PluginInstallation{
			PluginID: manifest.ID, Version: manifest.Version, Status: model.PluginStatusInstalled, Manifest: manifest,
		}, nil)
		repo.EXPECT().GetActivation(ctx, manifest.ID, orgID).Return(nil, repository.ErrNotFound)
		repo.EXPECT().GetInstallation(ctx, owner.ID).Return(&model.PluginInstallation{
			PluginID: owner.ID, Version: owner.Version, Status: model.PluginStatusActive, Manifest: owner,
		}, nil)
		repo.EXPECT().UpsertActivation(gomock.Any(), gomock.Any()).Times(0)

		err := svc.Enable(ctx, manifest.ID, orgID, json.RawMessage(
			`{"time_source":{"plugin_id":"com.elemo.timetracking","kind":"Nope"}}`,
		))
		require.ErrorIs(t, err, model.ErrPluginGraphBinding)
		assert.Equal(t, 0, rt.starts)
	})

	t.Run("accepts compatible graph binding", func(t *testing.T) {
		t.Parallel()
		ctx, orgID, lic, perm, repo, svc := newPluginServiceHarness(t)
		manifest := accountingEnableManifest()
		require.NoError(t, manifest.Validate())
		owner := timeSourceOwnerManifest()
		require.NoError(t, owner.Validate())
		rt := &stubRuntime{}
		putRegistryPlugin(t, svc, rt, manifest, t.TempDir(), model.PluginStatusInstalled)
		cfg := json.RawMessage(`{"time_source":{"plugin_id":"com.elemo.timetracking","kind":"TimeEntry"}}`)

		lic.EXPECT().HasFeature(ctx, license.FeaturePlugins).Return(true, nil)
		perm.EXPECT().CtxUserHas(ctx, orgID, model.ActionPluginManage).Return(true, nil)
		repo.EXPECT().GetInstallation(ctx, manifest.ID).Return(&model.PluginInstallation{
			PluginID: manifest.ID, Version: manifest.Version, Status: model.PluginStatusInstalled, Manifest: manifest,
		}, nil)
		repo.EXPECT().GetActivation(ctx, manifest.ID, orgID).Return(nil, repository.ErrNotFound)
		repo.EXPECT().GetInstallation(ctx, owner.ID).Return(&model.PluginInstallation{
			PluginID: owner.ID, Version: owner.Version, Status: model.PluginStatusActive, Manifest: owner,
		}, nil)
		perm.EXPECT().ListScopeAncestry(ctx, orgID).Return([]model.ID{orgID}, nil)
		repo.EXPECT().ListActivationsByScope(ctx, []model.ID{orgID}).Return([]*model.PluginActivation{{
			PluginID: owner.ID, ScopeID: orgID, Enabled: true,
		}}, nil)
		repo.EXPECT().UpsertActivation(ctx, gomock.AssignableToTypeOf(&model.PluginActivation{})).DoAndReturn(
			func(_ context.Context, act *model.PluginActivation) (*model.PluginActivation, error) {
				assert.True(t, act.Enabled)
				assert.JSONEq(t, string(cfg), string(act.Config))
				return act, nil
			},
		)

		require.NoError(t, svc.Enable(ctx, manifest.ID, orgID, cfg))
		assert.Equal(t, 1, rt.starts)
	})
}

func TestPluginService_Restore(t *testing.T) {
	t.Parallel()

	t.Run("puts even when wasm is missing then starts if enabled", func(t *testing.T) {
		t.Parallel()
		ctx, orgID, _, _, repo, svc := newPluginServiceHarness(t)
		manifest := frontendPluginManifest()
		manifest.Backend = &model.PluginBackendDecl{Entry: model.PluginBackendWASMPath}
		require.NoError(t, manifest.Validate())

		rt := &stubRuntime{}
		reg := elemoplugin.NewRegistry(rt)
		service.SetPluginRegistry(svc, reg)

		inst := &model.PluginInstallation{
			PluginID: manifest.ID,
			Version:  manifest.Version,
			Status:   model.PluginStatusInstalled,
			Manifest: manifest,
		}
		repo.EXPECT().ListInstallations(ctx).Return([]*model.PluginInstallation{inst}, nil)
		repo.EXPECT().ListActivations(ctx, manifest.ID).Return([]*model.PluginActivation{
			{PluginID: manifest.ID, ScopeID: orgID, Enabled: true},
		}, nil)

		require.NoError(t, svc.Restore(ctx))
		_, ok := reg.Get(manifest.ID)
		assert.True(t, ok)
		assert.Equal(t, 1, rt.starts)
	})

	t.Run("puts even when load fails then starts if enabled", func(t *testing.T) {
		t.Parallel()
		ctx, orgID, _, _, repo, svc := newPluginServiceHarness(t)
		manifest := frontendPluginManifest()
		require.NoError(t, manifest.Validate())

		rt := &stubRuntime{loadErr: errors.New("load failed")}
		reg := elemoplugin.NewRegistry(rt)
		service.SetPluginRegistry(svc, reg)

		inst := &model.PluginInstallation{
			PluginID: manifest.ID,
			Version:  manifest.Version,
			Status:   model.PluginStatusInstalled,
			Manifest: manifest,
		}
		repo.EXPECT().ListInstallations(ctx).Return([]*model.PluginInstallation{inst}, nil)
		repo.EXPECT().ListActivations(ctx, manifest.ID).Return([]*model.PluginActivation{
			{PluginID: manifest.ID, ScopeID: orgID, Enabled: true},
		}, nil)

		require.NoError(t, svc.Restore(ctx))
		_, ok := reg.Get(manifest.ID)
		assert.True(t, ok)
		assert.Equal(t, 1, rt.starts)
	})
}

func TestPluginService_DispatchExtensionEvent(t *testing.T) {
	t.Parallel()

	t.Run("delivers to plugins without foreign aliases", func(t *testing.T) {
		t.Parallel()
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
		bus := event.NewBus()
		rt := &recordingRuntime{}

		svc, err := service.NewPluginService(
			config.PluginConfig{Directory: t.TempDir()},
			repo,
			mockrepo.NewMockExtensionRepository(ctrl),
			perm,
			lic,
			nil, nil, nil, bus,
			service.WithLogger(logger),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)

		manifest := frontendPluginManifest()
		manifest.Events = []model.PluginEventType{model.PluginEventExtensionCreated}
		require.NoError(t, manifest.Validate())
		putRegistryPlugin(t, svc, rt, manifest, t.TempDir(), model.PluginStatusActive)

		extID := model.MustNewID(model.ResourceTypeExtension)
		require.NoError(t, bus.Publish(ctx, event.Event{
			Type:     model.PluginEventExtensionCreated,
			Resource: extID,
			Payload: map[string]any{
				"plugin_id": "com.elemo.timetracking",
				"kind":      "TimeEntry",
				"id":        extID.String(),
				"scope_id":  orgID.Composite(),
			},
		}))
		service.WaitPluginEvents(svc)
		assert.Equal(t, 1, rt.callCount("onEvent"))
	})

	t.Run("filters foreign events by resolved binding", func(t *testing.T) {
		t.Parallel()
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
		bus := event.NewBus()
		rt := &recordingRuntime{}

		svc, err := service.NewPluginService(
			config.PluginConfig{Directory: t.TempDir()},
			repo,
			mockrepo.NewMockExtensionRepository(ctrl),
			perm,
			lic,
			nil, nil, nil, bus,
			service.WithLogger(logger),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)

		manifest := accountingEnableManifest()
		manifest.Events = []model.PluginEventType{model.PluginEventExtensionCreated}
		require.NoError(t, manifest.Validate())
		putRegistryPlugin(t, svc, rt, manifest, t.TempDir(), model.PluginStatusActive)

		extID := model.MustNewID(model.ResourceTypeExtension)
		perm.EXPECT().ListScopeAncestry(gomock.Any(), extID).Return([]model.ID{extID, orgID}, nil).AnyTimes()
		repo.EXPECT().ListActivationsByScope(gomock.Any(), gomock.Any()).Return([]*model.PluginActivation{{
			PluginID: manifest.ID,
			ScopeID:  orgID,
			Enabled:  true,
			Config:   json.RawMessage(`{"time_source":{"plugin_id":"com.elemo.timetracking","kind":"TimeEntry"}}`),
		}}, nil).AnyTimes()

		require.NoError(t, bus.Publish(ctx, event.Event{
			Type:     model.PluginEventExtensionCreated,
			Resource: extID,
			Payload: map[string]any{
				"plugin_id": "com.elemo.timetracking",
				"kind":      "TimeEntry",
				"id":        extID.String(),
				"scope_id":  orgID.Composite(),
			},
		}))
		service.WaitPluginEvents(svc)
		assert.Equal(t, 1, rt.callCount("onEvent"))

		require.NoError(t, bus.Publish(ctx, event.Event{
			Type:     model.PluginEventExtensionCreated,
			Resource: extID,
			Payload: map[string]any{
				"plugin_id": "com.elemo.other",
				"kind":      "TimeEntry",
				"id":        extID.String(),
				"scope_id":  orgID.Composite(),
			},
		}))
		service.WaitPluginEvents(svc)
		assert.Equal(t, 1, rt.callCount("onEvent"))
	})
}

func TestPluginService_InvokeDoesNotDeadlockOnExtensionEvent(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0)).AnyTimes()
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Len(0)).Return(ctx, span).AnyTimes()

	logger := mocklog.NewMockLogger(ctrl)
	logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	lic := mocksvc.NewMockLicenseService(ctrl)
	lic.EXPECT().HasFeature(ctx, license.FeaturePlugins).Return(true, nil)

	perm := mocksvc.NewMockPermissionService(ctrl)
	perm.EXPECT().ListScopeAncestry(ctx, orgID).Return([]model.ID{orgID}, nil)

	manifest := frontendPluginManifest()
	manifest.Events = []model.PluginEventType{model.PluginEventExtensionCreated}
	require.NoError(t, manifest.Validate())

	repo := mockrepo.NewMockPluginRepository(ctrl)
	repo.EXPECT().ListActivationsByScope(ctx, []model.ID{orgID}).Return([]*model.PluginActivation{
		{PluginID: manifest.ID, ScopeID: orgID, Enabled: true},
	}, nil)

	bus := event.NewBus()
	rt := &serialEventRuntime{bus: bus, pluginID: manifest.ID, orgID: orgID}

	svc, err := service.NewPluginService(
		config.PluginConfig{Directory: t.TempDir()},
		repo,
		mockrepo.NewMockExtensionRepository(ctrl),
		perm,
		lic,
		nil, nil, nil, bus,
		service.WithLogger(logger),
		service.WithTracer(tracer),
	)
	require.NoError(t, err)
	putRegistryPlugin(t, svc, rt, manifest, t.TempDir(), model.PluginStatusActive)

	done := make(chan error, 1)
	go func() {
		_, invokeErr := svc.Invoke(ctx, manifest.ID, elemoplugin.InvokeRequest{
			Function: "account.create",
			ScopeID:  orgID.Composite(),
		})
		done <- invokeErr
	}()

	select {
	case invokeErr := <-done:
		require.NoError(t, invokeErr)
	case <-time.After(2 * time.Second):
		t.Fatal("plugin invoke deadlocked on synchronous extension event dispatch")
	}

	service.WaitPluginEvents(svc)
	assert.GreaterOrEqual(t, rt.callCount("onEvent"), 1)
}

type serialEventRuntime struct {
	mu       sync.Mutex
	calls    []string
	bus      *event.Bus
	pluginID string
	orgID    model.ID
}

func (*serialEventRuntime) Load(context.Context, elemoplugin.Plugin) error { return nil }

func (*serialEventRuntime) Start(context.Context, string) error { return nil }

func (*serialEventRuntime) Stop(context.Context, string) error { return nil }

func (r *serialEventRuntime) Call(ctx context.Context, pluginID, function string, _ []byte) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, function)
	if function != "onEvent" && r.bus != nil {
		extID := model.MustNewID(model.ResourceTypeExtension)
		_ = r.bus.Publish(ctx, event.Event{
			Type:     model.PluginEventExtensionCreated,
			Resource: extID,
			Payload: map[string]any{
				"plugin_id": pluginID,
				"kind":      "Account",
				"id":        extID.String(),
				"scope_id":  r.orgID.Composite(),
			},
		})
	}
	return []byte(`{"ok":true}`), nil
}

func (r *serialEventRuntime) callCount(function string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, call := range r.calls {
		if call == function {
			n++
		}
	}
	return n
}
