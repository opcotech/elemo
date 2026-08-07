package model

const (
	UserTokenContextConfirm       UserTokenContext = iota + 1 // confirm
	UserTokenContextResetPassword                             // reset_password
	UserTokenContextInvite                                    // invite
)

// UserTokenContext represents the reason of user token creation.
//
//go:generate go tool enumer -type=UserTokenContext -text -sql -transform=noop -linecomment -output=auth_user_token_context_gen.go
type UserTokenContext uint8
