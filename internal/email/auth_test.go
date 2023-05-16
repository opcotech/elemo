package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPasswordResetTemplateData(t *testing.T) {
	type args struct {
		username         string
		firstName        string
		passwordResetURL string
		supportEmail     string
	}
	tests := []struct {
		name    string
		args    args
		want    *PasswordResetTemplateData
		wantErr error
	}{
		{
			name: "valid template data",
			args: args{
				username:         "test-user",
				firstName:        "Test",
				passwordResetURL: "https://example.com/password/reset",
				supportEmail:     "info@example.com",
			},
			want: &PasswordResetTemplateData{
				Username:         "test-user",
				FirstName:        "Test",
				PasswordResetURL: "https://example.com/password/reset",
				SupportEmail:     "info@example.com",
			},
		},
		{
			name:    "invalid template data",
			args:    args{},
			wantErr: ErrInvalidPasswordResetTemplateData,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewPasswordResetTemplateData(tt.args.username, tt.args.firstName, tt.args.passwordResetURL, tt.args.supportEmail)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPasswordResetTemplateData_Validate(t *testing.T) {
	type args struct {
		username         string
		firstName        string
		passwordResetURL string
		supportEmail     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "valid email template",
			args: args{
				username:         "test-user",
				firstName:        "Test",
				passwordResetURL: "https://example.com/password/reset",
				supportEmail:     "info@example.com",
			},
		},
		{
			name: "invalid username",
			args: args{
				username:         "",
				firstName:        "Test",
				passwordResetURL: "https://example.com/password/reset",
				supportEmail:     "info@example.com",
			},
			wantErr: ErrInvalidPasswordResetTemplateData,
		},
		{
			name: "invalid password reset url",
			args: args{
				username:         "test-user",
				firstName:        "Test",
				passwordResetURL: "",
				supportEmail:     "info@example.com",
			},
			wantErr: ErrInvalidPasswordResetTemplateData,
		},
		{
			name: "invalid support email",
			args: args{
				username:         "test-user",
				firstName:        "Test",
				passwordResetURL: "https://example.com/password/reset",
				supportEmail:     "info@example",
			},
			wantErr: ErrInvalidPasswordResetTemplateData,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := &PasswordResetTemplateData{
				Username:         tt.args.username,
				FirstName:        tt.args.firstName,
				PasswordResetURL: tt.args.passwordResetURL,
				SupportEmail:     tt.args.supportEmail,
			}

			assert.ErrorIs(t, data.Validate(), tt.wantErr)
		})
	}
}

func TestPasswordResetTemplateData_Get(t *testing.T) {
	t.Parallel()

	data := &PasswordResetTemplateData{
		Username:         "test-user",
		FirstName:        "Test",
		PasswordResetURL: "https://example.com/password/reset",
		SupportEmail:     "info@example.com",
	}

	assert.Equal(t, data, data.Get())
}
