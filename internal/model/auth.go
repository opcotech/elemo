package model

const (
	UserTokenContextConfirm UserTokenContext = iota + 1
	UserTokenContextResetPassword
	UserTokenContextInvite
)

// UserTokenContext represents the reason of user token creation.
//
//go:generate go tool enumer -type=UserTokenContext -trimprefix=UserTokenContext -text -sql -transform=snake -output=auth_user_token_context_gen.go
type UserTokenContext uint8
