package mock

import (
	"context"
	"crypto/tls"
	"io"
	"net/smtp"

	"github.com/stretchr/testify/mock"

	"github.com/opcotech/elemo/internal/email"
)

type SMTPAuth struct {
	mock.Mock
}

func (a *SMTPAuth) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	args := a.Called(server)
	return args.String(0), args.Get(1).([]byte), args.Error(2)
}

func (a *SMTPAuth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	args := a.Called(fromServer, more)
	return args.Get(0).([]byte), args.Error(1)
}

type NetSMTPClientOld struct {
	mock.Mock
}

func (n *NetSMTPClientOld) Close() error {
	args := n.Called()
	return args.Error(0)
}

func (n *NetSMTPClientOld) Hello(localName string) error {
	args := n.Called(localName)
	return args.Error(0)
}

func (n *NetSMTPClientOld) StartTLS(config *tls.Config) error {
	args := n.Called(config)
	return args.Error(0)
}

func (n *NetSMTPClientOld) TLSConnectionState() (state tls.ConnectionState, ok bool) {
	args := n.Called()
	return args.Get(0).(tls.ConnectionState), args.Bool(1)
}

func (n *NetSMTPClientOld) Verify(addr string) error {
	args := n.Called(addr)
	return args.Error(0)
}

func (n *NetSMTPClientOld) Auth(a smtp.Auth) error {
	args := n.Called(a)
	return args.Error(0)
}

func (n *NetSMTPClientOld) Mail(from string) error {
	args := n.Called(from)
	return args.Error(0)
}

func (n *NetSMTPClientOld) Rcpt(to string) error {
	args := n.Called(to)
	return args.Error(0)
}

func (n *NetSMTPClientOld) Data() (io.WriteCloser, error) {
	args := n.Called()
	return args.Get(0).(io.WriteCloser), args.Error(1)
}

func (n *NetSMTPClientOld) Extension(ext string) (bool, string) {
	args := n.Called(ext)
	return args.Bool(0), args.String(1)
}

func (n *NetSMTPClientOld) Reset() error {
	args := n.Called()
	return args.Error(0)
}

func (n *NetSMTPClientOld) Noop() error {
	args := n.Called()
	return args.Error(0)
}

func (n *NetSMTPClientOld) Quit() error {
	args := n.Called()
	return args.Error(0)
}

type SMTPClientOld struct {
	mock.Mock
}

func (s *SMTPClientOld) SendEmail(ctx context.Context, subject, to string, template *email.Template) error {
	args := s.Called(ctx, subject, to, template)
	return args.Error(0)
}
