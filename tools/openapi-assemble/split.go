package main

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type fragment struct {
	file string
	keys []string
}

func pathFragments() []fragment {
	return []fragment{
		{
			file: "paths/users.yaml",
			keys: []string{
				"/v1/users",
				"/v1/users/reset",
				"/v1/users/{id}",
				"/v1/users/{id}/issues",
			},
		},
		{
			file: "paths/labels.yaml",
			keys: []string{"/v1/labels"},
		},
		{
			file: "paths/todos.yaml",
			keys: []string{"/v1/todos", "/v1/todos/{id}"},
		},
		{
			file: "paths/notifications.yaml",
			keys: []string{"/v1/notifications", "/v1/notifications/{id}"},
		},
		{
			file: "paths/organizations.yaml",
			keys: []string{
				"/v1/organizations",
				"/v1/organizations/{id}",
				"/v1/organizations/{id}/members",
				"/v1/organizations/{id}/members/invite",
				"/v1/organizations/{id}/members/{user_id}",
				"/v1/organizations/{id}/members/{user_id}/invite",
				"/v1/organizations/{id}/members/accept",
			},
		},
		{
			file: "paths/roles.yaml",
			keys: []string{
				"/v1/organizations/{id}/roles",
				"/v1/organizations/{id}/roles/{role_id}",
			},
		},
		{
			file: "paths/teams.yaml",
			keys: []string{
				"/v1/organizations/{id}/teams",
				"/v1/organizations/{id}/teams/{team_id}",
				"/v1/organizations/{id}/teams/{team_id}/members",
				"/v1/organizations/{id}/teams/{team_id}/members/{user_id}",
			},
		},
		{
			file: "paths/namespaces.yaml",
			keys: []string{
				"/v1/organizations/{id}/namespaces",
				"/v1/namespaces",
				"/v1/namespaces/{id}",
			},
		},
		{
			file: "paths/projects.yaml",
			keys: []string{
				"/v1/namespaces/{id}/projects",
				"/v1/projects/{id}",
			},
		},
		{
			file: "paths/issues.yaml",
			keys: []string{
				"/v1/namespaces/{id}/issues",
				"/v1/namespaces/{id}/issues/{key}",
				"/v1/projects/{id}/issues",
				"/v1/issues/{id}",
				"/v1/issues/{id}/relations",
				"/v1/issues/{id}/relations/{relation_id}",
			},
		},
		{
			file: "paths/documents.yaml",
			keys: []string{
				"/v1/organizations/{id}/documents",
				"/v1/namespaces/{id}/documents",
				"/v1/projects/{id}/documents",
				"/v1/projects/{id}/documents/{documentId}",
				"/v1/documents/{id}",
				"/v1/issues/{id}/documents",
				"/v1/issues/{id}/documents/{documentId}",
			},
		},
		{
			file: "paths/folders.yaml",
			keys: []string{
				"/v1/organizations/{id}/folders",
				"/v1/namespaces/{id}/folders",
				"/v1/folders/{id}",
			},
		},
		{
			file: "paths/permissions.yaml",
			keys: []string{
				"/v1/permissions",
				"/v1/permissions/{id}",
				"/v1/permissions/resources/{resourceId}",
			},
		},
		{
			file: "paths/search.yaml",
			keys: []string{"/v1/search"},
		},
		{
			file: "paths/system.yaml",
			keys: []string{
				"/v1/system/health",
				"/v1/system/heartbeat",
				"/v1/system/license",
				"/v1/system/version",
			},
		},
	}
}

func schemaFragments() []fragment {
	return []fragment{
		{
			file: "components/schemas/common.yaml",
			keys: []string{"PageInfo", "HTTPError", "Language"},
		},
		{
			file: "components/schemas/user.yaml",
			keys: []string{"User", "UserStatus", "PartialUser", "UserPage"},
		},
		{
			file: "components/schemas/organization.yaml",
			keys: []string{
				"OrganizationMember",
				"Organization",
				"OrganizationStatus",
				"PartialOrganization",
				"OrganizationPage",
				"OrganizationMemberPage",
			},
		},
		{
			file: "components/schemas/namespace.yaml",
			keys: []string{
				"Namespace",
				"AccessibleNamespace",
				"PartialNamespace",
				"NamespacePage",
				"AccessibleNamespacePage",
			},
		},
		{
			file: "components/schemas/project.yaml",
			keys: []string{"ProjectStatus", "PartialProject", "Project", "ProjectPage"},
		},
		{
			file: "components/schemas/issue.yaml",
			keys: []string{
				"PartialIssue",
				"IssueKind",
				"IssueStatus",
				"IssuePriority",
				"IssueResolution",
				"IssueRelationKind",
				"IssueRelationDirection",
				"IssueLink",
				"Issue",
				"IssueRelation",
				"PartialIssuePage",
				"IssueRelationPage",
			},
		},
		{
			file: "components/schemas/document.yaml",
			keys: []string{
				"PartialDocument",
				"Document",
				"DocumentLibrary",
				"DocumentFolder",
				"DocumentRelation",
				"PartialDocumentPage",
			},
		},
		{
			file: "components/schemas/folder.yaml",
			keys: []string{"Folder", "FolderPage"},
		},
		{
			file: "components/schemas/label.yaml",
			keys: []string{"PartialLabel", "Label", "LabelPage"},
		},
		{
			file: "components/schemas/todo.yaml",
			keys: []string{"Todo", "TodoPriority", "TodoPage"},
		},
		{
			file: "components/schemas/notification.yaml",
			keys: []string{"Notification", "NotificationPage"},
		},
		{
			file: "components/schemas/permission.yaml",
			keys: []string{
				"Action",
				"Grant",
				"EffectiveActions",
				"GrantPrincipalType",
				"Role",
				"ResourceType",
				"RolePage",
			},
		},
		{
			file: "components/schemas/team.yaml",
			keys: []string{"Team", "TeamPage"},
		},
		{
			file: "components/schemas/search.yaml",
			keys: []string{"SearchResult", "SearchPage"},
		},
		{
			file: "components/schemas/system.yaml",
			keys: []string{"SystemHealth", "SystemVersion", "SystemLicense"},
		},
	}
}

func requestBodyFragments() []fragment {
	return []fragment{
		{
			file: "components/request-bodies/user.yaml",
			keys: []string{"UserPatch", "UserCreate", "UserPasswordReset"},
		},
		{
			file: "components/request-bodies/organization.yaml",
			keys: []string{"OrganizationInvitationAccept", "OrganizationCreate", "OrganizationPatch"},
		},
		{
			file: "components/request-bodies/namespace.yaml",
			keys: []string{"NamespaceCreate", "NamespacePatch"},
		},
		{
			file: "components/request-bodies/project.yaml",
			keys: []string{"ProjectCreate", "ProjectPatch"},
		},
		{
			file: "components/request-bodies/issue.yaml",
			keys: []string{"IssueCreate", "IssuePatch", "IssueRelationCreate", "IssueRelationPatch"},
		},
		{
			file: "components/request-bodies/document.yaml",
			keys: []string{"DocumentCreate", "DocumentPatch"},
		},
		{
			file: "components/request-bodies/folder.yaml",
			keys: []string{"FolderCreate", "FolderPatch"},
		},
		{
			file: "components/request-bodies/todo.yaml",
			keys: []string{"TodoCreate", "TodoPatch"},
		},
		{
			file: "components/request-bodies/notification.yaml",
			keys: []string{"NotificationPatch"},
		},
		{
			file: "components/request-bodies/permission.yaml",
			keys: []string{"GrantCreate", "RoleCreate", "RolePatch"},
		},
		{
			file: "components/request-bodies/team.yaml",
			keys: []string{"TeamCreate", "TeamPatch"},
		},
	}
}

func split(bundlePath, srcDir string, layout Layout) error {
	root, err := loadMappingFile(bundlePath)
	if err != nil {
		return err
	}

	if err := writeRootFragment(root, srcDir); err != nil {
		return err
	}

	_, components := mappingLookup(root, "components")
	if components == nil {
		return fmt.Errorf("missing components in %s", bundlePath)
	}
	_, paths := mappingLookup(root, "paths")
	if paths == nil {
		return fmt.Errorf("missing paths in %s", bundlePath)
	}

	if err := writeFragments(srcDir, paths, pathFragments(), "paths"); err != nil {
		return err
	}

	_, schemas := mappingLookup(components, "schemas")
	if schemas == nil {
		return fmt.Errorf("missing components.schemas in %s", bundlePath)
	}
	if err := writeFragments(srcDir, schemas, schemaFragments(), "components.schemas"); err != nil {
		return err
	}

	_, requestBodies := mappingLookup(components, "requestBodies")
	if requestBodies == nil {
		return fmt.Errorf("missing components.requestBodies in %s", bundlePath)
	}
	if err := writeFragments(srcDir, requestBodies, requestBodyFragments(), "components.requestBodies"); err != nil {
		return err
	}

	if err := copyComponentFile(srcDir, components, "parameters", layout.ParametersFile); err != nil {
		return err
	}
	if err := copyComponentFile(srcDir, components, "responses", layout.ResponsesFile); err != nil {
		return err
	}
	if err := copyComponentFile(srcDir, components, "securitySchemes", layout.SecuritySchemesFile); err != nil {
		return err
	}

	return nil
}

func writeRootFragment(root *yaml.Node, srcDir string) error {
	out, err := extractKeys(root, []string{"openapi", "info", "tags", "servers"}, "root")
	if err != nil {
		return err
	}
	return writeYAMLFile(filepath.Join(srcDir, rootFile), out, "")
}

func writeFragments(srcDir string, src *yaml.Node, fragments []fragment, section string) error {
	assigned := make(map[string]string, len(src.Content)/2)
	for _, frag := range fragments {
		node, err := extractKeys(src, frag.keys, section)
		if err != nil {
			return err
		}
		for _, key := range frag.keys {
			if prev, ok := assigned[key]; ok {
				return fmt.Errorf("layout duplicate key %q: %s and %s", key, prev, frag.file)
			}
			assigned[key] = frag.file
		}
		if err := writeYAMLFile(filepath.Join(srcDir, frag.file), node, ""); err != nil {
			return err
		}
	}

	for _, key := range mappingKeys(src) {
		if _, ok := assigned[key]; !ok {
			return fmt.Errorf("unassigned key %q in %s", key, section)
		}
	}

	return nil
}

func copyComponentFile(srcDir string, components *yaml.Node, key, rel string) error {
	_, value := mappingLookup(components, key)
	if value == nil {
		return fmt.Errorf("missing components.%s", key)
	}
	return writeYAMLFile(filepath.Join(srcDir, rel), value, "")
}
