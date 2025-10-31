package log

import (
	"log/slog"
	"time"

	"github.com/rs/xid"
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
	FieldErrorCode             = "error_code"              // name of the error code field
	FieldEventID               = "event_id"                // name of the event id field
	FieldEventType             = "event_type"              // name of the event type field
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
	FieldMetadata              = "metadata"                // name of the metadata field
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
	FieldSessionID             = "session_id"              // name of the session id field
	FieldSize                  = "size"                    // name of the size field
	FieldStatus                = "status"                  // name of the user status field
	FieldSubject               = "subject"                 // name of the subject field
	FieldTTL                   = "ttl"                     // name of the ttl field
	FieldToken                 = "token"                   // name of the token field
	FieldTraceID               = "trace_id"                // name of the trace id field
	FieldURL                   = "url"                     // name of the url field
	FieldUser                  = "user"                    // name of the user field
	FieldUserAgent             = "user_agent"              // name of the user agent field
	FieldUserID                = "user_id"                 // name of the user id field
	FieldUsername              = "username"                // name of the username field
	FieldValue                 = "value"                   // name of the value field
)

// WithAction sets the action field.
func WithAction(action Action) Attr {
	return slog.String(FieldAction, action.String())
}

// WithAuthClient sets the auth client field.
func WithAuthClient(client any) Attr {
	return slog.Any(FieldAuthClient, client)
}

// WithAuthClientID sets the auth client id field.
func WithAuthClientID(clientID string) Attr {
	return slog.String(FieldAuthClientID, clientID)
}

// WithAuthCode sets the authorization code field.
func WithAuthCode(code string) Attr {
	return slog.String(FieldAuthCode, code)
}

// WithBindVars sets the query bind vars field.
func WithBindVars(bindVars map[string]any) Attr {
	return slog.Any(FieldBindVars, bindVars)
}

// WithCollectionOptions sets the index options field.
func WithCollectionOptions(collectionOpts any) Attr {
	return slog.Any(FieldCollectionOptions, collectionOpts)
}

// WithDatabase sets the database field.
func WithDatabase(database string) Attr {
	return slog.String(FieldDatabase, database)
}

// WithDetails sets the details field.
func WithDetails(details string) Attr {
	return slog.String(Field, details)
}

// WithDocument sets the document field.
func WithDocument(document any) Attr {
	return slog.Any(FieldDocument, document)
}

// WithDocumentCount sets the document count field.
func WithDocumentCount(count int64) Attr {
	return slog.Int64(FieldDocumentCount, count)
}

// WithDuration sets the duration field.
func WithDuration[D time.Duration | float64 | int64](duration D) Attr {
	return slog.Float64(FieldDuration, float64(duration))
}

// WithEmail sets the email field.
func WithEmail(email string) Attr {
	return slog.String(FieldEmail, email)
}

// WithEndpoints sets the endpoints field.
func WithEndpoints(endpoints []string) Attr {
	return slog.Any(FieldEndpoints, endpoints)
}

// WithError sets the error field.
func WithError(err error) Attr {
	return slog.Any("error", err)
}

// WithErrorCode sets the error code field.
func WithErrorCode(code string) Attr {
	return slog.String(FieldErrorCode, code)
}

// WithEventID sets the event id field.
func WithEventID(eventID string) Attr {
	return slog.String(FieldEventID, eventID)
}

// WithEventIDAuto generates and sets an event id field using XID.
func WithEventIDAuto() Attr {
	return slog.String(FieldEventID, xid.New().String())
}

// WithEventType sets the event type field (noun.verb format).
func WithEventType(eventType string) Attr {
	return slog.String(FieldEventType, eventType)
}

// WithSessionID sets the session id field.
func WithSessionID(sessionID string) Attr {
	return slog.String(FieldSessionID, sessionID)
}

// WithMetadata sets the metadata field (structured key-value pairs).
func WithMetadata(metadata map[string]any) Attr {
	return slog.Any(FieldMetadata, metadata)
}

// WithFilter sets the filter field.
func WithFilter(filter any) Attr {
	return slog.Any(FieldFilter, filter)
}

// WithIdleConnectionTimeout sets the idle connection timeout field.
func WithIdleConnectionTimeout(idleTimeout time.Duration) Attr {
	return slog.Duration(FieldIdleConnectionTimeout, idleTimeout)
}

// WithIndexFields sets the index fields.
func WithIndexFields(fields []string) Attr {
	return slog.Any(FieldIndexFields, fields)
}

// WithIndexOptions sets the index options field.
func WithIndexOptions(indexOptions any) Attr {
	return slog.Any(FieldIndexOptions, indexOptions)
}

// WithInput sets the input field.
func WithInput(input any) Attr {
	return slog.Any(FieldInput, input)
}

// WithKey sets the key field.
func WithKey(key string) Attr {
	return slog.String(FieldKey, key)
}

// WithKind sets the kind field.
func WithKind(kind string) Attr {
	return slog.String(FieldKind, kind)
}

// WithLimit sets the limit field.
func WithLimit(limit int) Attr {
	return slog.Int(FieldLimit, limit)
}

// WithMaxIdleConnections sets the max idle connections field.
func WithMaxIdleConnections(maxIdleConnections int) Attr {
	return slog.Int(FieldMaxIdleConnections, maxIdleConnections)
}

// WithMaxOpenConnections sets the max open connections field.
func WithMaxOpenConnections(maxOpenConnections int) Attr {
	return slog.Int(FieldMaxOpenConnections, maxOpenConnections)
}

// WithMethod sets the method field.
func WithMethod(method string) Attr {
	return slog.String(FieldMethod, method)
}

// WithOffset sets the offset field.
func WithOffset(offset int) Attr {
	return slog.Int(FieldOffset, offset)
}

// WithOperationID sets the operation id field.
func WithOperationID(operationID string) Attr {
	return slog.String(FieldOperationID, operationID)
}

// WithPath sets the path field.
func WithPath(path string) Attr {
	return slog.String(FieldPath, path)
}

// WithProtocol sets the protocol field.
func WithProtocol(protocol string) Attr {
	return slog.String(FieldProtocol, protocol)
}

// WithQuery sets the query field.
func WithQuery(query string) Attr {
	return slog.String(FieldQuery, query)
}

// WithRemoteAddr sets the remote address field.
func WithRemoteAddr(remoteAddr string) Attr {
	return slog.String(FieldRemoteAddr, remoteAddr)
}

// WithRequestID sets the request id field.
func WithRequestID(requestID string) Attr {
	return slog.String(FieldRequestID, requestID)
}

// WithScopes sets the scopes field.
func WithScopes(scopes []string) Attr {
	return slog.Any(FieldScopes, scopes)
}

// WithSession sets the session field.
func WithSession(session any) Attr {
	return slog.Any(FieldSession, session)
}

// WithSize sets the size field.
func WithSize(size int64) Attr {
	return slog.Int64(FieldSize, size)
}

// Status represents the status of an event.
type Status string

const (
	StatusSuccess  Status = "success"
	StatusFailure  Status = "failure"
	StatusPending  Status = "pending"
	StatusCanceled Status = "canceled"
)

// WithStatus sets the status field.
func WithStatus[S Status | string | int](status S) Attr {
	return slog.Any(FieldStatus, status)
}

// WithSubject sets the subject field.
func WithSubject(subject string) Attr {
	return slog.String(FieldSubject, subject)
}

// WithTTL sets the ttl field.
func WithTTL(ttl time.Duration) Attr {
	return slog.Duration(FieldTTL, ttl)
}

// WithToken sets the ttl field.
func WithToken(token string) Attr {
	return slog.String(FieldToken, token)
}

// WithTraceID sets the trace ID field.
func WithTraceID(id string) Attr {
	return slog.String(FieldTraceID, id)
}

// WithURL sets the url field.
func WithURL(url string) Attr {
	return slog.String(FieldURL, url)
}

// WithUserAgent sets the user agent field.
func WithUserAgent(userAgent string) Attr {
	return slog.String(FieldUserAgent, userAgent)
}

// WithUserID sets the user id field.
func WithUserID(userID string) Attr {
	return slog.String(FieldUserID, userID)
}

// WithUsername sets the username field.
func WithUsername(username string) Attr {
	return slog.String(FieldUsername, username)
}

// WithValue sets the value field.
func WithValue(value any) Attr {
	return slog.Any(FieldValue, value)
}

// WithContextObject sets a context object field for grouping related fields in complex events.
func WithContextObject(context map[string]any) Attr {
	return slog.Any("context", context)
}
