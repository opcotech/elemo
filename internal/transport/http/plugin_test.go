package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func testPluginInstallation() *model.PluginInstallation {
	now := time.Now().UTC()
	required := true
	_ = required
	return &model.PluginInstallation{
		ID:       "inst-1",
		PluginID: "com.elemo.timetracking",
		Version:  "1.0.0",
		Status:   model.PluginStatusInstalled,
		Error:    "load failed",
		Manifest: model.PluginManifest{
			SchemaVersion: model.PluginSchemaVersionV1,
			ID:            "com.elemo.timetracking",
			Name:          "Time Tracking",
			Version:       "1.0.0",
			Requires:      model.PluginRequires{PluginAPI: "^1"},
			Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js", Module: "timetracking"},
			Capabilities:  []model.PluginCapability{model.CapabilityIssuesRead, model.CapabilityGraphRead},
			Slots:         []model.PluginUISlot{model.PluginSlotIssueSidebar},
			Config: []model.PluginConfigFieldDecl{
				{Name: "time_source", Type: model.PluginConfigFieldTypeGraphBinding, Foreign: "LoggedTime", Required: true},
			},
			Graph: &model.PluginGraphSchema{
				Nodes: []model.PluginGraphNodeDecl{
					{
						Kind:  "TimeEntry",
						Scope: model.PluginGraphNodeScope{Parent: "Issue"},
						Properties: []model.PluginGraphPropertyDecl{
							{Name: "seconds", Type: model.PluginGraphPropertyTypeInteger, Required: true},
						},
					},
				},
				Foreign: []model.PluginGraphForeignDecl{
					{Name: "LoggedTime", Parent: "Issue"},
				},
			},
		},
		CreatedAt: &now,
		UpdatedAt: &now,
	}
}

func testExtension(t *testing.T) *model.Extension {
	t.Helper()
	ext, err := model.NewExtension("com.elemo.timetracking", "TimeEntry", map[string]any{"seconds": int64(30)})
	require.NoError(t, err)
	parent := model.MustNewID(model.ResourceTypeIssue)
	ext.Parent = &parent
	return ext
}

func pluginPackageReader(t *testing.T, zipBytes []byte) *multipart.Reader {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("package", "plugin.zip")
	require.NoError(t, err)
	_, err = part.Write(zipBytes)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return multipart.NewReader(bytes.NewReader(buf.Bytes()), w.Boundary())
}

func TestNewPluginController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c, err := NewPluginController(mocksvc.NewMockPluginService(ctrl))
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("no service", func(t *testing.T) {
		t.Parallel()
		_, err := NewPluginController(nil)
		assert.ErrorIs(t, err, ErrNoPluginService)
	})
}

func TestPluginToDTO(t *testing.T) {
	t.Parallel()

	assert.Equal(t, api.Plugin{}, pluginToDTO(nil, nil, nil))

	inst := testPluginInstallation()
	enabled := true
	dto := pluginToDTO(inst, &enabled, json.RawMessage(`{"time_source":{"plugin_id":"com.elemo.timetracking","kind":"TimeEntry"}}`))
	assert.Equal(t, inst.ID, dto.Id)
	assert.Equal(t, inst.PluginID, dto.PluginId)
	assert.Equal(t, inst.Manifest.Name, dto.Name)
	assert.Equal(t, api.PluginStatus(inst.Status.String()), dto.Status)
	require.NotNil(t, dto.Enabled)
	assert.True(t, *dto.Enabled)
	require.NotNil(t, dto.Error)
	assert.Equal(t, "load failed", *dto.Error)
	require.NotNil(t, dto.Config)
	require.NotNil(t, dto.ConfigSchema)
	require.NotNil(t, dto.Graph)
	require.NotNil(t, dto.Graph.Nodes)
	require.NotNil(t, dto.Graph.Foreign)
}

func TestRawConfigMap(t *testing.T) {
	t.Parallel()
	assert.Empty(t, rawConfigMap(nil))
	assert.Empty(t, rawConfigMap(json.RawMessage("null")))
	assert.Empty(t, rawConfigMap(json.RawMessage("[]")))
	got := rawConfigMap(json.RawMessage(`{"k":"v"}`))
	assert.Equal(t, "v", got["k"])
}

func TestParseEqualsFilter(t *testing.T) {
	t.Parallel()
	got, err := parseEqualsFilter(nil)
	require.NoError(t, err)
	assert.Nil(t, got)

	empty := "  "
	got, err = parseEqualsFilter(&empty)
	require.NoError(t, err)
	assert.Nil(t, got)

	raw := `{"seconds":30}`
	got, err = parseEqualsFilter(&raw)
	require.NoError(t, err)
	assert.Equal(t, float64(30), got["seconds"])

	bad := "not-json"
	_, err = parseEqualsFilter(&bad)
	require.Error(t, err)
}

func TestExtensionAndRelationToDTO(t *testing.T) {
	t.Parallel()
	assert.Equal(t, api.ExtensionNode{}, extensionToDTO(nil))
	assert.Equal(t, api.ExtensionRelation{}, relationToDTO(nil))

	ext := testExtension(t)
	ext.Properties = nil
	dto := extensionToDTO(ext)
	assert.Equal(t, ext.ID.String(), dto.Id)
	assert.Equal(t, ext.PluginID, dto.PluginId)
	assert.NotNil(t, dto.Properties)
	require.NotNil(t, dto.ParentId)
	require.NotNil(t, dto.ParentType)

	now := time.Now().UTC()
	rel := &model.ExtensionRelation{
		ID:        "rel-1",
		Kind:      "LOGGED_ON",
		From:      ext.ID,
		To:        *ext.Parent,
		CreatedAt: &now,
	}
	relDTO := relationToDTO(rel)
	assert.Equal(t, "rel-1", relDTO.Id)
	assert.Equal(t, ext.ID.String(), relDTO.From)
	assert.Equal(t, api.ResourceTypeExtension, relDTO.FromType)
}

func TestMapOrEmpty(t *testing.T) {
	t.Parallel()
	assert.Empty(t, mapOrEmpty(nil))
	in := map[string]any{"a": 1}
	assert.Equal(t, in, mapOrEmpty(&in))
}

func TestReadPluginPackage(t *testing.T) {
	t.Parallel()

	_, err := readPluginPackage(nil)
	require.Error(t, err)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	other, err := w.CreateFormField("other")
	require.NoError(t, err)
	_, err = io.WriteString(other, "skip")
	require.NoError(t, err)
	part, err := w.CreateFormFile("package", "plugin.zip")
	require.NoError(t, err)
	_, err = part.Write([]byte("zip-bytes"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	got, err := readPluginPackage(multipart.NewReader(bytes.NewReader(buf.Bytes()), w.Boundary()))
	require.NoError(t, err)
	assert.Equal(t, []byte("zip-bytes"), got)

	emptyBuf := bytes.Buffer{}
	emptyW := multipart.NewWriter(&emptyBuf)
	require.NoError(t, emptyW.Close())
	_, err = readPluginPackage(multipart.NewReader(bytes.NewReader(emptyBuf.Bytes()), emptyW.Boundary()))
	require.Error(t, err)
}

func TestPluginAssetContentType(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "text/javascript; charset=utf-8", pluginAssetContentType("a.js"))
	assert.Equal(t, "text/javascript; charset=utf-8", pluginAssetContentType("a.mjs"))
	assert.Equal(t, "text/css; charset=utf-8", pluginAssetContentType("a.css"))
	assert.Equal(t, "application/wasm", pluginAssetContentType("a.wasm"))
	assert.Equal(t, "application/json; charset=utf-8", pluginAssetContentType("a.json"))
	assert.Equal(t, "image/svg+xml", pluginAssetContentType("a.svg"))
	assert.Equal(t, "image/png", pluginAssetContentType("a.png"))
	assert.Equal(t, "image/jpeg", pluginAssetContentType("a.jpg"))
	assert.Equal(t, "font/woff2", pluginAssetContentType("a.woff2"))
	assert.Equal(t, "application/json", pluginAssetContentType("a.map"))
	assert.Equal(t, "application/octet-stream", pluginAssetContentType("a.bin"))
}

func TestPluginController_V1PluginsGet(t *testing.T) {
	t.Parallel()
	inst := testPluginInstallation()
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	scopeID := orgID.String()
	scopeType := api.ResourceTypeOrganization

	t.Run("catalog", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().List(gomock.Any()).Return([]*model.PluginInstallation{inst}, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginsGet(context.Background(), api.V1PluginsGetRequestObject{})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginsGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got, 1)
		assert.Equal(t, inst.PluginID, got[0].PluginId)
	})

	t.Run("managed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		enabled := true
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().ListManaged(gomock.Any(), orgID).Return([]service.PluginListItem{{
			Installation: inst,
			Enabled:      &enabled,
			Config:       json.RawMessage(`{"k":1}`),
		}}, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginsGet(context.Background(), api.V1PluginsGetRequestObject{
			Params: api.V1PluginsGetParams{ScopeId: &scopeID, ScopeType: &scopeType},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginsGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got, 1)
		require.NotNil(t, got[0].Enabled)
	})

	t.Run("incomplete scope", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		resp, err := c.V1PluginsGet(context.Background(), api.V1PluginsGetRequestObject{
			Params: api.V1PluginsGetParams{ScopeId: &scopeID},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginsGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().List(gomock.Any()).Return(nil, service.ErrNoPermission)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginsGet(context.Background(), api.V1PluginsGetRequestObject{})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginsGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("managed forbidden", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().ListManaged(gomock.Any(), orgID).Return(nil, service.ErrNoPermission)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginsGet(context.Background(), api.V1PluginsGetRequestObject{
			Params: api.V1PluginsGetParams{ScopeId: &scopeID, ScopeType: &scopeType},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginsGet403JSONResponse)
		assert.True(t, ok)
	})
}

func TestPluginController_V1PluginsCreate(t *testing.T) {
	t.Parallel()
	inst := testPluginInstallation()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Install(gomock.Any(), []byte("zip")).Return(inst, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginsCreate(context.Background(), api.V1PluginsCreateRequestObject{
			Body: pluginPackageReader(t, []byte("zip")),
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, inst.PluginID, got.PluginId)
	})

	t.Run("missing package", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		resp, err := c.V1PluginsCreate(context.Background(), api.V1PluginsCreateRequestObject{})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Install(gomock.Any(), gomock.Any()).Return(nil, errors.Join(service.ErrPluginInstall, repository.ErrPluginConflict))
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginsCreate(context.Background(), api.V1PluginsCreateRequestObject{
			Body: pluginPackageReader(t, []byte("zip")),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginsCreate409JSONResponse)
		assert.True(t, ok)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Install(gomock.Any(), gomock.Any()).Return(nil, service.ErrNoPermission)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginsCreate(context.Background(), api.V1PluginsCreateRequestObject{
			Body: pluginPackageReader(t, []byte("zip")),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginsCreate403JSONResponse)
		assert.True(t, ok)
	})
}

func TestPluginController_V1PluginsFrontendGet(t *testing.T) {
	t.Parallel()
	orgID := model.MustNewID(model.ResourceTypeOrganization)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().ListFrontend(gomock.Any(), orgID).Return([]service.FrontendPlugin{{
			ID:         "com.elemo.timetracking",
			Version:    "1.0.0",
			Entrypoint: "frontend/index.js",
			Module:     "timetracking",
			Slots:      []model.PluginUISlot{model.PluginSlotIssueSidebar},
		}}, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginsFrontendGet(context.Background(), api.V1PluginsFrontendGetRequestObject{
			Params: api.V1PluginsFrontendGetParams{ScopeId: orgID.String(), ScopeType: api.ResourceTypeOrganization},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginsFrontendGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got, 1)
		require.NotNil(t, got[0].Module)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().ListFrontend(gomock.Any(), gomock.Any()).Return(nil, repository.ErrNotFound)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginsFrontendGet(context.Background(), api.V1PluginsFrontendGetRequestObject{
			Params: api.V1PluginsFrontendGetParams{ScopeId: orgID.String(), ScopeType: api.ResourceTypeOrganization},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginsFrontendGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestPluginController_V1PluginGetDelete(t *testing.T) {
	t.Parallel()
	inst := testPluginInstallation()

	t.Run("get success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Get(gomock.Any(), inst.PluginID).Return(inst, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginGet(context.Background(), api.V1PluginGetRequestObject{PluginId: inst.PluginID})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, inst.PluginID, got.PluginId)
	})

	t.Run("get not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Get(gomock.Any(), inst.PluginID).Return(nil, repository.ErrNotFound)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginGet(context.Background(), api.V1PluginGetRequestObject{PluginId: inst.PluginID})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGet404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("delete success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Uninstall(gomock.Any(), inst.PluginID).Return(nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginDelete(context.Background(), api.V1PluginDeleteRequestObject{PluginId: inst.PluginID})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginDelete204Response)
		assert.True(t, ok)
	})

	t.Run("delete forbidden", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Uninstall(gomock.Any(), inst.PluginID).Return(service.ErrNoPermission)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginDelete(context.Background(), api.V1PluginDeleteRequestObject{PluginId: inst.PluginID})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginDelete403JSONResponse)
		assert.True(t, ok)
	})
}

func TestPluginController_EnableDisableConfig(t *testing.T) {
	t.Parallel()
	pluginID := "com.elemo.timetracking"
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	cfg := map[string]any{"k": "v"}

	t.Run("enable success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Enable(gomock.Any(), pluginID, orgID, gomock.Any()).Return(nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginEnable(context.Background(), api.V1PluginEnableRequestObject{
			PluginId: pluginID,
			Body:     &api.V1PluginEnableJSONRequestBody{ScopeId: orgID.String(), ScopeType: api.ResourceTypeOrganization, Config: &cfg},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginEnable204Response)
		assert.True(t, ok)
	})

	t.Run("enable missing body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		resp, err := c.V1PluginEnable(context.Background(), api.V1PluginEnableRequestObject{PluginId: pluginID})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginEnable400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("disable success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Disable(gomock.Any(), pluginID, orgID).Return(nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginDisable(context.Background(), api.V1PluginDisableRequestObject{
			PluginId: pluginID,
			Body:     &api.V1PluginDisableJSONRequestBody{ScopeId: orgID.String(), ScopeType: api.ResourceTypeOrganization},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginDisable204Response)
		assert.True(t, ok)
	})

	t.Run("disable missing body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		resp, err := c.V1PluginDisable(context.Background(), api.V1PluginDisableRequestObject{PluginId: pluginID})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginDisable400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("config get", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().GetManagedConfig(gomock.Any(), pluginID, orgID).Return(json.RawMessage(`{"k":"v"}`), nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginConfigGet(context.Background(), api.V1PluginConfigGetRequestObject{
			PluginId: pluginID,
			Params:   api.V1PluginConfigGetParams{ScopeId: orgID.String(), ScopeType: api.ResourceTypeOrganization},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginConfigGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "v", got.Config["k"])
	})

	t.Run("config patch", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().SetConfig(gomock.Any(), pluginID, orgID, gomock.Any()).Return(nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginConfigPatch(context.Background(), api.V1PluginConfigPatchRequestObject{
			PluginId: pluginID,
			Params:   api.V1PluginConfigPatchParams{ScopeId: orgID.String(), ScopeType: api.ResourceTypeOrganization},
			Body:     &api.V1PluginConfigPatchJSONRequestBody{Config: cfg},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginConfigPatch200JSONResponse)
		assert.True(t, ok)
	})

	t.Run("config patch missing body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		resp, err := c.V1PluginConfigPatch(context.Background(), api.V1PluginConfigPatchRequestObject{
			PluginId: pluginID,
			Params:   api.V1PluginConfigPatchParams{ScopeId: orgID.String(), ScopeType: api.ResourceTypeOrganization},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginConfigPatch400JSONResponse)
		assert.True(t, ok)
	})
}

func TestPluginController_UpgradeInvoke(t *testing.T) {
	t.Parallel()
	inst := testPluginInstallation()

	t.Run("upgrade success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Upgrade(gomock.Any(), inst.PluginID, []byte("zip")).Return(inst, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginUpgrade(context.Background(), api.V1PluginUpgradeRequestObject{
			PluginId: inst.PluginID,
			Body:     pluginPackageReader(t, []byte("zip")),
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginUpgrade200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, inst.PluginID, got.PluginId)
	})

	t.Run("upgrade missing package", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		resp, err := c.V1PluginUpgrade(context.Background(), api.V1PluginUpgradeRequestObject{PluginId: inst.PluginID})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginUpgrade400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("invoke success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Invoke(gomock.Any(), inst.PluginID, gomock.Any()).Return(elemoplugin.InvokeResponse{
			OK:    true,
			Data:  json.RawMessage(`{"running":true}`),
			Error: "",
		}, nil)
		c := newTestPluginController(t, svc)
		payload := any(map[string]any{"issueId": "x"})
		resp, err := c.V1PluginInvoke(context.Background(), api.V1PluginInvokeRequestObject{
			PluginId: inst.PluginID,
			Body: &api.V1PluginInvokeJSONRequestBody{
				Function: "timer.start",
				ScopeId:  "Issue:abc",
				Payload:  payload,
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginInvoke200JSONResponse)
		require.True(t, ok)
		assert.True(t, got.Ok)
	})

	t.Run("invoke missing body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		resp, err := c.V1PluginInvoke(context.Background(), api.V1PluginInvokeRequestObject{PluginId: inst.PluginID})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginInvoke400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("invoke guest error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().Invoke(gomock.Any(), inst.PluginID, gomock.Any()).Return(elemoplugin.InvokeResponse{
			OK:    false,
			Error: "guest failed",
		}, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginInvoke(context.Background(), api.V1PluginInvokeRequestObject{
			PluginId: inst.PluginID,
			Body:     &api.V1PluginInvokeJSONRequestBody{Function: "timer.start", ScopeId: "Issue:abc"},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginInvoke200JSONResponse)
		require.True(t, ok)
		require.NotNil(t, got.Error)
		assert.Equal(t, "guest failed", *got.Error)
	})
}

func TestPluginController_Graph(t *testing.T) {
	t.Parallel()
	pluginID := "com.elemo.timetracking"
	ext := testExtension(t)
	parent := *ext.Parent
	rel := &model.ExtensionRelation{ID: "rel-1", Kind: "LOGGED_ON", From: ext.ID, To: parent}

	t.Run("list nodes", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().ListNodes(gomock.Any(), pluginID, gomock.Any()).Return(repository.Page[*model.Extension]{
			Items: []*model.Extension{ext},
		}, nil)
		c := newTestPluginController(t, svc)
		equals := `{"seconds":30}`
		owner := "com.elemo.other"
		resp, err := c.V1PluginGraphNodesGet(context.Background(), api.V1PluginGraphNodesGetRequestObject{
			PluginId: pluginID,
			Params: api.V1PluginGraphNodesGetParams{
				Kind:          "TimeEntry",
				ScopeId:       parent.String(),
				ScopeType:     api.ResourceTypeIssue,
				Equals:        &equals,
				OwnerPluginId: &owner,
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginGraphNodesGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
	})

	t.Run("list nodes bad equals", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		bad := "nope"
		resp, err := c.V1PluginGraphNodesGet(context.Background(), api.V1PluginGraphNodesGetRequestObject{
			PluginId: pluginID,
			Params: api.V1PluginGraphNodesGetParams{
				Kind:      "TimeEntry",
				ScopeId:   parent.String(),
				ScopeType: api.ResourceTypeIssue,
				Equals:    &bad,
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGraphNodesGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("create node", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().CreateNode(gomock.Any(), pluginID, gomock.Any()).Return(ext, nil)
		c := newTestPluginController(t, svc)
		props := map[string]any{"seconds": 30}
		body := api.V1PluginGraphNodesCreateJSONRequestBody{
			Kind:       "TimeEntry",
			ParentId:   parent.String(),
			ParentType: api.ResourceTypeIssue,
			Properties: &props,
		}
		require.NoError(t, json.Unmarshal([]byte(`{"kind":"LOGGED_ON","to_id":"`+parent.String()+`","to_type":"Issue"}`), &body.Relation))
		resp, err := c.V1PluginGraphNodesCreate(context.Background(), api.V1PluginGraphNodesCreateRequestObject{
			PluginId: pluginID,
			Body:     &body,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGraphNodesCreate201JSONResponse)
		assert.True(t, ok)
	})

	t.Run("create node missing body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		resp, err := c.V1PluginGraphNodesCreate(context.Background(), api.V1PluginGraphNodesCreateRequestObject{PluginId: pluginID})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGraphNodesCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("get node", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().GetNode(gomock.Any(), pluginID, ext.ID, "").Return(ext, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginGraphNodeGet(context.Background(), api.V1PluginGraphNodeGetRequestObject{
			PluginId: pluginID,
			Id:       api.Id(ext.ID.String()),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGraphNodeGet200JSONResponse)
		assert.True(t, ok)
	})

	t.Run("get node bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		resp, err := c.V1PluginGraphNodeGet(context.Background(), api.V1PluginGraphNodeGetRequestObject{
			PluginId: pluginID,
			Id:       "bad",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGraphNodeGet404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("update node", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().UpdateNode(gomock.Any(), pluginID, ext.ID, gomock.Any()).Return(ext, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginGraphNodeUpdate(context.Background(), api.V1PluginGraphNodeUpdateRequestObject{
			PluginId: pluginID,
			Id:       api.Id(ext.ID.String()),
			Body:     &api.V1PluginGraphNodeUpdateJSONRequestBody{Properties: map[string]any{"seconds": 90}},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGraphNodeUpdate200JSONResponse)
		assert.True(t, ok)
	})

	t.Run("delete node", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().DeleteNode(gomock.Any(), pluginID, ext.ID).Return(nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginGraphNodeDelete(context.Background(), api.V1PluginGraphNodeDeleteRequestObject{
			PluginId: pluginID,
			Id:       api.Id(ext.ID.String()),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGraphNodeDelete204Response)
		assert.True(t, ok)
	})

	t.Run("move node", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().MoveNode(gomock.Any(), pluginID, ext.ID, parent).Return(ext, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginGraphNodeMove(context.Background(), api.V1PluginGraphNodeMoveRequestObject{
			PluginId: pluginID,
			Id:       api.Id(ext.ID.String()),
			Body:     &api.V1PluginGraphNodeMoveJSONRequestBody{ParentId: parent.String(), ParentType: api.ResourceTypeIssue},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGraphNodeMove200JSONResponse)
		assert.True(t, ok)
	})

	t.Run("list relations", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().ListRelations(gomock.Any(), pluginID, gomock.Any()).Return(repository.Page[*model.ExtensionRelation]{
			Items: []*model.ExtensionRelation{rel},
		}, nil)
		c := newTestPluginController(t, svc)
		dir := api.PluginGraphRelationDirectionOutgoing
		resp, err := c.V1PluginGraphRelationsGet(context.Background(), api.V1PluginGraphRelationsGetRequestObject{
			PluginId: pluginID,
			Params: api.V1PluginGraphRelationsGetParams{
				Kind:      "LOGGED_ON",
				NodeId:    ext.ID.String(),
				NodeType:  api.ResourceTypeExtension,
				Direction: &dir,
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1PluginGraphRelationsGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
	})

	t.Run("create relation", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().CreateRelation(gomock.Any(), pluginID, gomock.Any()).Return(rel, nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginGraphRelationsCreate(context.Background(), api.V1PluginGraphRelationsCreateRequestObject{
			PluginId: pluginID,
			Body: &api.V1PluginGraphRelationsCreateJSONRequestBody{
				Kind:     "LOGGED_ON",
				FromId:   ext.ID.String(),
				FromType: api.ResourceTypeExtension,
				ToId:     parent.String(),
				ToType:   api.ResourceTypeIssue,
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGraphRelationsCreate201JSONResponse)
		assert.True(t, ok)
	})

	t.Run("delete relation", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().DeleteRelation(gomock.Any(), pluginID, "rel-1").Return(nil)
		c := newTestPluginController(t, svc)
		resp, err := c.V1PluginGraphRelationDelete(context.Background(), api.V1PluginGraphRelationDeleteRequestObject{
			PluginId: pluginID,
			Id:       "rel-1",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1PluginGraphRelationDelete204Response)
		assert.True(t, ok)
	})
}

func TestPluginController_ServePluginAsset(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		dir := t.TempDir()
		path := filepath.Join(dir, "index.js")
		require.NoError(t, os.WriteFile(path, []byte("export default {}"), 0o600))

		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().AssetPath(gomock.Any(), "com.elemo.timetracking", "1.0.0", "frontend/index.js").Return(path, nil)
		c := newTestPluginController(t, svc)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/plugins/com.elemo.timetracking/assets/1.0.0/frontend/index.js", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("pluginId", "com.elemo.timetracking")
		rctx.URLParams.Add("version", "1.0.0")
		rctx.URLParams.Add("*", "frontend/index.js")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rec := httptest.NewRecorder()
		c.ServePluginAsset(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "javascript")
	})

	t.Run("missing params", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := newTestPluginController(t, mocksvc.NewMockPluginService(ctrl))
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
		rec := httptest.NewRecorder()
		c.ServePluginAsset(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		svc := mocksvc.NewMockPluginService(ctrl)
		svc.EXPECT().AssetPath(gomock.Any(), "com.elemo.timetracking", "1.0.0", "missing.js").Return("", repository.ErrNotFound)
		c := newTestPluginController(t, svc)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("pluginId", "com.elemo.timetracking")
		rctx.URLParams.Add("version", "1.0.0")
		rctx.URLParams.Add("*", "missing.js")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rec := httptest.NewRecorder()
		c.ServePluginAsset(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
