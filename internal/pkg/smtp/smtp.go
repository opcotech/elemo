package smtp

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"

	"go.uber.org/zap/buffer"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

// SendEmailFunc represents a function that sends an email to a remote SMTP
// server.
type SendEmailFunc func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error

// Client is a struct that represents an SMTP client.
type Client struct {
	Address       string        `validate:"required,hostname_port"`
	Auth          smtp.Auth     `validate:"required"`
	SendEmailFunc SendEmailFunc `validate:"required"`
}

// Validate validates the SMTP client.
func (c *Client) Validate() error {
	if err := validate.Struct(c); err != nil {
		return errors.Join(ErrInvalidClient, err)
	}
	return nil
}

// SendEmail sends an email.
func (c *Client) SendEmail(_ context.Context, email *Email) error {
	if err := validate.Struct(email); err != nil {
		return err
	}

	body := new(buffer.Buffer)

	if _, err := body.Write([]byte(fmt.Sprintf("%s \n%s\n\n", email.Subject, email.MimeHeader()))); err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	content, err := email.Template.Body()
	if err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	if _, err := body.Write([]byte(content)); err != nil {
		return errors.Join(ErrComposeEmail, err)
	}

	if err := c.SendEmailFunc(c.Address, c.Auth, email.From, email.To, body.Bytes()); err != nil {
		return errors.Join(ErrSendEmail, err)
	}

	return nil
}

// NewClient returns a new instance of an SMTP client.
func NewClient(address string, auth smtp.Auth, sendEmailFunc SendEmailFunc) (*Client, error) {
	c := &Client{
		Address:       address,
		Auth:          auth,
		SendEmailFunc: sendEmailFunc,
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}
