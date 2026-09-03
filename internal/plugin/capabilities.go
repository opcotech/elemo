package plugin

import (
	"github.com/opcotech/elemo/internal/model"
)

// HasCapability reports whether the plugin was granted the capability at install.
func HasCapability(manifest model.PluginManifest, capability model.PluginCapability) bool {
	return manifest.HasCapability(capability)
}

// RequireCapability returns ErrCapabilityDenied when the capability is missing.
func RequireCapability(manifest model.PluginManifest, capability model.PluginCapability) error {
	if !manifest.HasCapability(capability) {
		return ErrCapabilityDenied
	}
	return nil
}

func methodCapability(method string) (model.PluginCapability, bool) {
	switch method {
	case "issues.get", "issues.list":
		return model.CapabilityIssuesRead, true
	case "issues.update":
		return model.CapabilityIssuesUpdate, true
	case "projects.get", "projects.list":
		return model.CapabilityProjectsRead, true
	case "users.get":
		return model.CapabilityUsersRead, true
	case "permissions.check":
		return model.CapabilityPermissionsCheck, true
	case "plugin.storage.get", "plugin.storage.list":
		return model.CapabilityPluginStorageRead, true
	case "plugin.storage.set", "plugin.storage.delete":
		return model.CapabilityPluginStorageWrite, true
	case "graph.nodes.get", "graph.nodes.list", "graph.relations.list":
		return model.CapabilityGraphRead, true
	case "graph.nodes.create", "graph.nodes.update", "graph.nodes.delete",
		"graph.nodes.move", "graph.relations.create", "graph.relations.delete":
		return model.CapabilityGraphWrite, true
	case "events.publish":
		return model.CapabilityEventsPublish, true
	default:
		return "", false
	}
}

// CapabilityForMethod maps a host method name to its required capability.
func CapabilityForMethod(method string) (model.PluginCapability, error) {
	capability, ok := methodCapability(method)
	if !ok {
		return "", ErrUnknownHostMethod
	}
	return capability, nil
}
