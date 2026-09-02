package main

// Layout describes which fragment files to merge, in output order.
type Layout struct {
	PathFiles           []string
	SchemaFiles         []string
	RequestBodyFiles    []string
	ParametersFile      string
	ResponsesFile       string
	SecuritySchemesFile string
}

// elemoLayout is the production OpenAPI fragment layout.
func elemoLayout() Layout {
	return Layout{
		PathFiles: []string{
			"paths/users.yaml",
			"paths/labels.yaml",
			"paths/todos.yaml",
			"paths/notifications.yaml",
			"paths/organizations.yaml",
			"paths/roles.yaml",
			"paths/teams.yaml",
			"paths/namespaces.yaml",
			"paths/projects.yaml",
			"paths/issues.yaml",
			"paths/documents.yaml",
			"paths/folders.yaml",
			"paths/permissions.yaml",
			"paths/search.yaml",
			"paths/custom_fields.yaml",
			"paths/system.yaml",
		},
		SchemaFiles: []string{
			"components/schemas/common.yaml",
			"components/schemas/user.yaml",
			"components/schemas/organization.yaml",
			"components/schemas/namespace.yaml",
			"components/schemas/project.yaml",
			"components/schemas/issue.yaml",
			"components/schemas/document.yaml",
			"components/schemas/folder.yaml",
			"components/schemas/label.yaml",
			"components/schemas/todo.yaml",
			"components/schemas/notification.yaml",
			"components/schemas/permission.yaml",
			"components/schemas/team.yaml",
			"components/schemas/search.yaml",
			"components/schemas/custom_field.yaml",
			"components/schemas/system.yaml",
		},
		RequestBodyFiles: []string{
			"components/request-bodies/user.yaml",
			"components/request-bodies/organization.yaml",
			"components/request-bodies/namespace.yaml",
			"components/request-bodies/project.yaml",
			"components/request-bodies/issue.yaml",
			"components/request-bodies/document.yaml",
			"components/request-bodies/folder.yaml",
			"components/request-bodies/todo.yaml",
			"components/request-bodies/notification.yaml",
			"components/request-bodies/permission.yaml",
			"components/request-bodies/team.yaml",
			"components/request-bodies/custom_field.yaml",
		},
		ParametersFile:      "components/parameters.yaml",
		ResponsesFile:       "components/responses.yaml",
		SecuritySchemesFile: "components/security-schemes.yaml",
	}
}
