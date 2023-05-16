package smtp

import (
	"context"
	"errors"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/testutil"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func testSendEmailFunc(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	return nil
}

func testSendEmailFailingFunc(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	return errors.New("failed to send email")
}

func TestNewClient(t *testing.T) {
	type args struct {
		address       string
		auth          smtp.Auth
		sendEmailFunc SendEmailFunc
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr error
	}{
		{
			name: "new smtp client",
			args: args{
				address:       "smtp.email.server:123",
				auth:          new(mock.SMTPAuth),
				sendEmailFunc: testSendEmailFunc,
			},
			want: &Client{
				Address:       "smtp.email.server:123",
				Auth:          new(mock.SMTPAuth),
				SendEmailFunc: testSendEmailFunc,
			},
		},
		{
			name: "invalid smtp client",
			args: args{
				address:       "smtp.email.server",
				auth:          new(mock.SMTPAuth),
				sendEmailFunc: testSendEmailFunc,
			},
			wantErr: ErrInvalidClient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewClient(tt.args.address, tt.args.auth, tt.args.sendEmailFunc)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.Equal(t, tt.want.Address, got.Address)
				assert.Equal(t, tt.want.Auth, got.Auth)
				assert.NotNil(t, got.SendEmailFunc)
			}
		})
	}
}

func TestClient_Validate(t *testing.T) {
	type fields struct {
		Address       string
		Auth          smtp.Auth
		SendEmailFunc SendEmailFunc
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid smtp client",
			fields: fields{
				Address:       "smtp.email.server:123",
				Auth:          new(mock.SMTPAuth),
				SendEmailFunc: testSendEmailFunc,
			},
		},
		{
			name: "invalid smtp client address",
			fields: fields{
				Address:       "smtp.email.server",
				Auth:          new(mock.SMTPAuth),
				SendEmailFunc: testSendEmailFunc,
			},
			wantErr: ErrInvalidClient,
		},
		{
			name: "invalid smtp client auth",
			fields: fields{
				Address:       "smtp.email.server:123",
				Auth:          nil,
				SendEmailFunc: testSendEmailFunc,
			},
			wantErr: ErrInvalidClient,
		},
		{
			name: "invalid smtp client send email func",
			fields: fields{
				Address: "smtp.email.server:123",
				Auth:    new(mock.SMTPAuth),
			},
			wantErr: ErrInvalidClient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &Client{
				Address:       tt.fields.Address,
				Auth:          tt.fields.Auth,
				SendEmailFunc: tt.fields.SendEmailFunc,
			}

			assert.ErrorIs(t, c.Validate(), tt.wantErr)
		})
	}
}

func TestClient_SendEmail(t *testing.T) {
	type fields struct {
		Address       string
		Auth          smtp.Auth
		SendEmailFunc SendEmailFunc
	}
	type args struct {
		ctx   context.Context
		email *Email
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "send email",
			fields: fields{
				Address:       "smtp.email.server:123",
				Auth:          new(mock.SMTPAuth),
				SendEmailFunc: testSendEmailFunc,
			},
			args: args{
				ctx: context.Background(),
				email: &Email{
					From: "info@example.com",
					To: []string{
						"recipient1@example.com",
						"recipient2@example.com",
					},
					Subject: "test subject",
					Template: &email.Template{
						EmailMimeType: email.MimeTypeHTML,
						Path:          testutil.NewTempFile(t, testutil.GenerateRandomString(10), "{{ .Field }}"),
						Data: &testTemplateData{
							Field: "value",
						},
					},
				},
			},
		},
		{
			name: "no template file",
			fields: fields{
				Address:       "smtp.email.server:123",
				Auth:          new(mock.SMTPAuth),
				SendEmailFunc: testSendEmailFunc,
			},
			args: args{
				ctx: context.Background(),
				email: &Email{
					From: "info@example.com",
					To: []string{
						"recipient1@example.com",
						"recipient2@example.com",
					},
					Subject: "test subject",
					Template: &email.Template{
						EmailMimeType: email.MimeTypeHTML,
						Path:          "/invalid/path.html",
						Data: &testTemplateData{
							Field: "value",
						},
					},
				},
			},
			wantErr: ErrComposeEmail,
		},
		{
			name: "send email failed",
			fields: fields{
				Address:       "smtp.email.server:123",
				Auth:          new(mock.SMTPAuth),
				SendEmailFunc: testSendEmailFailingFunc,
			},
			args: args{
				ctx: context.Background(),
				email: &Email{
					From: "info@example.com",
					To: []string{
						"recipient1@example.com",
						"recipient2@example.com",
					},
					Subject: "test subject",
					Template: &email.Template{
						EmailMimeType: email.MimeTypeHTML,
						Path:          testutil.NewTempFile(t, testutil.GenerateRandomString(10), "{{ .Field }}"),
						Data: &testTemplateData{
							Field: "value",
						},
					},
				},
			},
			wantErr: ErrSendEmail,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &Client{
				Address:       tt.fields.Address,
				Auth:          tt.fields.Auth,
				SendEmailFunc: tt.fields.SendEmailFunc,
			}

			err := c.SendEmail(context.Background(), tt.args.email)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
