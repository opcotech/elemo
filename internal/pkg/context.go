package pkg

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

const (
	CtxKeyUserID        CtxKey = "userID"        // ID of the user who made the request
	CtxKeyOAuthClientID CtxKey = "oauthClientID" // OAuth client ID from the access token
	CtxKeyLogger        CtxKey = "logger"        // request-scoped logger
)

const (
	CtxMachineUser CtxMachineUserKind = "machine"
)

// CtxKey is the type alias for the context key.
type CtxKey string

// CtxMachineUserKind is the type alias for the machine user kind.
type CtxMachineUserKind string

// CtxUserID returns the context user ID as a string. If no ID found, an
// empty string returned.
func CtxUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(CtxKeyUserID).(CtxMachineUserKind); ok && userID == CtxMachineUser {
		return string(userID)
	}

	if userID, ok := CtxUserIDValue(ctx); ok {
		return userID.String()
	}

	return ""
}

// CtxUserIDValue returns the authenticated user ID stored on ctx.
func CtxUserIDValue(ctx context.Context) (model.ID, bool) {
	userID, ok := ctx.Value(CtxKeyUserID).(model.ID)
	return userID, ok
}

// CtxOAuthClientID returns the OAuth client ID stored on ctx.
func CtxOAuthClientID(ctx context.Context) (string, bool) {
	clientID, ok := ctx.Value(CtxKeyOAuthClientID).(string)
	if !ok || clientID == "" {
		return "", false
	}
	return clientID, true
}
