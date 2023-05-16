package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOrganizationInviteTemplateData(t *testing.T) {
	type args struct {
		username         string
		firstName        string
		organizationName string
		invitationURL    string
		supportEmail     string
	}
	tests := []struct {
		name    string
		args    args
		want    *OrganizationInviteTemplateData
		wantErr error
	}{
		{
			name: "valid template data",
			args: args{
				username:         "test-user",
				firstName:        "Test",
				organizationName: "ACME Inc.",
				invitationURL:    "https://example.com/org/1/invite",
				supportEmail:     "info@example.com",
			},
			want: &OrganizationInviteTemplateData{
				Username:         "test-user",
				FirstName:        "Test",
				OrganizationName: "ACME Inc.",
				InvitationURL:    "https://example.com/org/1/invite",
				SupportEmail:     "info@example.com",
			},
		},
		{
			name:    "invalid template data",
			args:    args{},
			wantErr: ErrInvalidOrganizationInviteTemplateData,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewOrganizationInviteTemplateData(
				tt.args.username,
				tt.args.firstName,
				tt.args.organizationName,
				tt.args.invitationURL,
				tt.args.supportEmail,
			)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOrganizationInviteTemplateData_Validate(t *testing.T) {
	type args struct {
		username         string
		firstName        string
		organizationName string
		invitationURL    string
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
				organizationName: "ACME Inc.",
				invitationURL:    "https://example.com/org/1/invite",
				supportEmail:     "info@example.com",
			},
		},
		{
			name: "invalid username",
			args: args{
				username:         "",
				firstName:        "Test",
				organizationName: "ACME Inc.",
				invitationURL:    "https://example.com/org/1/invite",
				supportEmail:     "info@example.com",
			},
			wantErr: ErrInvalidOrganizationInviteTemplateData,
		},
		{
			name: "invalid organization name",
			args: args{
				username:         "test-user",
				firstName:        "Test",
				organizationName: "",
				invitationURL:    "https://example.com/org/1/invite",
				supportEmail:     "info@example.com",
			},
			wantErr: ErrInvalidOrganizationInviteTemplateData,
		},
		{
			name: "invalid invitation url",
			args: args{
				username:         "test-user",
				firstName:        "Test",
				organizationName: "ACME Inc.",
				invitationURL:    "",
				supportEmail:     "info@example.com",
			},
			wantErr: ErrInvalidOrganizationInviteTemplateData,
		},
		{
			name: "invalid support email",
			args: args{
				username:         "test-user",
				firstName:        "Test",
				organizationName: "ACME Inc.",
				invitationURL:    "https://example.com/org/1/invite",
				supportEmail:     "info@example",
			},
			wantErr: ErrInvalidOrganizationInviteTemplateData,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := &OrganizationInviteTemplateData{
				Username:         tt.args.username,
				FirstName:        tt.args.firstName,
				OrganizationName: tt.args.organizationName,
				InvitationURL:    tt.args.invitationURL,
				SupportEmail:     tt.args.supportEmail,
			}

			assert.ErrorIs(t, data.Validate(), tt.wantErr)
		})
	}
}

func TestOrganizationInviteTemplateData_Get(t *testing.T) {
	t.Parallel()

	data := &OrganizationInviteTemplateData{
		Username:         "test-user",
		FirstName:        "Test",
		OrganizationName: "ACME Inc.",
		InvitationURL:    "https://example.com/org/1/invite",
		SupportEmail:     "info@example.com",
	}

	assert.Equal(t, data, data.Get())
}
