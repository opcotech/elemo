package service_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/service"
)

// methodKind is the entitlement review classification for a service method.
type methodKind string

const (
	kindQuery     methodKind = "query"
	kindMutation  methodKind = "mutation"
	kindException methodKind = "exception"
)

// classifiedServiceMethods is the review checklist for every exported service
// method. New methods must be added here as query, mutation, or an explicit
// operational exception so they cannot bypass entitlement review.
var classifiedServiceMethods = map[reflect.Type]map[string]methodKind{
	reflect.TypeOf((*service.CustomFieldService)(nil)).Elem(): {
		"AbortForResource":  kindMutation,
		"ArchiveDefinition": kindMutation,
		"CommitForResource": kindMutation,
		"CreateDefinition":  kindMutation,
		"DeleteDefinition":  kindMutation,
		"DeleteForResource": kindMutation,
		"DeleteValue":       kindMutation,
		"GetDefinition":     kindQuery,
		"ListDefinitions":   kindQuery,
		"ListEffective":     kindQuery,
		"ReconcilePending":  kindException,
		"Search":            kindQuery,
		"SetValue":          kindMutation,
		"StageForResource":  kindMutation,
		"UpdateDefinition":  kindMutation,
	},
	reflect.TypeOf((*service.DocumentService)(nil)).Elem(): {
		"Create":       kindMutation,
		"Delete":       kindMutation,
		"Get":          kindQuery,
		"ListLibrary":  kindQuery,
		"ListRelated":  kindQuery,
		"MoveLibrary":  kindMutation,
		"MoveToFolder": kindMutation,
		"Relate":       kindMutation,
		"Unrelate":     kindMutation,
		"Update":       kindMutation,
	},
	reflect.TypeOf((*service.EmailService)(nil)).Elem(): {
		"SendAuthPasswordResetEmail":      kindException,
		"SendLicenseExpiryEmail":          kindException,
		"SendOrganizationInvitationEmail": kindException,
		"SendUserWelcomeEmail":            kindException,
	},
	reflect.TypeOf((*service.EntitlementService)(nil)).Elem(): {
		"Get": kindQuery,
	},
	reflect.TypeOf((*service.FolderService)(nil)).Elem(): {
		"Create": kindMutation,
		"Delete": kindMutation,
		"Get":    kindQuery,
		"List":   kindQuery,
		"Update": kindMutation,
	},
	reflect.TypeOf((*service.IssueService)(nil)).Elem(): {
		"AddRelation":     kindMutation,
		"Create":          kindMutation,
		"Delete":          kindMutation,
		"Get":             kindQuery,
		"GetByKey":        kindQuery,
		"List":            kindQuery,
		"ListByNamespace": kindQuery,
		"ListByUser":      kindQuery,
		"ListRelations":   kindQuery,
		"RemoveRelation":  kindMutation,
		"Update":          kindMutation,
		"UpdateRelation":  kindMutation,
	},
	reflect.TypeOf((*service.LabelService)(nil)).Elem(): {
		"List": kindQuery,
	},
	reflect.TypeOf((*service.NamespaceService)(nil)).Elem(): {
		"Create":         kindMutation,
		"Delete":         kindMutation,
		"Get":            kindQuery,
		"GetByRef":       kindQuery,
		"List":           kindQuery,
		"ListAccessible": kindQuery,
		"Resolve":        kindQuery,
		"Update":         kindMutation,
	},
	reflect.TypeOf((*service.NotificationService)(nil)).Elem(): {
		"Create":          kindMutation,
		"Delete":          kindMutation,
		"Get":             kindQuery,
		"ListByRecipient": kindQuery,
		"Update":          kindMutation,
	},
	reflect.TypeOf((*service.OrganizationService)(nil)).Elem(): {
		"AcceptInvitation": kindMutation,
		"AddMember":        kindMutation,
		"Create":           kindMutation,
		"Delete":           kindMutation,
		"Get":              kindQuery,
		"GetByRef":         kindQuery,
		"InviteMember":     kindMutation,
		"List":             kindQuery,
		"ListMembers":      kindQuery,
		"RemoveMember":     kindMutation,
		"Resolve":          kindQuery,
		"RevokeInvitation": kindMutation,
		"Update":           kindMutation,
	},
	reflect.TypeOf((*service.PermissionService)(nil)).Elem(): {
		"BootstrapCreator":        kindMutation,
		"BumpGeneration":          kindException,
		"Create":                  kindMutation,
		"CtxUserCreate":           kindMutation,
		"CtxUserDelete":           kindMutation,
		"CtxUserEffectiveActions": kindQuery,
		"CtxUserHas":              kindQuery,
		"CtxUserListGrantScopes":  kindQuery,
		"Delete":                  kindMutation,
		"EffectiveActions":        kindQuery,
		"Explain":                 kindQuery,
		"Get":                     kindQuery,
		"GrantRole":               kindMutation,
		"Has":                     kindQuery,
		"LinkInScopeOf":           kindMutation,
		"ListByPrincipal":         kindQuery,
		"ListByScope":             kindQuery,
		"ListGrantScopes":         kindQuery,
		"ListScopeAncestry":       kindQuery,
	},
	reflect.TypeOf((*service.PluginService)(nil)).Elem(): {
		"AssetPath":        kindQuery,
		"CreateNode":       kindMutation,
		"CreateRelation":   kindMutation,
		"DeleteNode":       kindMutation,
		"DeleteRelation":   kindMutation,
		"Disable":          kindMutation,
		"Enable":           kindMutation,
		"Get":              kindQuery,
		"GetConfig":        kindQuery,
		"GetManagedConfig": kindQuery,
		"GetNode":          kindQuery,
		"Install":          kindMutation,
		"Invoke":           kindMutation,
		"List":             kindQuery,
		"ListFrontend":     kindQuery,
		"ListManaged":      kindQuery,
		"ListNodes":        kindQuery,
		"ListRelations":    kindQuery,
		"MoveNode":         kindMutation,
		"OpenAsset":        kindQuery,
		"Restore":          kindException,
		"SetConfig":        kindMutation,
		"Uninstall":        kindMutation,
		"UpdateNode":       kindMutation,
		"Upgrade":          kindMutation,
	},
	reflect.TypeOf((*service.ProjectService)(nil)).Elem(): {
		"Create":   kindMutation,
		"Delete":   kindMutation,
		"Get":      kindQuery,
		"GetByKey": kindQuery,
		"List":     kindQuery,
		"Update":   kindMutation,
	},
	reflect.TypeOf((*service.RoleService)(nil)).Elem(): {
		"AddMember":     kindMutation,
		"Create":        kindMutation,
		"Delete":        kindMutation,
		"Get":           kindQuery,
		"ListBelongsTo": kindQuery,
		"ListMembers":   kindQuery,
		"RemoveMember":  kindMutation,
		"Update":        kindMutation,
	},
	reflect.TypeOf((*service.SearchService)(nil)).Elem(): {
		"Delete":        kindException,
		"DeleteAll":     kindException,
		"DeleteByScope": kindException,
		"EnqueueIndex":  kindException,
		"Index":         kindException,
		"IndexIDs":      kindException,
		"Reindex":       kindException,
		"Search":        kindQuery,
	},
	reflect.TypeOf((*service.StaticFileService)(nil)).Elem(): {
		"Create": kindMutation,
		"Delete": kindMutation,
		"Get":    kindQuery,
		"Update": kindMutation,
	},
	reflect.TypeOf((*service.SystemService)(nil)).Elem(): {
		"GetHealth":    kindQuery,
		"GetHeartbeat": kindQuery,
		"GetVersion":   kindQuery,
	},
	reflect.TypeOf((*service.TeamService)(nil)).Elem(): {
		"AddMember":     kindMutation,
		"Create":        kindMutation,
		"Delete":        kindMutation,
		"Get":           kindQuery,
		"ListBelongsTo": kindQuery,
		"ListMembers":   kindQuery,
		"RemoveMember":  kindMutation,
		"Update":        kindMutation,
	},
	reflect.TypeOf((*service.TodoService)(nil)).Elem(): {
		"Create": kindMutation,
		"Delete": kindMutation,
		"Get":    kindQuery,
		"List":   kindQuery,
		"Update": kindMutation,
	},
	reflect.TypeOf((*service.UserService)(nil)).Elem(): {
		"Create":      kindMutation,
		"CreateToken": kindException,
		"Delete":      kindMutation,
		"DeleteToken": kindException,
		"Get":         kindQuery,
		"GetByEmail":  kindQuery,
		"List":        kindQuery,
		"Update":      kindMutation,
		"VerifyToken": kindException,
	},
}

func TestServiceMethodEntitlementClassification(t *testing.T) {
	t.Parallel()

	for iface, classified := range classifiedServiceMethods {
		t.Run(iface.Name(), func(t *testing.T) {
			t.Parallel()
			require.Equal(t, reflect.Interface, iface.Kind())

			seen := make(map[string]struct{}, iface.NumMethod())
			for i := 0; i < iface.NumMethod(); i++ {
				name := iface.Method(i).Name
				seen[name] = struct{}{}
				kind, ok := classified[name]
				require.Truef(t, ok, "unclassified %s.%s: add it as query, mutation, or exception", iface.Name(), name)
				require.Contains(t, []methodKind{kindQuery, kindMutation, kindException}, kind)
			}

			for name := range classified {
				_, ok := seen[name]
				require.Truef(t, ok, "stale classification for %s.%s", iface.Name(), name)
			}
		})
	}
}
