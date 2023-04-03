package log

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWithAction(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithAction(ActionHealthCheck), zap.String(FieldAction, ActionHealthCheck.String()))
}

func TestWithAuthClient(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithAuthClient("client"), zap.Any(FieldAuthClient, "client"))
}

func TestWithAuthClientID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithAuthClientID("clientID"), zap.String(FieldAuthClientID, "clientID"))
}

func TestWithAuthCode(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithAuthCode("code"), zap.String(FieldAuthCode, "code"))
}

func TestWithBindVars(t *testing.T) {
	t.Parallel()

	vars := map[string]any{"a": 1}
	assert.Equal(t, WithBindVars(vars), zap.Any(FieldBindVars, vars))
}

func TestWithCollectionOptions(t *testing.T) {
	t.Parallel()

	opts := map[string]any{"a": 1}
	assert.Equal(t, WithCollectionOptions(opts), zap.Any(FieldCollectionOptions, opts))
}

func TestWithDatabase(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithDatabase("database"), zap.String(FieldDatabase, "database"))
}

func TestWith(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithDetails("details"), zap.String(Field, "details"))
}

func TestWithDocument(t *testing.T) {
	t.Parallel()

	doc := map[string]any{"a": 1}
	assert.Equal(t, WithDocument(doc), zap.Any(FieldDocument, doc))
}

func TestWithDocumentCount(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithDocumentCount(10), zap.Int(FieldDocumentCount, 10))
}

func TestWithDuration(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithDuration(10*time.Second), zap.Float64(FieldDuration, 10000000000))
}

func TestWithEmail(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithEmail("test@example.com"), zap.String(FieldEmail, "test@example.com"))
}

func TestWithEndpoints(t *testing.T) {
	t.Parallel()

	endpoints := []string{"http://localhost:2379"}
	assert.Equal(t, WithEndpoints(endpoints), zap.Strings(FieldEndpoints, endpoints))
}

func TestWithError(t *testing.T) {
	t.Parallel()

	err := errors.New("error")
	assert.Equal(t, WithError(err), zap.Error(err))
}

func TestWithFilter(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithFilter("filter"), zap.String(FieldFilter, "filter"))
}

func TestWithIdleConnectionTimeout(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithIdleConnectionTimeout(10*time.Second), zap.Duration(FieldIdleConnectionTimeout, 10*time.Second))
}

func TestWithIndexOptions(t *testing.T) {
	t.Parallel()

	opts := map[string]any{"a": 1}
	assert.Equal(t, WithIndexOptions(opts), zap.Any(FieldIndexOptions, opts))
}

func TestWithInput(t *testing.T) {
	t.Parallel()

	input := map[string]any{"a": 1}
	assert.Equal(t, WithInput(input), zap.Any(FieldInput, input))
}

func TestWithIndexFields(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithIndexFields([]string{"field"}), zap.Strings(FieldIndexFields, []string{"field"}))
}

func TestWithKey(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithKey("key"), zap.String(FieldKey, "key"))
}

func TestWithKind(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithKind("kind"), zap.String(FieldKind, "kind"))
}

func TestWithLimit(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithLimit(10), zap.Int(FieldLimit, 10))
}

func TestWithMaxIdleConnections(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithMaxIdleConnections(10), zap.Int(FieldMaxIdleConnections, 10))
}

func TestWithMaxOpenConnections(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithMaxOpenConnections(10), zap.Int(FieldMaxOpenConnections, 10))
}

func TestWithMethod(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithMethod("method"), zap.String(FieldMethod, "method"))
}

func TestWithOffset(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithOffset(10), zap.Int(FieldOffset, 10))
}

func TestWithOperationID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithOperationID("operationID"), zap.String(FieldOperationID, "operationID"))
}

func TestWithPath(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithPath("path"), zap.String(FieldPath, "path"))
}

func TestWithProtocol(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithProtocol("protocol"), zap.String(FieldProtocol, "protocol"))
}

func TestWithQuery(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithQuery("query"), zap.String(FieldQuery, "query"))
}

func TestWithRemoteAddr(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithRemoteAddr("localhost:2379"), zap.String(FieldRemoteAddr, "localhost:2379"))
}

func TestWithRequestID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithRequestID("requestID"), zap.String(FieldRequestID, "requestID"))
}

func TestWithScopes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithScopes([]string{"scope"}), zap.Strings(FieldScopes, []string{"scope"}))
}

func TestWithSession(t *testing.T) {
	t.Parallel()

	session := map[string]any{"a": 1}
	assert.Equal(t, WithSession(session), zap.Any(FieldSession, session))
}

func TestWithSize(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithSize(10), zap.Int(FieldSize, 10))
}

func TestWithStatus(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithStatus(200), zap.Int(FieldStatus, 200))
}

func TestWithTTL(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithTTL(10*time.Second), zap.Duration(FieldTTL, 10*time.Second))
}

func TestWithToken(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithToken("token"), zap.String(FieldToken, "token"))
}

func TestWithURL(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithURL("https://example.com"), zap.String(FieldURL, "https://example.com"))
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithUserAgent("userAgent"), zap.String(FieldUserAgent, "userAgent"))
}

func TestWithUserID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithUserID("userID"), zap.String(FieldUserID, "userID"))
}

func TestWithUserStatus(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithStatus("userStatus"), zap.String(FieldStatus, "userStatus"))
}

func TestWithUsername(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithUsername("username"), zap.String(FieldUsername, "username"))
}

func TestWithValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithValue("value"), zap.String(FieldValue, "value"))
}
