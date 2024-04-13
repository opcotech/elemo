package smtp

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/smtp"

	"github.com/Shopify/gomail"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/pkg/validate"
)

// WrappedClient is the interface that wraps the SMTP client methods used by
// the Client.
type WrappedClient interface {
	Close() error
	Hello(localName string) error
	StartTLS(config *tls.Config) error
	TLSConnectionState() (state tls.ConnectionState, ok bool)
	Verify(addr string) error
	Auth(a smtp.Auth) error
	Mail(from string) error
	Rcpt(to string) error
	Data() (io.WriteCloser, error)
	Extension(ext string) (bool, string)
	Reset() error
	Noop() error
	Quit() error
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

// Authenticate initiates the SMTP handshake and authenticates the client.
func (c *Client) Authenticate(ctx context.Context) error {
	_, span := c.tracer.Start(ctx, "smtp.Client/Authenticate")
	defer span.End()

	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
	if err := c.client.Auth(auth); err != nil {
		return errors.Join(ErrAuthFailed, err)
	}

	return nil
}

// composeMessage composes an email message with the given subject and body,
// then returns the message.
func (c *Client) composeMessage(_ context.Context, subject, to string, template *email.Template) (*gomail.Message, error) {
	if err := validate.Struct(template); err != nil {
		return nil, errors.Join(ErrComposeEmail, err)
	}

	htmlBody, err := template.Render()
	if err != nil {
		return nil, errors.Join(ErrComposeEmail, err)
	}

	headers := map[string][]string{
		"From":                      {c.config.FromAddress},
		"To":                        {to},
		"Subject":                   {subject},
		"Reply-To":                  {c.config.ReplyToAddress},
		"Content-Transfer-Encoding": {"8bit"},
		"Auto-Submitted":            {"auto-generated"},
		"Precedence":                {"bulk"},
	}

	message := gomail.NewMessage()
	message.SetHeaders(headers)
	message.SetBody("text/html", htmlBody)

	return message, nil
}

// SendEmail sends an email with the given subject and body to the recipient.
func (c *Client) SendEmail(ctx context.Context, subject, to string, template *email.Template) error {
	_, span := c.tracer.Start(ctx, "smtp.Client/SendEmail")
	defer span.End()

	message, err := c.composeMessage(ctx, subject, to, template)
	if err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	if err := c.client.Mail(c.config.FromAddress); err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	if err := c.client.Rcpt(to); err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	w, err := c.client.Data()
	if err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	if wrote, err := message.WriteTo(w); err != nil || wrote == 0 {
		if err == nil {
			err = ErrNoBytesWritten
		}
		return errors.Join(ErrComposeEmail, err)
	}

	if err := w.Close(); err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

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
