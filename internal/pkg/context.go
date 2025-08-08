package pkg

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

const (
	CtxKeyUserID CtxKey = "userID" // ID of the user who made the request
	CtxKeyLogger CtxKey = "logger" // request-scoped logger
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

	if user, ok := ctx.Value(CtxKeyUserID).(*model.User); ok {
		return user.ID.String()
	}

	return ""
}
