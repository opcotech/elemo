package log

import (
	"time"

	"go.uber.org/zap"
)

const (
	FieldAction                = "action"                  // name of the action field
	FieldAuthClient            = "auth_client"             // name of the auth client field
	FieldAuthClientID          = "auth_client_id"          // name of the auth client id field
	FieldAuthCode              = "auth_code"               // name of the code field
	FieldBindVars              = "bind_vars"               // name of the query bind vars field
	FieldCollection            = "collection"              // name of the collection field
	FieldCollectionOptions     = "collection_options"      // name of the collection field
	FieldDatabase              = "database"                // name of the database field
	Field                      = "details"                 // name of the details field
	FieldDocument              = "document"                // name of the document field
	FieldDocumentCount         = "document_count"          // name of the document count field
	FieldDuration              = "duration"                // name of the duration field
	FieldEmail                 = "email"                   // name of the email field
	FieldEndpoints             = "endpoints"               // name of the endpoints field
	FieldFilter                = "filter"                  // name of the filter field
	FieldIdleConnectionTimeout = "idle_connection_timeout" // name of the idle connection timeout field
	FieldIndexFields           = "fields"                  // name of the index fields
	FieldIndexOptions          = "index_options"           // name of the index options field
	FieldInput                 = "input"                   // name of the input field
	FieldKey                   = "key"                     // name of the key field
	FieldKind                  = "kind"                    // name of the kind field
	FieldLimit                 = "limit"                   // name of the limit field
	FieldMaxIdleConnections    = "max_idle_connections"    // name of the max idle connections field
	FieldMaxOpenConnections    = "max_open_connections"    // name of the max open connections field
	FieldMethod                = "method"                  // name of the method field
	FieldOffset                = "offset"                  // name of the offset field
	FieldOperationID           = "operation_id"            // name of the operation id field
	FieldPath                  = "path"                    // name of the path field
	FieldProtocol              = "protocol"                // name of the protocol field
	FieldQuery                 = "query"                   // name of the query field
	FieldRemoteAddr            = "remote_addr"             // name of the remote address field
	FieldRequestID             = "request_id"              // name of the request id field
	FieldRoles                 = "roles"                   // name of the roles field
	FieldScopes                = "scopes"                  // name of the scopes field
	FieldSession               = "session"                 // name of the session field
	FieldSize                  = "size"                    // name of the size field
	FieldStatus                = "status"                  // name of the user status field
	FieldTTL                   = "ttl"                     // name of the ttl field
	FieldToken                 = "token"                   // name of the token field
	FieldURL                   = "url"                     // name of the url field
	FieldUser                  = "user"                    // name of the user field
	FieldUserAgent             = "user_agent"              // name of the user agent field
	FieldUserID                = "user_id"                 // name of the user id field
	FieldUsername              = "username"                // name of the username field
	FieldValue                 = "value"                   // name of the value field
)

// WithAction sets the action field.
func WithAction(action Action) zap.Field {
	return zap.String(FieldAction, action.String())
}

// WithAuthClient sets the auth client field.
func WithAuthClient(client any) zap.Field {
	return zap.Any(FieldAuthClient, client)
}

// WithAuthClientID sets the auth client id field.
func WithAuthClientID(clientID string) zap.Field {
	return zap.String(FieldAuthClientID, clientID)
}

// WithAuthCode sets the authorization code field.
func WithAuthCode(code string) zap.Field {
	return zap.String(FieldAuthCode, code)
}

// WithBindVars sets the query bind vars field.
func WithBindVars(bindVars map[string]any) zap.Field {
	return zap.Any(FieldBindVars, bindVars)
}

// WithCollectionOptions sets the index options field.
func WithCollectionOptions(collectionOpts any) zap.Field {
	return zap.Any(FieldCollectionOptions, collectionOpts)
}

// WithDatabase sets the database field.
func WithDatabase(database string) zap.Field {
	return zap.String(FieldDatabase, database)
}

// With sets the details field.
func With(details string) zap.Field {
	return zap.String(Field, details)
}

// WithDocument sets the document field.
func WithDocument(document any) zap.Field {
	return zap.Any(FieldDocument, document)
}

// WithDocumentCount sets the document count field.
func WithDocumentCount(count int64) zap.Field {
	return zap.Int64(FieldDocumentCount, count)
}

// WithDuration sets the duration field.
func WithDuration[D time.Duration | float64 | int64](duration D) zap.Field {
	return zap.Float64(FieldDuration, float64(duration))
}

// WithEmail sets the email field.
func WithEmail(email string) zap.Field {
	return zap.String(FieldEmail, email)
}

// WithEndpoints sets the endpoints field.
func WithEndpoints(endpoints []string) zap.Field {
	return zap.Strings(FieldEndpoints, endpoints)
}

// WithError sets the error field.
func WithError(err error) zap.Field {
	return zap.Error(err)
}

// WithFilter sets the filter field.
func WithFilter(filter any) zap.Field {
	return zap.Any(FieldFilter, filter)
}

// WithIdleConnectionTimeout sets the idle connection timeout field.
func WithIdleConnectionTimeout(idleTimeout time.Duration) zap.Field {
	return zap.Duration(FieldIdleConnectionTimeout, idleTimeout)
}

// WithIndexFields sets the index fields.
func WithIndexFields(fields []string) zap.Field {
	return zap.Strings(FieldIndexFields, fields)
}

// WithIndexOptions sets the index options field.
func WithIndexOptions(indexOptions any) zap.Field {
	return zap.Any(FieldIndexOptions, indexOptions)
}

// WithInput sets the input field.
func WithInput(input any) zap.Field {
	return zap.Any(FieldInput, input)
}

// WithKey sets the key field.
func WithKey(key string) zap.Field {
	return zap.String(FieldKey, key)
}

// WithKind sets the kind field.
func WithKind(kind string) zap.Field {
	return zap.String(FieldKind, kind)
}

// WithLimit sets the limit field.
func WithLimit(limit int) zap.Field {
	return zap.Int(FieldLimit, limit)
}

// WithMaxIdleConnections sets the max idle connections field.
func WithMaxIdleConnections(maxIdleConnections int) zap.Field {
	return zap.Int(FieldMaxIdleConnections, maxIdleConnections)
}

// WithMaxOpenConnections sets the max open connections field.
func WithMaxOpenConnections(maxOpenConnections int) zap.Field {
	return zap.Int(FieldMaxOpenConnections, maxOpenConnections)
}

// WithMethod sets the method field.
func WithMethod(method string) zap.Field {
	return zap.String(FieldMethod, method)
}

// WithOffset sets the offset field.
func WithOffset(offset int) zap.Field {
	return zap.Int(FieldOffset, offset)
}

// WithOperationID sets the operation id field.
func WithOperationID(operationID string) zap.Field {
	return zap.String(FieldOperationID, operationID)
}

// WithPath sets the path field.
func WithPath(path string) zap.Field {
	return zap.String(FieldPath, path)
}

// WithProtocol sets the protocol field.
func WithProtocol(protocol string) zap.Field {
	return zap.String(FieldProtocol, protocol)
}

// WithQuery sets the query field.
func WithQuery(query string) zap.Field {
	return zap.String(FieldQuery, query)
}

// WithRemoteAddr sets the remote address field.
func WithRemoteAddr(remoteAddr string) zap.Field {
	return zap.String(FieldRemoteAddr, remoteAddr)
}

// WithRequestID sets the request id field.
func WithRequestID(requestID string) zap.Field {
	return zap.String(FieldRequestID, requestID)
}

// WithScopes sets the scopes field.
func WithScopes(scopes []string) zap.Field {
	return zap.Strings(FieldScopes, scopes)
}

// WithSession sets the session field.
func WithSession(session any) zap.Field {
	return zap.Any(FieldSession, session)
}

// WithSize sets the size field.
func WithSize(size int64) zap.Field {
	return zap.Int64(FieldSize, size)
}

// WithStatus sets the status code field.
func WithStatus[S string | int](status S) zap.Field {
	return zap.Any(FieldStatus, status)
}

// WithTTL sets the ttl field.
func WithTTL(ttl time.Duration) zap.Field {
	return zap.Duration(FieldTTL, ttl)
}

// WithToken sets the ttl field.
func WithToken(token string) zap.Field {
	return zap.String(FieldToken, token)
}

// WithURL sets the url field.
func WithURL(url string) zap.Field {
	return zap.String(FieldURL, url)
}

// WithUserAgent sets the user agent field.
func WithUserAgent(userAgent string) zap.Field {
	return zap.String(FieldUserAgent, userAgent)
}

// WithUserID sets the user id field.
func WithUserID(userID string) zap.Field {
	return zap.String(FieldUserID, userID)
}

// WithUsername sets the username field.
func WithUsername(username string) zap.Field {
	return zap.String(FieldUsername, username)
}

// WithValue sets the value field.
func WithValue(value any) zap.Field {
	return zap.Any(FieldValue, value)
}
