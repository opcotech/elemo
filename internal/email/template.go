package email

import (
	"errors"
	"html/template"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

// TemplateData represents the data needed to render an email template.
type TemplateData interface {
	// Get returns the template data.
	RenderData() any
}

// Template is a struct that represents the data needed to render
// an email template.
type Template struct {
	Path string       `validate:"required,filepath"`
	Data TemplateData `validate:"required"`
}

// Validate validates the password reset email template.
func (t *Template) Validate() error {
	if err := validate.Struct(t); err != nil {
		return errors.Join(ErrTemplateInvalid, err)
	}
	return nil
}

// Render returns the rendered template.
func (t *Template) Render() (string, error) {
	return emailBody[*template.Template](t.Path, t.Data.RenderData(), template.ParseFiles)
}

// NewTemplate returns a new email template.
func NewTemplate(path string, data TemplateData) (*Template, error) {
	t := &Template{
		Path: path,
		Data: data,
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}
