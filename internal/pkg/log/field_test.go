package log

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithAction(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithAction(ActionHealthCheck), slog.String(FieldAction, ActionHealthCheck.String()))
}

func TestWithAuthClient(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithAuthClient("client"), slog.Any(FieldAuthClient, "client"))
}

func TestWithAuthClientID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithAuthClientID("clientID"), slog.String(FieldAuthClientID, "clientID"))
}

func TestWithAuthCode(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithAuthCode("code"), slog.String(FieldAuthCode, "code"))
}

func TestWithBindVars(t *testing.T) {
	t.Parallel()

	vars := map[string]any{"a": 1}
	assert.Equal(t, WithBindVars(vars), slog.Any(FieldBindVars, vars))
}

func TestWithCollectionOptions(t *testing.T) {
	t.Parallel()

	opts := map[string]any{"a": 1}
	assert.Equal(t, WithCollectionOptions(opts), slog.Any(FieldCollectionOptions, opts))
}

func TestWithDatabase(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithDatabase("database"), slog.String(FieldDatabase, "database"))
}

func TestWithDetails(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithDetails("details"), slog.String(Field, "details"))
}

func TestWithDocument(t *testing.T) {
	t.Parallel()

	doc := map[string]any{"a": 1}
	assert.Equal(t, WithDocument(doc), slog.Any(FieldDocument, doc))
}

func TestWithDocumentCount(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithDocumentCount(10), slog.Int64(FieldDocumentCount, 10))
}

func TestWithDuration(t *testing.T) {
	t.Parallel()

	t.Run("time.Duration", func(t *testing.T) {
		t.Parallel()
		attr := WithDuration(10 * time.Second)
		assert.Equal(t, FieldDuration, attr.Key)
		assert.Equal(t, 10000000000.0, attr.Value.Any())
	})

	t.Run("float64", func(t *testing.T) {
		t.Parallel()
		attr := WithDuration(1.5)
		assert.Equal(t, FieldDuration, attr.Key)
		assert.Equal(t, 1.5, attr.Value.Any())
	})

	t.Run("int64", func(t *testing.T) {
		t.Parallel()
		attr := WithDuration(int64(1000))
		assert.Equal(t, FieldDuration, attr.Key)
		assert.Equal(t, 1000.0, attr.Value.Any())
	})
}

func TestWithEmail(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithEmail("test@example.com"), slog.String(FieldEmail, "test@example.com"))
}

func TestWithEndpoints(t *testing.T) {
	t.Parallel()

	endpoints := []string{"http://localhost:2379"}
	assert.Equal(t, WithEndpoints(endpoints), slog.Any(FieldEndpoints, endpoints))
}

func TestWithError(t *testing.T) {
	t.Parallel()

	err := assert.AnError
	assert.Equal(t, WithError(err), slog.Any("error", err))
}

func TestWithErrorCode(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithErrorCode("E001"), slog.String(FieldErrorCode, "E001"))
}

func TestWithEventID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithEventID("event123"), slog.String(FieldEventID, "event123"))
}

func TestWithEventIDAuto(t *testing.T) {
	t.Parallel()

	attr := WithEventIDAuto()
	assert.Equal(t, FieldEventID, attr.Key)
	assert.NotEmpty(t, attr.Value.String())

	// Generate another to ensure uniqueness
	attr2 := WithEventIDAuto()
	assert.NotEqual(t, attr.Value.String(), attr2.Value.String())
}

func TestWithEventType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithEventType("user.created"), slog.String(FieldEventType, "user.created"))
}

func TestWithFilter(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithFilter("filter"), slog.String(FieldFilter, "filter"))
}

func TestWithIdleConnectionTimeout(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithIdleConnectionTimeout(10*time.Second), slog.Duration(FieldIdleConnectionTimeout, 10*time.Second))
}

func TestWithIndexOptions(t *testing.T) {
	t.Parallel()

	opts := map[string]any{"a": 1}
	assert.Equal(t, WithIndexOptions(opts), slog.Any(FieldIndexOptions, opts))
}

func TestWithInput(t *testing.T) {
	t.Parallel()

	input := map[string]any{"a": 1}
	assert.Equal(t, WithInput(input), slog.Any(FieldInput, input))
}

func TestWithIndexFields(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithIndexFields([]string{"field"}), slog.Any(FieldIndexFields, []string{"field"}))
}

func TestWithKey(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithKey("key"), slog.String(FieldKey, "key"))
}

func TestWithKind(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithKind("kind"), slog.String(FieldKind, "kind"))
}

func TestWithLimit(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithLimit(10), slog.Int(FieldLimit, 10))
}

func TestWithMaxIdleConnections(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithMaxIdleConnections(10), slog.Int(FieldMaxIdleConnections, 10))
}

func TestWithMaxOpenConnections(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithMaxOpenConnections(10), slog.Int(FieldMaxOpenConnections, 10))
}

func TestWithMethod(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithMethod("method"), slog.String(FieldMethod, "method"))
}

func TestWithOffset(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithOffset(10), slog.Int(FieldOffset, 10))
}

func TestWithOperationID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithOperationID("operationID"), slog.String(FieldOperationID, "operationID"))
}

func TestWithPath(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithPath("path"), slog.String(FieldPath, "path"))
}

func TestWithProtocol(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithProtocol("protocol"), slog.String(FieldProtocol, "protocol"))
}

func TestWithQuery(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithQuery("query"), slog.String(FieldQuery, "query"))
}

func TestWithRemoteAddr(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithRemoteAddr("localhost:2379"), slog.String(FieldRemoteAddr, "localhost:2379"))
}

func TestWithRequestID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithRequestID("requestID"), slog.String(FieldRequestID, "requestID"))
}

func TestWithScopes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithScopes([]string{"scope"}), slog.Any(FieldScopes, []string{"scope"}))
}

func TestWithSession(t *testing.T) {
	t.Parallel()

	session := map[string]any{"a": 1}
	assert.Equal(t, WithSession(session), slog.Any(FieldSession, session))
}

func TestWithSessionID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithSessionID("session123"), slog.String(FieldSessionID, "session123"))
}

func TestWithMetadata(t *testing.T) {
	t.Parallel()

	metadata := map[string]any{"key": "value"}
	assert.Equal(t, WithMetadata(metadata), slog.Any(FieldMetadata, metadata))
}

func TestWithSize(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithSize(10), slog.Int64(FieldSize, 10))
}

func TestWithStatus(t *testing.T) {
	t.Parallel()

	t.Run("Status constant", func(t *testing.T) {
		t.Parallel()
		attr := WithStatus(StatusSuccess)
		assert.Equal(t, FieldStatus, attr.Key)
		assert.Equal(t, StatusSuccess, attr.Value.Any())
	})

	t.Run("string status", func(t *testing.T) {
		t.Parallel()
		attr := WithStatus("pending")
		assert.Equal(t, FieldStatus, attr.Key)
		assert.Equal(t, "pending", attr.Value.Any())
	})

	t.Run("int status", func(t *testing.T) {
		t.Parallel()
		attr := WithStatus(200)
		assert.Equal(t, FieldStatus, attr.Key)
		// slog converts int to int64
		assert.Equal(t, int64(200), attr.Value.Any())
	})
}

func TestWithStatusConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithStatus(StatusSuccess), slog.Any(FieldStatus, StatusSuccess))
	assert.Equal(t, WithStatus(StatusFailure), slog.Any(FieldStatus, StatusFailure))
	assert.Equal(t, WithStatus(StatusPending), slog.Any(FieldStatus, StatusPending))
	assert.Equal(t, WithStatus(StatusCanceled), slog.Any(FieldStatus, StatusCanceled))
}

func TestWithSubject(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithSubject("subject"), slog.String(FieldSubject, "subject"))
}

func TestWithTTL(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithTTL(10*time.Second), slog.Duration(FieldTTL, 10*time.Second))
}

func TestWithToken(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithToken("token"), slog.String(FieldToken, "token"))
}

func TestWithTraceID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithTraceID("12345678"), slog.String(FieldTraceID, "12345678"))
}

func TestWithURL(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithURL("https://example.com"), slog.String(FieldURL, "https://example.com"))
}

func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithUserAgent("userAgent"), slog.String(FieldUserAgent, "userAgent"))
}

func TestWithUserID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithUserID("userID"), slog.String(FieldUserID, "userID"))
}

func TestWithUsername(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithUsername("username"), slog.String(FieldUsername, "username"))
}

func TestWithValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, WithValue("value"), slog.Any(FieldValue, "value"))
}

func TestWithContextObject(t *testing.T) {
	t.Parallel()

	ctx := map[string]any{"user_id": "u123", "payment_id": "p456"}
	assert.Equal(t, WithContextObject(ctx), slog.Any("context", ctx))
}
