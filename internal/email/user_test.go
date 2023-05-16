package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserWelcomeTemplateData(t *testing.T) {
	type args struct {
		username     string
		firstName    string
		serverURL    string
		supportEmail string
	}
	tests := []struct {
		name    string
		args    args
		want    *UserWelcomeTemplateData
		wantErr error
	}{
		{
			name: "valid template data",
			args: args{
				username:     "test-user",
				firstName:    "Test",
				serverURL:    "https://example.com",
				supportEmail: "info@example.com",
			},
			want: &UserWelcomeTemplateData{
				Username:     "test-user",
				FirstName:    "Test",
				LoginURL:     "https://example.com",
				SupportEmail: "info@example.com",
			},
		},
		{
			name:    "invalid template data",
			args:    args{},
			wantErr: ErrInvalidUserWelcomeTemplateData,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewUserWelcomeTemplateData(tt.args.username, tt.args.firstName, tt.args.serverURL, tt.args.supportEmail)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserWelcomeTemplateData_Validate(t *testing.T) {
	type args struct {
		username     string
		firstName    string
		serverURL    string
		supportEmail string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "valid email template",
			args: args{
				username:     "test-user",
				firstName:    "Test",
				serverURL:    "https://example.com",
				supportEmail: "info@example.com",
			},
		},
		{
			name: "invalid username",
			args: args{
				username:     "",
				firstName:    "Test",
				serverURL:    "https://example.com",
				supportEmail: "info@example.com",
			},
			wantErr: ErrInvalidUserWelcomeTemplateData,
		},
		{
			name: "invalid server url",
			args: args{
				username:     "test-user",
				firstName:    "Test",
				serverURL:    "",
				supportEmail: "info@example.com",
			},
			wantErr: ErrInvalidUserWelcomeTemplateData,
		},
		{
			name: "invalid support email",
			args: args{
				username:     "test-user",
				firstName:    "Test",
				serverURL:    "https://example.com",
				supportEmail: "info@example",
			},
			wantErr: ErrInvalidUserWelcomeTemplateData,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := &UserWelcomeTemplateData{
				Username:     tt.args.username,
				FirstName:    tt.args.firstName,
				LoginURL:     tt.args.serverURL,
				SupportEmail: tt.args.supportEmail,
			}

			assert.ErrorIs(t, data.Validate(), tt.wantErr)
		})
	}
}

func TestUserWelcomeTemplateData_Get(t *testing.T) {
	t.Parallel()

	data := &UserWelcomeTemplateData{
		Username:     "test-user",
		FirstName:    "Test",
		LoginURL:     "https://example.com",
		SupportEmail: "info@example.com",
	}

	assert.Equal(t, data, data.Get())
}
