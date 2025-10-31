package smtp

import (
	"context"
	"errors"

	"github.com/Shopify/gomail"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/pkg/validate"
)

// WrappedClient is the interface that wraps the SMTP client methods used by
// the Client.
//
//go:generate mockgen -source=smtp.go -destination=../../testutil/mock/smtp_gen.go -package=mock -mock_names WrappedClient=WrappedClient
type WrappedClient interface {
	DialAndSend(messages ...*gomail.Message) error
}

// Option is a function that configures a Client.
type Option func(*Client) error

// WithWrappedClient returns an Option that configures a Client with the given
// wrapped SMTP client. The given SMTP client is used as a base that is
// configured.
func WithWrappedClient(wrapped WrappedClient) Option {
	return func(c *Client) error {
		if wrapped == nil {
			return ErrNoSMTPClient
		}
		c.client = wrapped
		return nil
	}
}

// WithConfig returns an Option that configures a Client with the given
// configuration.
func WithConfig(cfg *config.SMTPConfig) Option {
	return func(c *Client) error {
		if cfg == nil {
			return config.ErrNoConfig
		}
		c.config = cfg
		return nil
	}
}

// WithLogger returns an Option that configures a Client with the given
// logger.
func WithLogger(logger log.Logger) Option {
	return func(c *Client) error {
		if logger == nil {
			return log.ErrNoLogger
		}
		c.logger = logger
		return nil
	}
}

// WithTracer returns an Option that configures a Client with the given
// tracer.
func WithTracer(tracer tracing.Tracer) Option {
	return func(c *Client) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}
		c.tracer = tracer
		return nil
	}
}

// Client is the simplified SMTP client used for sending notification emails.
type Client struct {
	client WrappedClient      `validate:"required"`
	config *config.SMTPConfig `validate:"required"`
	logger log.Logger         `validate:"required"`
	tracer tracing.Tracer     `validate:"required"`
}

// SendEmail sends an email with the given subject and body to the recipient.
func (c *Client) SendEmail(ctx context.Context, subject, to string, template *email.Template) error {
	_, span := c.tracer.Start(ctx, "smtp.Client/SendEmail")
	defer span.End()

	if err := validate.Struct(template); err != nil {
		c.logger.Error(
			ctx,
			ErrSendEmail.Error(),
			log.WithSubject(subject),
			log.WithValue(template),
			log.WithAction(log.ActionEmailSend),
			log.WithError(err),
		)
		return errors.Join(ErrSendEmail, err)
	}

	htmlBody, err := template.Render()
	if err != nil {
		c.logger.Error(
			ctx,
			ErrSendEmail.Error(),
			log.WithSubject(subject),
			log.WithValue(template),
			log.WithAction(log.ActionEmailSend),
			log.WithError(err),
		)
		return errors.Join(ErrSendEmail, err)
	}

	message := gomail.NewMessage()
	message.SetBody("text/html", htmlBody)
	message.SetHeaders(map[string][]string{
		"From":     {message.FormatAddress(c.config.FromAddress, "")},
		"To":       {message.FormatAddress(to, "")},
		"Subject":  {subject},
		"Reply-To": {c.config.ReplyToAddress},
	})

	if err := c.client.DialAndSend(message); err != nil {
		c.logger.Error(
			ctx,
			ErrSendEmail.Error(),
			log.WithSubject(subject),
			log.WithValue(template),
			log.WithAction(log.ActionEmailSend),
			log.WithError(err),
		)
		return errors.Join(ErrSendEmail, err)
	}

	c.logger.Info(
		ctx,
		"email sent",
		log.WithSubject(subject),
		log.WithAction(log.ActionEmailSend),
	)

	return nil
}

// NewClient returns a new Client.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}
