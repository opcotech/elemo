package pkg

const (
	CtxKeyUserID CtxKey = "userID" // ID of the user who made the request
	CtxKeyLogger CtxKey = "logger" // request-scoped logger
)

// CtxKey is the type alias for the context key.
type CtxKey string
