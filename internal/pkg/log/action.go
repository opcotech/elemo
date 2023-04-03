package log

// Action represents the action that is being performed upon logging.
type Action int

const (
	// HTTP request actions

	ActionHTTPRequestHandle Action = iota // handle an HTTP request

	// GraphDatabase operation actions

	ActionDBCollectionCreate // create a database collection
	ActionDBDocumentCreate   // create a document in a collection
	ActionDBDocumentDelete   // delete a document from a collection
	ActionDBDocumentRead     // read a document from a collection
	ActionDBDocumentUpdate   // update a document in a collection
	ActionDBIndexCreate      // create an index in the collection
	ActionDBInitialize       // initialize a database
	ActionDBQueryExecute     // execute a database query

	// Authentication and authorization actions

	ActionAuthRequestAuthorize // authorize a request
	ActionAuthSessionCreate    // create a session
	ActionAuthStoreCreate      // create auth store
	ActionAuthTokenCreate      // create a token
	ActionAuthTokenIntrospect  // introspect a token
	ActionAuthTokenRevoke      // revoke a token
	ActionAuthUserAuthenticate // authenticate a user
	ActionRequestDeserialize   // deserialize a request
	ActionRequestTransform     // transform a request
	ActionRequesterValidate    // validate a requester

	// Configuration actions

	ActionCertificateLoad // load a certificate
	ActionConfigLoad      // load a configuration
	ActionPrivateKeyLoad  // load a private key

	// System actions

	ActionHealthCheck // check the health of a component
)

// String returns the string representation of the action.
func (a Action) String() string {
	return [...]string{
		"handle_http_request",
		"create_collection",
		"create_document",
		"delete_document",
		"read_document",
		"update_document",
		"create_index",
		"init_db",
		"execute_query",
		"authorize_request",
		"create_session",
		"create_auth_store",
		"create_token",
		"introspect_token",
		"revoke_token",
		"authenticate_user",
		"deserialize_request",
		"transform_request",
		"validate_requester",
		"load_certificate",
		"load_config",
		"load_private_key",
		"health_check",
	}[a]
}
