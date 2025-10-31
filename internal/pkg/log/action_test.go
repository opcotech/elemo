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
			name:   "action put_file",
			action: ActionFilePut,
			want:   "put_file",
		},
		{
			name:   "action get_file",
			action: ActionFileGet,
			want:   "get_file",
		},
		{
			name:   "action update_file",
			action: ActionFileUpdate,
			want:   "update_file",
		},
		{
			name:   "action delete_file",
			action: ActionFileDelete,
			want:   "delete_file",
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
			name:   "action send_email",
			action: ActionEmailSend,
			want:   "send_email",
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

func TestAction_EventType(t *testing.T) {
	tests := []struct {
		name   string
		action Action
		want   string
	}{
		{
			name:   "event type http.request.served",
			action: ActionHTTPRequestHandle,
			want:   "http.request.served",
		},
		{
			name:   "event type collection.created",
			action: ActionDBCollectionCreate,
			want:   "collection.created",
		},
		{
			name:   "event type document.created",
			action: ActionDBDocumentCreate,
			want:   "document.created",
		},
		{
			name:   "event type document.deleted",
			action: ActionDBDocumentDelete,
			want:   "document.deleted",
		},
		{
			name:   "event type document.read",
			action: ActionDBDocumentRead,
			want:   "document.read",
		},
		{
			name:   "event type document.updated",
			action: ActionDBDocumentUpdate,
			want:   "document.updated",
		},
		{
			name:   "event type index.created",
			action: ActionDBIndexCreate,
			want:   "index.created",
		},
		{
			name:   "event type database.initialized",
			action: ActionDBInitialize,
			want:   "database.initialized",
		},
		{
			name:   "event type query.executed",
			action: ActionDBQueryExecute,
			want:   "query.executed",
		},
		{
			name:   "event type file.put",
			action: ActionFilePut,
			want:   "file.put",
		},
		{
			name:   "event type file.get",
			action: ActionFileGet,
			want:   "file.get",
		},
		{
			name:   "event type file.updated",
			action: ActionFileUpdate,
			want:   "file.updated",
		},
		{
			name:   "event type file.deleted",
			action: ActionFileDelete,
			want:   "file.deleted",
		},
		{
			name:   "event type request.authorized",
			action: ActionAuthRequestAuthorize,
			want:   "request.authorized",
		},
		{
			name:   "event type session.created",
			action: ActionAuthSessionCreate,
			want:   "session.created",
		},
		{
			name:   "event type auth_store.created",
			action: ActionAuthStoreCreate,
			want:   "auth_store.created",
		},
		{
			name:   "event type token.created",
			action: ActionAuthTokenCreate,
			want:   "token.created",
		},
		{
			name:   "event type token.introspected",
			action: ActionAuthTokenIntrospect,
			want:   "token.introspected",
		},
		{
			name:   "event type token.revoked",
			action: ActionAuthTokenRevoke,
			want:   "token.revoked",
		},
		{
			name:   "event type user.authenticated",
			action: ActionAuthUserAuthenticate,
			want:   "user.authenticated",
		},
		{
			name:   "event type request.deserialized",
			action: ActionRequestDeserialize,
			want:   "request.deserialized",
		},
		{
			name:   "event type request.transformed",
			action: ActionRequestTransform,
			want:   "request.transformed",
		},
		{
			name:   "event type requester.validated",
			action: ActionRequesterValidate,
			want:   "requester.validated",
		},
		{
			name:   "event type email.sent",
			action: ActionEmailSend,
			want:   "email.sent",
		},
		{
			name:   "event type certificate.loaded",
			action: ActionCertificateLoad,
			want:   "certificate.loaded",
		},
		{
			name:   "event type config.loaded",
			action: ActionConfigLoad,
			want:   "config.loaded",
		},
		{
			name:   "event type private_key.loaded",
			action: ActionPrivateKeyLoad,
			want:   "private_key.loaded",
		},
		{
			name:   "event type health.checked",
			action: ActionHealthCheck,
			want:   "health.checked",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.action.EventType())
		})
	}
}
