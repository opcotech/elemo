package plugin_test

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/plugin"
)

func TestParseManifest(t *testing.T) {
	t.Parallel()

	raw := []byte(`
schemaVersion: 1
id: com.elemo.timetracking
name: Time Tracking
version: 1.0.0
requires:
  pluginApi: "^1"
backend:
  entry: backend/plugin.wasm
capabilities:
  - issues.read
  - graph.read
  - graph.write
graph:
  nodes:
    - kind: TimeEntry
      scope:
        parent: Issue
      properties:
        - name: seconds
          type: integer
          required: true
        - name: note
          type: string
  relations:
    - kind: LOGGED_ON
      from: TimeEntry
      to: Issue
      cardinality: many-to-one
`)
	manifest, err := plugin.ParseManifest(raw)
	require.NoError(t, err)
	assert.Equal(t, "com.elemo.timetracking", manifest.ID)
	require.NotNil(t, manifest.Graph)
	_, ok := manifest.Graph.NodeKind("TimeEntry")
	assert.True(t, ok)
}

func TestParseManifest_AccountingReference(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "..", "plugins", "accounting", "plugin.yaml"))
	require.NoError(t, err)
	manifest, err := plugin.ParseManifest(raw)
	require.NoError(t, err)
	assert.Equal(t, "com.elemo.accounting", manifest.ID)
	require.NotNil(t, manifest.Graph)
	_, ok := manifest.Graph.ForeignKind("LoggedTime")
	assert.True(t, ok)
	require.NotEmpty(t, manifest.Config)
	assert.Equal(t, model.PluginConfigFieldTypeGraphBinding, manifest.Config[0].Type)
	assert.Equal(t, "LoggedTime", manifest.Config[0].Foreign)
}

func TestParseManifest_UnknownSchema(t *testing.T) {
	t.Parallel()
	_, err := plugin.ParseManifest([]byte("schemaVersion: 9\nid: com.elemo.x\nname: X\nversion: 1\nrequires: {pluginApi: \"^1\"}\nfrontend: {entry: frontend/index.js}\n"))
	require.ErrorIs(t, err, model.ErrPluginSchemaVersion)
}

func TestExtractZip_RejectsTraversal(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("../evil.yaml")
	require.NoError(t, err)
	_, err = w.Write([]byte("nope"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	dir := t.TempDir()
	_, err = plugin.ExtractZip(buf.Bytes(), dir)
	require.Error(t, err)
}

func TestExtractZip_Valid(t *testing.T) {
	t.Parallel()

	manifest := model.PluginManifest{
		SchemaVersion: 1,
		ID:            "com.elemo.example",
		Name:          "Example",
		Version:       "0.1.0",
		Requires:      model.PluginRequires{PluginAPI: "^1"},
		Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
	}
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

	dir := t.TempDir()
	pkg, err := plugin.ExtractZip(buf.Bytes(), dir)
	require.NoError(t, err)
	assert.Equal(t, "com.elemo.example", pkg.Manifest.ID)
	_, err = os.Stat(filepath.Join(pkg.Root, "frontend", "index.js"))
	require.NoError(t, err)
}

func TestRegistry_Lifecycle(t *testing.T) {
	t.Parallel()

	rt := plugin.NewHandlerRuntime()
	reg := plugin.NewRegistry(rt)
	manifest := model.PluginManifest{
		SchemaVersion: 1,
		ID:            "com.elemo.example",
		Name:          "Example",
		Version:       "1.0.0",
		Requires:      model.PluginRequires{PluginAPI: "^1"},
		Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
	}
	require.NoError(t, manifest.Validate())

	err := reg.Put(context.Background(), plugin.LoadedPlugin{
		ID:       manifest.ID,
		Version:  manifest.Version,
		Manifest: manifest,
		Status:   model.PluginStatusInstalled,
	}, nil)
	require.NoError(t, err)

	require.NoError(t, reg.Start(context.Background(), manifest.ID))
	got, ok := reg.Get(manifest.ID)
	require.True(t, ok)
	assert.Equal(t, model.PluginStatusActive, got.Status)

	out, err := reg.Call(context.Background(), manifest.ID, "ping", []byte(`{}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(out))

	require.NoError(t, reg.Stop(context.Background(), manifest.ID, 0))
	_, err = reg.Call(context.Background(), manifest.ID, "ping", []byte(`{}`))
	require.ErrorIs(t, err, plugin.ErrInvokeDisabled)
}

func TestRegistry_CallDoesNotHoldLock(t *testing.T) {
	t.Parallel()

	rt := plugin.NewHandlerRuntime()
	reg := plugin.NewRegistry(rt)
	manifest := model.PluginManifest{
		SchemaVersion: 1,
		ID:            "com.elemo.lock",
		Name:          "Lock",
		Version:       "1.0.0",
		Requires:      model.PluginRequires{PluginAPI: "^1"},
		Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
	}
	require.NoError(t, manifest.Validate())
	require.NoError(t, reg.Put(context.Background(), plugin.LoadedPlugin{
		ID: manifest.ID, Version: "1.0.0", Manifest: manifest, Status: model.PluginStatusInstalled,
	}, nil))
	require.NoError(t, reg.Start(context.Background(), manifest.ID))

	var wg sync.WaitGroup
	rt.SetHandler(manifest.ID, func(string, []byte) ([]byte, error) {
		_, _ = reg.Get(manifest.ID)
		return []byte(`{"ok":true}`), nil
	})
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := reg.Call(context.Background(), manifest.ID, "x", nil)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}

func TestEncodeGuestRequest_PreservesScope(t *testing.T) {
	t.Parallel()
	body, err := plugin.EncodeGuestRequest("timer.start", []byte(`{"function":"timer.start","scopeId":"Issue:abc","payload":{"issueId":"abc"}}`))
	require.NoError(t, err)
	var req plugin.InvokeRequest
	require.NoError(t, plugin.DecodeJSON(body, &req))
	assert.Equal(t, "timer.start", req.Function)
	assert.Equal(t, "Issue:abc", req.ScopeID)
}

func TestEncodeGuestRequest_WrapsPayload(t *testing.T) {
	t.Parallel()
	body, err := plugin.EncodeGuestRequest("ping", []byte(`{"ok":true}`))
	require.NoError(t, err)
	var req plugin.InvokeRequest
	require.NoError(t, plugin.DecodeJSON(body, &req))
	assert.Equal(t, "ping", req.Function)
	assert.JSONEq(t, `{"ok":true}`, string(req.Payload))
}

func TestCapabilityForMethod(t *testing.T) {
	t.Parallel()
	capability, err := plugin.CapabilityForMethod("graph.nodes.create")
	require.NoError(t, err)
	assert.Equal(t, model.CapabilityGraphWrite, capability)

	capability, err = plugin.CapabilityForMethod("graph.nodes.move")
	require.NoError(t, err)
	assert.Equal(t, model.CapabilityGraphWrite, capability)

	_, err = plugin.CapabilityForMethod("neo4j.query")
	require.ErrorIs(t, err, plugin.ErrUnknownHostMethod)

	_, err = plugin.CapabilityForMethod("plugin.config.get")
	require.ErrorIs(t, err, plugin.ErrUnknownHostMethod)
}

func TestRelationType(t *testing.T) {
	t.Parallel()
	got, err := plugin.RelationType("com.elemo.timetracking", "LOGGED_ON")
	require.NoError(t, err)
	assert.Equal(t, "EXT__com_elemo_timetracking__LOGGED_ON", got)

	_, err = plugin.RelationType("com.elemo.timetracking", "IN_SCOPE_OF")
	require.Error(t, err)

	a, err := plugin.RelationType("com.acme.one", "LOGGED_ON")
	require.NoError(t, err)
	b, err := plugin.RelationType("com.acme.two", "LOGGED_ON")
	require.NoError(t, err)
	assert.NotEqual(t, a, b)
	assert.True(t, strings.HasPrefix(plugin.RelationPrefix("com.acme.one"), "EXT__"))
}

func TestReadWASM_NilBackend(t *testing.T) {
	t.Parallel()
	data, err := plugin.ReadWASM(plugin.ExtractedPackage{
		Manifest: model.PluginManifest{ID: "com.elemo.x"},
	})
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestRemoveVersion_LastVersionCleansParent(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	pluginID := "com.elemo.x"
	dir := plugin.InstallDirectory(base, pluginID, "1.0.0")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, plugin.RemoveVersion(base, pluginID, "1.0.0"))
	_, err := os.Stat(filepath.Join(base, pluginID))
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveVersion_KeepsSiblingVersions(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	pluginID := "com.elemo.x"
	require.NoError(t, os.MkdirAll(plugin.InstallDirectory(base, pluginID, "1.0.0"), 0o755))
	keep := plugin.InstallDirectory(base, pluginID, "1.1.0")
	require.NoError(t, os.MkdirAll(keep, 0o755))
	require.NoError(t, plugin.RemoveVersion(base, pluginID, "1.0.0"))
	_, err := os.Stat(keep)
	require.NoError(t, err)
}

func TestHostOKAndHostError(t *testing.T) {
	t.Parallel()

	ok, err := plugin.HostOK(map[string]any{"pong": true})
	require.NoError(t, err)
	assert.True(t, ok.OK)
	assert.JSONEq(t, `{"pong":true}`, string(ok.Data))

	nilOK, err := plugin.HostOK(nil)
	require.NoError(t, err)
	assert.True(t, nilOK.OK)
	assert.Empty(t, nilOK.Data)

	errResp := plugin.HostError(plugin.ErrUnknownHostMethod)
	assert.False(t, errResp.OK)
	assert.Equal(t, plugin.ErrUnknownHostMethod.Error(), errResp.Error)

	nilErr := plugin.HostError(nil)
	assert.True(t, nilErr.OK)
}

func TestInvokeError(t *testing.T) {
	t.Parallel()
	resp := plugin.InvokeError(plugin.ErrInvokeDisabled)
	assert.False(t, resp.OK)
	assert.Equal(t, plugin.ErrInvokeDisabled.Error(), resp.Error)

	ok := plugin.InvokeError(nil)
	assert.True(t, ok.OK)
}

func TestHasCapabilityAndRequireCapability(t *testing.T) {
	t.Parallel()
	manifest := model.PluginManifest{
		Capabilities: []model.PluginCapability{model.CapabilityIssuesRead},
	}
	assert.True(t, plugin.HasCapability(manifest, model.CapabilityIssuesRead))
	assert.False(t, plugin.HasCapability(manifest, model.CapabilityGraphWrite))
	require.NoError(t, plugin.RequireCapability(manifest, model.CapabilityIssuesRead))
	require.ErrorIs(t, plugin.RequireCapability(manifest, model.CapabilityGraphWrite), plugin.ErrCapabilityDenied)
}
