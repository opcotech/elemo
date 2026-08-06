package model

import (
	"testing"

	"database/sql/driver"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
