package smtp

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/smtp"

	"github.com/Shopify/gomail"
	"go.opentelemetry.io/otel/trace"

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
		c.Client = wrapped
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
		c.Config = cfg
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
		c.Logger = logger
		return nil
	}
}

// WithTracer returns an Option that configures a Client with the given
// tracer.
func WithTracer(tracer trace.Tracer) Option {
	return func(c *Client) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}
		c.Tracer = tracer
		return nil
	}
}

// Client is the simplified SMTP client used for sending notification emails.
type Client struct {
	Client WrappedClient      `validate:"required"`
	Config *config.SMTPConfig `validate:"required"`
	Logger log.Logger         `validate:"required"`
	Tracer trace.Tracer       `validate:"required"`
}

// Authenticate initiates the SMTP handshake and authenticates the client.
func (c *Client) Authenticate(ctx context.Context) error {
	_, span := c.Tracer.Start(ctx, "smtp.Client/ComposeMessage")
	defer span.End()

	auth := smtp.PlainAuth("", c.Config.Username, c.Config.Password, c.Config.Host)
	if err := c.Client.Auth(auth); err != nil {
		return errors.Join(ErrAuthFailed, err)
	}

	return nil
}

// composeMessage composes an email message with the given subject and body,
// then returns the message.
func (c *Client) composeMessage(ctx context.Context, subject, to string, template *email.Template) (*gomail.Message, error) {
	_, span := c.Tracer.Start(ctx, "smtp.Client/ComposeMessage")
	defer span.End()

	if err := validate.Struct(template); err != nil {
		return nil, errors.Join(ErrComposeEmail, err)
	}

	htmlBody, err := template.Render()
	if err != nil {
		return nil, errors.Join(ErrComposeEmail, err)
	}

	headers := map[string][]string{
		"From":                      {c.Config.FromAddress},
		"To":                        {to},
		"Subject":                   {subject},
		"Reply-To":                  {c.Config.ReplyToAddress},
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
	_, span := c.Tracer.Start(ctx, "smtp.Client/SendEmail")
	defer span.End()

	message, err := c.composeMessage(ctx, subject, to, template)
	if err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	if err := c.Client.Mail(c.Config.FromAddress); err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	if err := c.Client.Rcpt(to); err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	w, err := c.Client.Data()
	if err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	if _, err := message.WriteTo(w); err != nil {
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
		Logger: log.DefaultLogger(),
		Tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}
