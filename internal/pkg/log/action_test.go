package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAction_String(t *testing.T) {
	tests := []struct {
		name   string
		action Action
		want   string
	}{
		{
			name:   "action handle_http_request",
			action: ActionHTTPRequestHandle,
			want:   "handle_http_request",
		},
		{
			name:   "action create_collection",
			action: ActionDBCollectionCreate,
			want:   "create_collection",
		},
		{
			name:   "action create_document",
			action: ActionDBDocumentCreate,
			want:   "create_document",
		},
		{
			name:   "action delete_document",
			action: ActionDBDocumentDelete,
			want:   "delete_document",
		},
		{
			name:   "action read_document",
			action: ActionDBDocumentRead,
			want:   "read_document",
		},
		{
			name:   "action update_document",
			action: ActionDBDocumentUpdate,
			want:   "update_document",
		},
		{
			name:   "action create_index",
			action: ActionDBIndexCreate,
			want:   "create_index",
		},
		{
			name:   "action init_db",
			action: ActionDBInitialize,
			want:   "init_db",
		},
		{
			name:   "action execute_query",
			action: ActionDBQueryExecute,
			want:   "execute_query",
		},
		{
			name:   "action authorize_request",
			action: ActionAuthRequestAuthorize,
			want:   "authorize_request",
		},
		{
			name:   "action create_session",
			action: ActionAuthSessionCreate,
			want:   "create_session",
		},
		{
			name:   "action create_auth_store",
			action: ActionAuthStoreCreate,
			want:   "create_auth_store",
		},
		{
			name:   "action create_token",
			action: ActionAuthTokenCreate,
			want:   "create_token",
		},
		{
			name:   "action introspect_token",
			action: ActionAuthTokenIntrospect,
			want:   "introspect_token",
		},
		{
			name:   "action revoke_token",
			action: ActionAuthTokenRevoke,
			want:   "revoke_token",
		},
		{
			name:   "action authenticate_user",
			action: ActionAuthUserAuthenticate,
			want:   "authenticate_user",
		},
		{
			name:   "action deserialize_request",
			action: ActionRequestDeserialize,
			want:   "deserialize_request",
		},
		{
			name:   "action transform_request",
			action: ActionRequestTransform,
			want:   "transform_request",
		},
		{
			name:   "action validate_requester",
			action: ActionRequesterValidate,
			want:   "validate_requester",
		},
		{
			name:   "action load_certificate",
			action: ActionCertificateLoad,
			want:   "load_certificate",
		},
		{
			name:   "action load_config",
			action: ActionConfigLoad,
			want:   "load_config",
		},
		{
			name:   "action load_private_key",
			action: ActionPrivateKeyLoad,
			want:   "load_private_key",
		},
		{
			name:   "action health_check",
			action: ActionHealthCheck,
			want:   "health_check",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.action.String())
		})
	}
}
