package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validManifest() PluginManifest {
	return PluginManifest{
		SchemaVersion: PluginSchemaVersionV1,
		ID:            "com.elemo.timetracking",
		Name:          "Time Tracking",
		Version:       "1.0.0",
		Requires: PluginRequires{
			PluginAPI: "^1",
		},
		Backend: &PluginBackendDecl{Entry: PluginBackendWASMPath},
		Capabilities: []PluginCapability{
			CapabilityIssuesRead,
			CapabilityGraphRead,
			CapabilityGraphWrite,
		},
		Graph: &PluginGraphSchema{
			Nodes: []PluginGraphNodeDecl{
				{
					Kind:  "TimeEntry",
					Scope: PluginGraphNodeScope{Parent: "Issue"},
					Properties: []PluginGraphPropertyDecl{
						{Name: "seconds", Type: PluginGraphPropertyTypeInteger, Required: true},
						{Name: "note", Type: PluginGraphPropertyTypeStr},
					},
				},
			},
			Relations: []PluginGraphRelationDecl{
				{
					Kind:        "LOGGED_ON",
					From:        "TimeEntry",
					To:          "Issue",
					Cardinality: PluginGraphCardinalityManyToOne,
				},
			},
		},
	}
}

func TestPluginManifest_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		require.NoError(t, m.Validate())
	})

	t.Run("unknown schema version", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.SchemaVersion = 99
		require.ErrorIs(t, m.Validate(), ErrPluginSchemaVersion)
	})

	t.Run("incompatible plugin api", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.Requires.PluginAPI = "2"
		require.ErrorIs(t, m.Validate(), ErrPluginAPIIncompatible)
	})

	t.Run("invalid id", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.ID = "NotADomain"
		require.ErrorIs(t, m.Validate(), ErrInvalidPluginID)
	})

	t.Run("unknown capability", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.Capabilities = []PluginCapability{"cypher.write"}
		require.ErrorIs(t, m.Validate(), ErrInvalidPluginCapability)
	})

	t.Run("reserved relation kind", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.Graph.Relations[0].Kind = "IN_SCOPE_OF"
		require.Error(t, m.Validate())
	})

	t.Run("reserved property name", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.Graph.Nodes[0].Properties[0].Name = "plugin_id"
		require.Error(t, m.Validate())
	})

	t.Run("scope cycle", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.Graph.Nodes = []PluginGraphNodeDecl{
			{Kind: "Account", Scope: PluginGraphNodeScope{Parent: "Budget"}},
			{Kind: "Budget", Scope: PluginGraphNodeScope{Parent: "Account"}},
		}
		m.Graph.Relations = nil
		require.Error(t, m.Validate())
	})

	t.Run("scope chain to core", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.Graph.Nodes = []PluginGraphNodeDecl{
			{
				Kind:  "Account",
				Scope: PluginGraphNodeScope{Parent: "Organization"},
				Properties: []PluginGraphPropertyDecl{
					{Name: "name", Type: PluginGraphPropertyTypeStr, Required: true},
				},
			},
			{
				Kind:  "Budget",
				Scope: PluginGraphNodeScope{Parent: "Account"},
				Properties: []PluginGraphPropertyDecl{
					{Name: "amount", Type: PluginGraphPropertyTypeDecimal, Required: true},
				},
			},
		}
		m.Graph.Relations = []PluginGraphRelationDecl{
			{
				Kind:        "HAS_BUDGET",
				From:        "Account",
				To:          "Budget",
				Cardinality: PluginGraphCardinalityOneToMany,
			},
		}
		require.NoError(t, m.Validate())
	})

	t.Run("foreign alias relation", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.Graph.Foreign = []PluginGraphForeignDecl{
			{
				Name:   "LoggedTime",
				Parent: "Issue",
				Properties: []PluginGraphPropertyDecl{
					{Name: "seconds", Type: PluginGraphPropertyTypeInteger, Required: true},
				},
			},
		}
		m.Graph.Nodes = append(m.Graph.Nodes, PluginGraphNodeDecl{
			Kind:  "Budget",
			Scope: PluginGraphNodeScope{Parent: "Organization"},
			Properties: []PluginGraphPropertyDecl{
				{Name: "seconds", Type: PluginGraphPropertyTypeInteger, Required: true},
			},
		})
		m.Graph.Relations = append(m.Graph.Relations, PluginGraphRelationDecl{
			Kind:        "COUNTED_AGAINST",
			From:        "LoggedTime",
			To:          "Budget",
			Cardinality: PluginGraphCardinalityManyToOne,
		})
		m.Config = []PluginConfigFieldDecl{
			{Name: "time_source", Type: PluginConfigFieldTypeGraphBinding, Foreign: "LoggedTime"},
		}
		require.NoError(t, m.Validate())
	})

	t.Run("foreign collides with local kind", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.Graph.Foreign = []PluginGraphForeignDecl{
			{Name: "TimeEntry", Parent: "Issue"},
		}
		require.Error(t, m.Validate())
	})

	t.Run("config graph_binding missing foreign", func(t *testing.T) {
		t.Parallel()
		m := validManifest()
		m.Config = []PluginConfigFieldDecl{
			{Name: "time_source", Type: PluginConfigFieldTypeGraphBinding},
		}
		require.Error(t, m.Validate())
	})
}

func TestBindingMatches(t *testing.T) {
	t.Parallel()

	fields := []PluginConfigFieldDecl{
		{Name: "time_source", Type: PluginConfigFieldTypeGraphBinding, Foreign: "LoggedTime"},
	}
	config := []byte(`{"time_source":{"plugin_id":"com.elemo.timetracking","kind":"TimeEntry"}}`)

	assert.True(t, BindingMatches(config, fields, "com.elemo.timetracking", "TimeEntry"))
	assert.False(t, BindingMatches(config, fields, "com.elemo.timetracking", "Other"))
	assert.False(t, BindingMatches(config, fields, "com.elemo.other", "TimeEntry"))
	assert.False(t, BindingMatches([]byte(`{}`), fields, "com.elemo.timetracking", "TimeEntry"))

	binding, ok := GraphBinding(config, fields, "LoggedTime")
	require.True(t, ok)
	assert.Equal(t, "com.elemo.timetracking", binding.PluginID)
	assert.Equal(t, "TimeEntry", binding.Kind)
}

func TestNamespacedRelationType(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "EXT__com_elemo_timetracking__LOGGED_ON",
		NamespacedRelationType("com.elemo.timetracking", "LOGGED_ON"))
}

func TestNewExtension(t *testing.T) {
	t.Parallel()
	ext, err := NewExtension("com.elemo.timetracking", "TimeEntry", map[string]any{"seconds": int64(60)})
	require.NoError(t, err)
	assert.Equal(t, ResourceTypeExtension, ext.ID.Type)
	assert.False(t, ext.ID.IsNil())
}

func TestPluginCapability_Valid(t *testing.T) {
	t.Parallel()
	assert.True(t, CapabilityGraphWrite.Valid())
	assert.False(t, PluginCapability("fs.write").Valid())
}

func TestIsPluginActivationScopeType(t *testing.T) {
	t.Parallel()
	assert.True(t, IsPluginActivationScopeType(ResourceTypeOrganization))
	assert.True(t, IsPluginActivationScopeType(ResourceTypeProject))
	assert.False(t, IsPluginActivationScopeType(ResourceTypeIssue))
}

func TestPluginStatus_ServesFrontend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status PluginStatus
		want   bool
	}{
		{PluginStatusUnknown, false},
		{PluginStatusInstalled, true},
		{PluginStatusStarting, true},
		{PluginStatusActive, true},
		{PluginStatusDisabling, false},
		{PluginStatusDisabled, false},
		{PluginStatusFailed, true},
	}
	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.status.ServesFrontend())
		})
	}
}
