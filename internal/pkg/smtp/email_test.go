package smtp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opcotech/elemo/internal/email"
)

type testTemplateData struct {
	Field string `validate:"required"`
}

func (d *testTemplateData) Get() interface{} {
	return d
}

func TestNewEmail(t *testing.T) {
	type args struct {
		subject  string
		from     string
		to       []string
		template Template
	}
	tests := []struct {
		name    string
		args    args
		want    *Email
		wantErr error
	}{
		{
			name: "valid email",
			args: args{
				subject: "subject",
				from:    "info@example.com",
				to: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				template: &email.Template{
					EmailMimeType: email.MimeTypeHTML,
					Path:          "/path/to/template.html",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
			want: &Email{
				Subject: "subject",
				From:    "info@example.com",
				To: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				Template: &email.Template{
					EmailMimeType: email.MimeTypeHTML,
					Path:          "/path/to/template.html",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
		},
		{
			name: "invalid email",
			args: args{
				subject: "",
				from:    "info@example.com",
				to: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				template: &email.Template{
					EmailMimeType: email.MimeTypeHTML,
					Path:          "/path/to/template.html",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
			wantErr: ErrInvalidEmail,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e, err := NewEmail(tt.args.subject, tt.args.from, tt.args.to, tt.args.template)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, e)
		})
	}
}

func TestEmail_MimeHeader(t *testing.T) {
	type fields struct {
		Subject  string
		From     string
		To       []string
		Template Template
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "html email",
			fields: fields{
				Subject: "subject",
				From:    "info@example.com",
				To: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				Template: &email.Template{
					EmailMimeType: email.MimeTypeHTML,
					Path:          "/path/to/template.html",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
			want: "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n",
		},
		{
			name: "plain text email",
			fields: fields{
				Subject: "subject",
				From:    "info@example.com",
				To: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				Template: &email.Template{
					EmailMimeType: email.MimeTypePlainText,
					Path:          "/path/to/template.txt",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
			want: "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &Email{
				Subject:  tt.fields.Subject,
				From:     tt.fields.From,
				To:       tt.fields.To,
				Template: tt.fields.Template,
			}

			assert.Equal(t, tt.want, e.MimeHeader())
		})
	}
}

func TestEmail_Validate(t *testing.T) {
	type args struct {
		subject  string
		from     string
		to       []string
		template Template
	}
	tests := []struct {
		name    string
		args    args
		want    *Email
		wantErr error
	}{
		{
			name: "valid email",
			args: args{
				subject: "subject",
				from:    "info@example.com",
				to: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				template: &email.Template{
					EmailMimeType: email.MimeTypeHTML,
					Path:          "/path/to/template.html",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
			want: &Email{
				Subject: "subject",
				From:    "info@example.com",
				To: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				Template: &email.Template{
					EmailMimeType: email.MimeTypeHTML,
					Path:          "/path/to/template.html",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
		},
		{
			name: "invalid subject",
			args: args{
				subject: "",
				from:    "info@example.com",
				to: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				template: &email.Template{
					EmailMimeType: email.MimeTypeHTML,
					Path:          "/path/to/template.html",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "invalid from",
			args: args{
				subject: "subject",
				from:    "",
				to: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				template: &email.Template{
					EmailMimeType: email.MimeTypeHTML,
					Path:          "/path/to/template.html",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "invalid to",
			args: args{
				subject: "subject",
				from:    "info@example.com",
				to: []string{
					"recipient1@example",
				},
				template: &email.Template{
					EmailMimeType: email.MimeTypeHTML,
					Path:          "/path/to/template.html",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "empty to",
			args: args{
				subject: "subject",
				from:    "info@example.com",
				to:      []string{},
				template: &email.Template{
					EmailMimeType: email.MimeTypeHTML,
					Path:          "/path/to/template.html",
					Data: &testTemplateData{
						Field: "value",
					},
				},
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "invalid template",
			args: args{
				subject: "subject",
				from:    "info@example.com",
				to: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				template: &email.Template{},
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "missing template",
			args: args{
				subject: "subject",
				from:    "info@example.com",
				to: []string{
					"recipient1@example.com",
					"recipient2@example.com",
				},
				template: nil,
			},
			wantErr: ErrInvalidEmail,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e, err := NewEmail(tt.args.subject, tt.args.from, tt.args.to, tt.args.template)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, e)
		})
	}
}
