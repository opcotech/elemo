package plugin

import (
	"fmt"
	"regexp"

	"github.com/opcotech/elemo/internal/model"
)

var namespacedRelPattern = regexp.MustCompile(`^EXT__[A-Za-z0-9_]+__[A-Z][A-Za-z0-9_]+$`)

// RelationType returns the allowlisted Neo4j type for a plugin domain edge.
func RelationType(pluginID, kind string) (string, error) {
	if model.IsReservedEdgeKind(kind) {
		return "", fmt.Errorf("relation kind %s is reserved", kind)
	}
	out := model.NamespacedRelationType(pluginID, kind)
	if !namespacedRelPattern.MatchString(out) {
		return "", fmt.Errorf("invalid namespaced relation type")
	}
	return out, nil
}

// RelationTypeFromManifest validates kind against the plugin schema then
// returns the namespaced Neo4j type.
func RelationTypeFromManifest(manifest model.PluginManifest, kind string) (string, error) {
	if manifest.Graph == nil {
		return "", model.ErrPluginGraphSchema
	}
	if _, ok := manifest.Graph.RelationKind(kind); !ok {
		return "", fmt.Errorf("undeclared relation kind %s", kind)
	}
	return RelationType(manifest.ID, kind)
}

// RelationPrefix is the uninstall match prefix for a plugin's domain edges.
func RelationPrefix(pluginID string) string {
	return model.NamespacedRelationType(pluginID, "")
}
