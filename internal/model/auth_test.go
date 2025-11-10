package model

import (
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createValidToken creates a 60-character token for testing
func createValidToken() string {
	return strings.Repeat("a", 60)
}

// createLongToken creates a 73-character token for testing (exceeds max of 72)
func createLongToken() string {
	return strings.Repeat("a", 73)
}

func TestUserTokenContext_String(t *testing.T) {
	tests := []struct {
		name     string
		context  UserTokenContext
		expected string
	}{
		{
			name:     "confirm context",
			context:  UserTokenContextConfirm,
			expected: "confirm",
		},
		{
			name:     "reset password context",
			context:  UserTokenContextResetPassword,
			expected: "reset_password",
		},
		{
			name:     "invalid context",
			context:  UserTokenContext(99),
			expected: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.context.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserTokenContext_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		context UserTokenContext
		want    []byte
		wantErr error
	}{
		{
			name:    "marshal confirm context",
			context: UserTokenContextConfirm,
			want:    []byte("confirm"),
			wantErr: nil,
		},
		{
			name:    "marshal reset password context",
			context: UserTokenContextResetPassword,
			want:    []byte("reset_password"),
			wantErr: nil,
		},
		{
			name:    "marshal invite context",
			context: UserTokenContextInvite,
			want:    []byte("invite"),
			wantErr: nil,
		},
		{
			name:    "marshal invalid context",
			context: UserTokenContext(99),
			want:    nil,
			wantErr: ErrInvalidUserTokenContext,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.context.MarshalText()
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserTokenContext_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		text    []byte
		want    UserTokenContext
		wantErr error
	}{
		{
			name:    "unmarshal confirm context",
			text:    []byte("confirm"),
			want:    UserTokenContextConfirm,
			wantErr: nil,
		},
		{
			name:    "unmarshal reset password context",
			text:    []byte("reset_password"),
			want:    UserTokenContextResetPassword,
			wantErr: nil,
		},
		{
			name:    "unmarshal invalid context",
			text:    []byte("invalid"),
			want:    UserTokenContext(0),
			wantErr: ErrInvalidUserTokenContext,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var context UserTokenContext
			err := context.UnmarshalText(tt.text)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want, context)
			}
		})
	}
}

func TestUserTokenContext_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    UserTokenContext
		wantErr error
	}{
		{
			name:    "scan confirm context",
			value:   "confirm",
			want:    UserTokenContextConfirm,
			wantErr: nil,
		},
		{
			name:    "scan reset password context",
			value:   "reset_password",
			want:    UserTokenContextResetPassword,
			wantErr: nil,
		},
		{
			name:    "scan invalid context",
			value:   "invalid",
			want:    UserTokenContext(0),
			wantErr: ErrInvalidUserTokenContext,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var context UserTokenContext
			err := context.Scan(tt.value)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want, context)
			}
		})
	}
}

func TestUserTokenContext_Value(t *testing.T) {
	tests := []struct {
		name    string
		context UserTokenContext
		want    driver.Value
		wantErr error
	}{
		{
			name:    "value confirm context",
			context: UserTokenContextConfirm,
			want:    "confirm",
			wantErr: nil,
		},
		{
			name:    "value reset password context",
			context: UserTokenContextResetPassword,
			want:    "reset_password",
			wantErr: nil,
		},
		{
			name:    "value invalid context",
			context: UserTokenContext(99),
			want:    "",
			wantErr: ErrInvalidUserTokenContext,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.context.Value()
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewUserToken(t *testing.T) {
	type args struct {
		userID  ID
		sentTo  string
		token   string
		context UserTokenContext
	}
	tests := []struct {
		name    string
		args    args
		want    *UserToken
		wantErr error
	}{
		{
			name: "create new user token",
			args: args{
				userID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				sentTo:  "test@example.com",
				token:   createValidToken(),
				context: UserTokenContextConfirm,
			},
			want: &UserToken{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:  "test@example.com",
				Token:   createValidToken(),
				Context: UserTokenContextConfirm,
			},
		},
		{
			name: "create new user token with reset password context",
			args: args{
				userID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				sentTo:  "reset@example.com",
				token:   createValidToken(),
				context: UserTokenContextResetPassword,
			},
			want: &UserToken{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:  "reset@example.com",
				Token:   createValidToken(),
				Context: UserTokenContextResetPassword,
			},
		},
		{
			name: "create new user token with invalid email",
			args: args{
				userID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				sentTo:  "invalid-email",
				token:   createValidToken(),
				context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "create new user token with short token",
			args: args{
				userID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				sentTo:  "test@example.com",
				token:   "short",
				context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "create new user token with long token",
			args: args{
				userID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				sentTo:  "test@example.com",
				token:   createLongToken(),
				context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "create new user token with empty sentTo",
			args: args{
				userID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				sentTo:  "",
				token:   createValidToken(),
				context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "create new user token with empty token",
			args: args{
				userID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				sentTo:  "test@example.com",
				token:   "",
				context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewUserToken(tt.args.userID, tt.args.sentTo, tt.args.token, tt.args.context)
			require.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUserToken_Validate(t *testing.T) {
	now := time.Now()
	type fields struct {
		ID        ID
		UserID    ID
		SentTo    string
		Token     string
		Context   UserTokenContext
		CreatedAt *time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid user token",
			fields: fields{
				ID:        ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:    ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:    "test@example.com",
				Token:     createValidToken(),
				Context:   UserTokenContextConfirm,
				CreatedAt: &now,
			},
		},
		{
			name: "valid user token without created at",
			fields: fields{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:  "test@example.com",
				Token:   createValidToken(),
				Context: UserTokenContextResetPassword,
			},
		},
		{
			name: "invalid user token ID",
			fields: fields{
				ID:      ID{Inner: xid.NilID(), Type: ResourceType(0)},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:  "test@example.com",
				Token:   createValidToken(),
				Context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "invalid user ID",
			fields: fields{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceType(0)},
				SentTo:  "test@example.com",
				Token:   createValidToken(),
				Context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "invalid email",
			fields: fields{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:  "invalid-email",
				Token:   createValidToken(),
				Context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "empty email",
			fields: fields{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:  "",
				Token:   createValidToken(),
				Context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "short token",
			fields: fields{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:  "test@example.com",
				Token:   "short",
				Context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "long token",
			fields: fields{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:  "test@example.com",
				Token:   createLongToken(),
				Context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "empty token",
			fields: fields{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:  "test@example.com",
				Token:   "",
				Context: UserTokenContextConfirm,
			},
			wantErr: ErrInvalidUserToken,
		},
		{
			name: "invalid context",
			fields: fields{
				ID:      ID{Inner: xid.NilID(), Type: ResourceTypeUserToken},
				UserID:  ID{Inner: xid.NilID(), Type: ResourceTypeUser},
				SentTo:  "test@example.com",
				Token:   createValidToken(),
				Context: UserTokenContext(99),
			},
			wantErr: ErrInvalidUserToken,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ut := &UserToken{
				ID:        tt.fields.ID,
				UserID:    tt.fields.UserID,
				SentTo:    tt.fields.SentTo,
				Token:     tt.fields.Token,
				Context:   tt.fields.Context,
				CreatedAt: tt.fields.CreatedAt,
			}
			require.ErrorIs(t, ut.Validate(), tt.wantErr)
		})
	}
}
