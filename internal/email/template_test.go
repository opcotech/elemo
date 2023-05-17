package email

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opcotech/elemo/internal/testutil"
)

type testTemplateData struct {
	Username     string `validate:"required"`
	FirstName    string `validate:"omitempty"`
	ServerURL    string `validate:"required,url"`
	SupportEmail string `validate:"required,email"`
}

func (d *testTemplateData) Get() interface{} {
	return d
}

func TestNewTemplate(t *testing.T) {
	type args struct {
		emailMimeType string
		path          string
		data          TemplateData
	}
	tests := []struct {
		name    string
		args    args
		want    *Template
		wantErr error
	}{
		{
			name: "valid email template",
			args: args{
				emailMimeType: MimeTypeHTML,
				path:          "/test.html",
				data: &testTemplateData{
					Username:     "test-user",
					FirstName:    "Test",
					ServerURL:    "https://example.com",
					SupportEmail: "info@example.com",
				},
			},
			want: &Template{
				Path: "/test.html",
				Data: &testTemplateData{
					Username:     "test-user",
					FirstName:    "Test",
					ServerURL:    "https://example.com",
					SupportEmail: "info@example.com",
				},
			},
		},
		{
			name: "invalid path",
			args: args{
				emailMimeType: MimeTypeHTML,
				path:          "",
				data: &testTemplateData{
					Username:     "test-user",
					FirstName:    "Test",
					ServerURL:    "https://example.com",
					SupportEmail: "info@example.com",
				},
			},
			wantErr: ErrTemplateInvalid,
		},
		{
			name: "invalid data",
			args: args{
				emailMimeType: MimeTypeHTML,
				path:          "/test.html",
				data: &testTemplateData{
					Username: "",
				},
			},
			wantErr: ErrTemplateInvalid,
		},
		{
			name: "invalid data type",
			args: args{
				emailMimeType: MimeTypeHTML,
				path:          "/test.html",
				data:          nil,
			},
			wantErr: ErrTemplateInvalid,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewTemplate(tt.args.emailMimeType, tt.args.path, tt.args.data)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTemplate_Validate(t *testing.T) {
	type fields struct {
		EmailMimeType string
		Path          string
		Data          TemplateData
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid email template",
			fields: fields{
				EmailMimeType: MimeTypeHTML,
				Path:          "/test.html",
				Data: &testTemplateData{
					Username:     "test-user",
					FirstName:    "Test",
					ServerURL:    "https://example.com",
					SupportEmail: "info@example.com",
				},
			},
		},
		{
			name: "invalid path",
			fields: fields{
				EmailMimeType: MimeTypeHTML,
				Path:          "",
				Data: &testTemplateData{
					Username:     "test-user",
					FirstName:    "Test",
					ServerURL:    "https://example.com",
					SupportEmail: "info@example.com",
				},
			},
			wantErr: ErrTemplateInvalid,
		},
		{
			name: "invalid data",
			fields: fields{
				EmailMimeType: MimeTypeHTML,
				Path:          "/test.html",
				Data: &testTemplateData{
					Username: "",
				},
			},
			wantErr: ErrTemplateInvalid,
		},
		{
			name: "invalid data type",
			fields: fields{
				EmailMimeType: MimeTypeHTML,
				Path:          "/test.html",
				Data:          nil,
			},
			wantErr: ErrTemplateInvalid,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpl := &Template{
				Path: tt.fields.Path,
				Data: tt.fields.Data,
			}

			assert.ErrorIs(t, tmpl.Validate(), tt.wantErr)
		})
	}
}

func TestTemplate_Body(t *testing.T) {
	type fields struct {
		EmailMimeType string
		Path          string
		Data          TemplateData
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr error
	}{
		{
			name: "valid email template",
			fields: fields{
				Data: &testTemplateData{
					Username:     "test-user",
					FirstName:    "Test",
					ServerURL:    "https://example.com",
					SupportEmail: "info@example.com",
				},
				EmailMimeType: MimeTypeHTML,
				Path: testutil.NewTempFile(t, testutil.GenerateRandomString(10),
					"Hello {{ .FirstName }} ({{ .Username }})!\n"+
						"Welcome to {{ .ServerURL }}. If you have any questions, please contact {{ .SupportEmail }}.",
				),
			},
			want: "Hello Test (test-user)!\nWelcome to https://example.com. If you have any questions, please contact info@example.com.",
		},
		{
			name: "invalid email template",
			fields: fields{
				Data: &testTemplateData{
					Username:     "test-user",
					FirstName:    "Test",
					ServerURL:    "https://example.com",
					SupportEmail: "info@example.com",
				},
				EmailMimeType: MimeTypeHTML,
				Path: testutil.NewTempFile(t, testutil.GenerateRandomString(10),
					"{{ ?? }}",
				),
			},
			wantErr: ErrTemplateParse,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpl := &Template{
				Path: tt.fields.Path,
				Data: tt.fields.Data,
			}

			got, err := tmpl.Render()

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
