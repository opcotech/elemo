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
			expected: "UserTokenContext(99)",
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
			want:    []byte("UserTokenContext(99)"),
			wantErr: nil,
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
		wantErr bool
	}{
		{
			name:    "unmarshal confirm context",
			text:    []byte("confirm"),
			want:    UserTokenContextConfirm,
			wantErr: false,
		},
		{
			name:    "unmarshal reset password context",
			text:    []byte("reset_password"),
			want:    UserTokenContextResetPassword,
			wantErr: false,
		},
		{
			name:    "unmarshal invalid context",
			text:    []byte("invalid"),
			want:    UserTokenContext(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var context UserTokenContext
			err := context.UnmarshalText(tt.text)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
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
		wantErr bool
	}{
		{
			name:    "scan confirm context",
			value:   "confirm",
			want:    UserTokenContextConfirm,
			wantErr: false,
		},
		{
			name:    "scan reset password context",
			value:   "reset_password",
			want:    UserTokenContextResetPassword,
			wantErr: false,
		},
		{
			name:    "scan invalid context",
			value:   "invalid",
			want:    UserTokenContext(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var context UserTokenContext
			err := context.Scan(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
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
			want:    "UserTokenContext(99)",
			wantErr: nil,
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
