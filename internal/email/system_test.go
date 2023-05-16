package email

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"

	"github.com/opcotech/elemo/internal/license"
)

func TestNewLicenseExpiryTemplateData(t *testing.T) {
	licenseID := xid.New()

	type args struct {
		username     string
		firstName    string
		license      *license.License
		renewEmail   string
		supportEmail string
	}
	tests := []struct {
		name    string
		args    args
		want    *LicenseExpiryTemplateData
		wantErr error
	}{
		{
			name: "valid template data",
			args: args{
				username:     "test-user",
				firstName:    "Test",
				license:      &license.License{ID: licenseID},
				renewEmail:   "renew@example.com",
				supportEmail: "info@example.com",
			},
			want: &LicenseExpiryTemplateData{
				Username:     "test-user",
				FirstName:    "Test",
				License:      &license.License{ID: licenseID},
				RenewEmail:   "renew@example.com",
				SupportEmail: "info@example.com",
			},
		},
		{
			name:    "invalid template data",
			args:    args{},
			wantErr: ErrInvalidLicenseExpiryTemplateData,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewLicenseExpiryTemplateData(tt.args.username, tt.args.firstName, tt.args.renewEmail, tt.args.supportEmail, tt.args.license)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLicenseExpiryTemplateData_Validate(t *testing.T) {
	type args struct {
		username     string
		firstName    string
		license      *license.License
		renewEmail   string
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
				license:      &license.License{ID: xid.New()},
				renewEmail:   "renew@example.com",
				supportEmail: "info@example.com",
			},
		},
		{
			name: "invalid username",
			args: args{
				username:     "",
				firstName:    "Test",
				license:      &license.License{ID: xid.New()},
				renewEmail:   "renew@example.com",
				supportEmail: "info@example.com",
			},
			wantErr: ErrInvalidLicenseExpiryTemplateData,
		},
		{
			name: "invalid license id",
			args: args{
				username:     "test-user",
				firstName:    "Test",
				license:      &license.License{ID: xid.NilID()},
				renewEmail:   "renew@example.com",
				supportEmail: "info@example.com",
			},
			wantErr: ErrInvalidLicenseExpiryTemplateData,
		},
		{
			name: "invalid renew email",
			args: args{
				username:     "test-user",
				firstName:    "Test",
				license:      &license.License{ID: xid.New()},
				renewEmail:   "renew@example",
				supportEmail: "info@example.com",
			},
			wantErr: ErrInvalidLicenseExpiryTemplateData,
		},
		{
			name: "invalid support email",
			args: args{
				username:     "test-user",
				firstName:    "Test",
				license:      &license.License{ID: xid.New()},
				renewEmail:   "renew@example.com",
				supportEmail: "info@example",
			},
			wantErr: ErrInvalidLicenseExpiryTemplateData,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := &LicenseExpiryTemplateData{
				Username:     tt.args.username,
				FirstName:    tt.args.firstName,
				License:      tt.args.license,
				RenewEmail:   tt.args.renewEmail,
				SupportEmail: tt.args.supportEmail,
			}

			assert.ErrorIs(t, data.Validate(), tt.wantErr)
		})
	}
}

func TestLicenseExpiryTemplateData_Get(t *testing.T) {
	t.Parallel()

	data := &LicenseExpiryTemplateData{
		Username:     "test-user",
		FirstName:    "Test",
		License:      &license.License{ID: xid.New()},
		RenewEmail:   "renew@example.com",
		SupportEmail: "info@example.com",
	}

	assert.Equal(t, data, data.Get())
}
