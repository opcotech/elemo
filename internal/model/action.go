package model

import (
	"strings"
)

const (
	ActionOrganizationCreate        Action = "organization.create"
	ActionOrganizationRead          Action = "organization.read"
	ActionOrganizationUpdate        Action = "organization.update"
	ActionOrganizationDelete        Action = "organization.delete"
	ActionOrganizationMembersManage Action = "organization.members.manage"
	ActionNamespaceCreate           Action = "namespace.create"
	ActionNamespaceRead             Action = "namespace.read"
	ActionNamespaceUpdate           Action = "namespace.update"
	ActionNamespaceDelete           Action = "namespace.delete"
	ActionProjectCreate             Action = "project.create"
	ActionProjectRead               Action = "project.read"
	ActionProjectUpdate             Action = "project.update"
	ActionProjectDelete             Action = "project.delete"
	ActionProjectMembersManage      Action = "project.members.manage"
	ActionIssueCreate               Action = "issue.create"
	ActionIssueRead                 Action = "issue.read"
	ActionIssueUpdate               Action = "issue.update"
	ActionIssueDelete               Action = "issue.delete"
	ActionIssueAssign               Action = "issue.assign"
	ActionDocumentCreate            Action = "document.create"
	ActionDocumentRead              Action = "document.read"
	ActionDocumentUpdate            Action = "document.update"
	ActionDocumentDelete            Action = "document.delete"
	ActionFolderCreate              Action = "folder.create"
	ActionRoleManage                Action = "role.manage"
	ActionTeamManage                Action = "team.manage"
	ActionPermissionManage          Action = "permission.manage"
	ActionCustomFieldManage         Action = "custom_field.manage"
	ActionPluginInstall             Action = "plugin.install"
	ActionPluginManage              Action = "plugin.manage"
	ActionExtensionCreate           Action = "extension.create"
	ActionExtensionRead             Action = "extension.read"
	ActionExtensionUpdate           Action = "extension.update"
	ActionExtensionDelete           Action = "extension.delete"
)

const (
	RoleKeyOrgAdmin           = "org-admin"
	RoleKeyOrgMember          = "org-member"
	RoleKeyNamespaceAdmin     = "namespace-admin"
	RoleKeyProjectMaintainer  = "project-maintainer"
	RoleKeyProjectViewer      = "project-viewer"
	RoleKeyIssueMaintainer    = "issue-maintainer"
	RoleKeyDocumentMaintainer = "document-maintainer"
)

const LabelPrincipal = "Principal"

// Action is a fine-grained authorization capability.
type Action string

// String returns the dotted action identifier.
func (a Action) String() string {
	return string(a)
}

// MarshalText encodes the action as its dotted identifier.
func (a Action) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText parses and validates an action identifier.
func (a *Action) UnmarshalText(text []byte) error {
	parsed := Action(strings.TrimSpace(string(text)))
	if !parsed.Valid() {
		return ErrInvalidAction
	}
	*a = parsed
	return nil
}

// Valid reports whether the action is one of the known Action values.
func (a Action) Valid() bool {
	_, ok := actionSet[a]
	return ok
}

// Validate returns ErrInvalidAction when the action is unknown.
func (a Action) Validate() error {
	if !a.Valid() {
		return ErrInvalidAction
	}
	return nil
}

var Actions = []Action{
	ActionOrganizationCreate,
	ActionOrganizationRead,
	ActionOrganizationUpdate,
	ActionOrganizationDelete,
	ActionOrganizationMembersManage,
	ActionNamespaceCreate,
	ActionNamespaceRead,
	ActionNamespaceUpdate,
	ActionNamespaceDelete,
	ActionProjectCreate,
	ActionProjectRead,
	ActionProjectUpdate,
	ActionProjectDelete,
	ActionProjectMembersManage,
	ActionIssueCreate,
	ActionIssueRead,
	ActionIssueUpdate,
	ActionIssueDelete,
	ActionIssueAssign,
	ActionDocumentCreate,
	ActionDocumentRead,
	ActionDocumentUpdate,
	ActionDocumentDelete,
	ActionFolderCreate,
	ActionRoleManage,
	ActionTeamManage,
	ActionPermissionManage,
	ActionCustomFieldManage,
	ActionPluginInstall,
	ActionPluginManage,
	ActionExtensionCreate,
	ActionExtensionRead,
	ActionExtensionUpdate,
	ActionExtensionDelete,
}

var actionSet = func() map[Action]struct{} {
	out := make(map[Action]struct{}, len(Actions))
	for _, action := range Actions {
		out[action] = struct{}{}
	}
	return out
}()

// RoleTemplate is a named, inspectable bundle of actions with no intrinsic authority.
type RoleTemplate struct {
	Key         string
	Name        string
	Description string
	Actions     []Action
}

// ActionStrings returns the dotted identifiers for the template's actions.
func (t RoleTemplate) ActionStrings() []string {
	out := make([]string, len(t.Actions))
	for i, action := range t.Actions {
		out[i] = action.String()
	}
	return out
}

var orgScopedActions = []Action{
	ActionOrganizationRead,
	ActionOrganizationUpdate,
	ActionOrganizationDelete,
	ActionOrganizationMembersManage,
	ActionNamespaceCreate,
	ActionNamespaceRead,
	ActionNamespaceUpdate,
	ActionNamespaceDelete,
	ActionProjectCreate,
	ActionProjectRead,
	ActionProjectUpdate,
	ActionProjectDelete,
	ActionProjectMembersManage,
	ActionIssueCreate,
	ActionIssueRead,
	ActionIssueUpdate,
	ActionIssueDelete,
	ActionIssueAssign,
	ActionDocumentCreate,
	ActionDocumentRead,
	ActionDocumentUpdate,
	ActionDocumentDelete,
	ActionFolderCreate,
	ActionRoleManage,
	ActionTeamManage,
	ActionPermissionManage,
	ActionCustomFieldManage,
	ActionPluginManage,
	ActionExtensionCreate,
	ActionExtensionRead,
	ActionExtensionUpdate,
	ActionExtensionDelete,
}

var namespaceScopedActions = []Action{
	ActionNamespaceRead,
	ActionNamespaceUpdate,
	ActionNamespaceDelete,
	ActionProjectCreate,
	ActionProjectRead,
	ActionProjectUpdate,
	ActionProjectDelete,
	ActionProjectMembersManage,
	ActionIssueCreate,
	ActionIssueRead,
	ActionIssueUpdate,
	ActionIssueDelete,
	ActionIssueAssign,
	ActionDocumentCreate,
	ActionDocumentRead,
	ActionDocumentUpdate,
	ActionDocumentDelete,
	ActionFolderCreate,
	ActionTeamManage,
	ActionPermissionManage,
	ActionCustomFieldManage,
	ActionPluginManage,
	ActionExtensionCreate,
	ActionExtensionRead,
	ActionExtensionUpdate,
	ActionExtensionDelete,
}

var projectMaintainerActions = []Action{
	ActionProjectRead,
	ActionProjectUpdate,
	ActionProjectDelete,
	ActionProjectMembersManage,
	ActionIssueCreate,
	ActionIssueRead,
	ActionIssueUpdate,
	ActionIssueDelete,
	ActionIssueAssign,
	ActionDocumentCreate,
	ActionDocumentRead,
	ActionDocumentUpdate,
	ActionDocumentDelete,
	ActionFolderCreate,
	ActionTeamManage,
	ActionPermissionManage,
	ActionCustomFieldManage,
	ActionPluginManage,
	ActionExtensionCreate,
	ActionExtensionRead,
	ActionExtensionUpdate,
	ActionExtensionDelete,
}

var projectViewerActions = []Action{
	ActionProjectRead,
	ActionIssueRead,
	ActionDocumentRead,
	ActionExtensionRead,
}

var issueMaintainerActions = []Action{
	ActionIssueRead,
	ActionIssueUpdate,
	ActionIssueDelete,
	ActionIssueAssign,
}

var documentMaintainerActions = []Action{
	ActionDocumentRead,
	ActionDocumentUpdate,
	ActionDocumentDelete,
	ActionFolderCreate,
}

// RoleTemplates are copied onto each new organization. None include organization.create.
var RoleTemplates = []RoleTemplate{
	{
		Key:         RoleKeyOrgAdmin,
		Name:        "Organization admin",
		Description: "Full authority within an organization scope, excluding organization.create.",
		Actions:     orgScopedActions,
	},
	{
		Key:         RoleKeyOrgMember,
		Name:        "Organization member",
		Description: "Read the organization they belong to, including plugin graph nodes.",
		Actions:     []Action{ActionOrganizationRead, ActionExtensionRead},
	},
	{
		Key:         RoleKeyNamespaceAdmin,
		Name:        "Namespace admin",
		Description: "Administer a namespace and its descendants.",
		Actions:     namespaceScopedActions,
	},
	{
		Key:         RoleKeyProjectMaintainer,
		Name:        "Project maintainer",
		Description: "Maintain a project and its issues and documents.",
		Actions:     projectMaintainerActions,
	},
	{
		Key:         RoleKeyProjectViewer,
		Name:        "Project viewer",
		Description: "Read a project and its issues and documents.",
		Actions:     projectViewerActions,
	},
	{
		Key:         RoleKeyIssueMaintainer,
		Name:        "Issue maintainer",
		Description: "Update and assign an issue.",
		Actions:     issueMaintainerActions,
	},
	{
		Key:         RoleKeyDocumentMaintainer,
		Name:        "Document maintainer",
		Description: "Update a document or folder.",
		Actions:     documentMaintainerActions,
	},
}

var roleTemplatesByKey = func() map[string]RoleTemplate {
	out := make(map[string]RoleTemplate, len(RoleTemplates))
	for _, tmpl := range RoleTemplates {
		out[tmpl.Key] = tmpl
	}
	return out
}()

// RoleTemplateByKey returns the RoleTemplate with the given key.
func RoleTemplateByKey(key string) (RoleTemplate, error) {
	tmpl, ok := roleTemplatesByKey[key]
	if !ok {
		return RoleTemplate{}, ErrUnknownRoleKey
	}
	return tmpl, nil
}

// ParseActions converts string identifiers to Actions, skipping duplicates.
// A nil slice becomes empty. It returns ErrInvalidAction if any value is unknown.
func ParseActions(values []string) ([]Action, error) {
	if values == nil {
		return []Action{}, nil
	}
	out := make([]Action, 0, len(values))
	seen := make(map[Action]struct{}, len(values))
	for _, value := range values {
		action := Action(strings.TrimSpace(value))
		if err := action.Validate(); err != nil {
			return nil, err
		}
		if _, exists := seen[action]; exists {
			continue
		}
		seen[action] = struct{}{}
		out = append(out, action)
	}
	return out, nil
}

// ActionStrings returns the dotted identifiers for actions, preserving order.
func ActionStrings(actions []Action) []string {
	out := make([]string, len(actions))
	for i, action := range actions {
		out[i] = action.String()
	}
	return out
}

// IsPrincipalType reports whether rt can hold grants (user, team, or organization).
func IsPrincipalType(rt ResourceType) bool {
	switch rt {
	case ResourceTypeUser, ResourceTypeTeam, ResourceTypeOrganization:
		return true
	default:
		return false
	}
}

// InstallationID returns the singleton installation resource ID used as the
// parent of organization.create.
func InstallationID() ID {
	return MustNewNilID(ResourceTypeInstallation)
}

// ReadActionFor returns the action that authorizes listing or reading rt.
// The installation maps to organization.create. The second result is false
// when rt has no read action.
func ReadActionFor(rt ResourceType) (Action, bool) {
	switch rt {
	case ResourceTypeInstallation:
		return ActionOrganizationCreate, true
	case ResourceTypeOrganization:
		return ActionOrganizationRead, true
	case ResourceTypeNamespace:
		return ActionNamespaceRead, true
	case ResourceTypeProject:
		return ActionProjectRead, true
	case ResourceTypeIssue:
		return ActionIssueRead, true
	case ResourceTypeDocument, ResourceTypeFolder:
		return ActionDocumentRead, true
	case ResourceTypeExtension:
		return ActionExtensionRead, true
	default:
		return "", false
	}
}

// UpdateActionFor returns the action that authorizes updating rt.
// The second result is false when rt has no update action.
func UpdateActionFor(rt ResourceType) (Action, bool) {
	switch rt {
	case ResourceTypeOrganization:
		return ActionOrganizationUpdate, true
	case ResourceTypeNamespace:
		return ActionNamespaceUpdate, true
	case ResourceTypeProject:
		return ActionProjectUpdate, true
	case ResourceTypeIssue:
		return ActionIssueUpdate, true
	case ResourceTypeDocument, ResourceTypeFolder:
		return ActionDocumentUpdate, true
	case ResourceTypeExtension:
		return ActionExtensionUpdate, true
	default:
		return "", false
	}
}

// IsCustomFieldScopeType reports whether rt may own a custom-field definition.
func IsCustomFieldScopeType(rt ResourceType) bool {
	switch rt {
	case ResourceTypeOrganization, ResourceTypeNamespace, ResourceTypeProject:
		return true
	default:
		return false
	}
}

// IsCustomFieldTargetType reports whether rt may carry custom-field values.
func IsCustomFieldTargetType(rt ResourceType) bool {
	switch rt {
	case ResourceTypeIssue,
		ResourceTypeDocument,
		ResourceTypeFolder,
		ResourceTypeProject,
		ResourceTypeNamespace,
		ResourceTypeOrganization:
		return true
	default:
		return false
	}
}

// IsCustomFieldReferenceType reports whether rt may be selected as a
// resource_reference allowed type. Permission and the meta ResourceType
// sentinel are excluded. Plugin-registered node types should be accepted
// here when they exist.
func IsCustomFieldReferenceType(rt ResourceType) bool {
	if !rt.IsAResourceType() {
		return false
	}
	return rt != ResourceTypeKind && rt != ResourceTypePermission
}
