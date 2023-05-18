package email

import (
	"bytes"
	"errors"
	"io"
)

const (
	MimeTypeHTML      = "text/html"
	MimeTypePlainText = "text/plain"
)

// Executor executes parsed templates.
type Executor interface {
	Execute(wr io.Writer, data any) error
}

// TemplateParserFunc parses email templates.
type TemplateParserFunc[T Executor] func(filenames ...string) (T, error)

// emailBody returns the rendered email body.
func emailBody[T Executor](path string, data any, parser TemplateParserFunc[T]) (string, error) {
	tmpl, err := parser(path)
	if err != nil {
		return "", errors.Join(ErrTemplateParse, err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return "", errors.Join(ErrTemplateExecute, err)
	}

	return buf.String(), nil
}
