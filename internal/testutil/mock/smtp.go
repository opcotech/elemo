package mock

import (
	"context"
	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/email"
)

type SMTPClientOld struct {
	mock.Mock
}

func (s *SMTPClientOld) SendEmail(ctx context.Context, subject, to string, template *email.Template) error {
	args := s.Called(ctx, subject, to, template)
	return args.Error(0)
}
