package mock

import (
	"crypto/tls"
	"io"
	"net/smtp"

	"github.com/stretchr/testify/mock"
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

type NetSMTPClient struct {
	mock.Mock
}

func (n *NetSMTPClient) Close() error {
	args := n.Called()
	return args.Error(0)
}

func (n *NetSMTPClient) Hello(localName string) error {
	args := n.Called(localName)
	return args.Error(0)
}

func (n *NetSMTPClient) StartTLS(config *tls.Config) error {
	args := n.Called(config)
	return args.Error(0)
}

func (n *NetSMTPClient) TLSConnectionState() (state tls.ConnectionState, ok bool) {
	args := n.Called()
	return args.Get(0).(tls.ConnectionState), args.Bool(1)
}

func (n *NetSMTPClient) Verify(addr string) error {
	args := n.Called(addr)
	return args.Error(0)
}

func (n *NetSMTPClient) Auth(a smtp.Auth) error {
	args := n.Called(a)
	return args.Error(0)
}

func (n *NetSMTPClient) Mail(from string) error {
	args := n.Called(from)
	return args.Error(0)
}

func (n *NetSMTPClient) Rcpt(to string) error {
	args := n.Called(to)
	return args.Error(0)
}

func (n *NetSMTPClient) Data() (io.WriteCloser, error) {
	args := n.Called()
	return args.Get(0).(io.WriteCloser), args.Error(1)
}

func (n *NetSMTPClient) Extension(ext string) (bool, string) {
	args := n.Called(ext)
	return args.Bool(0), args.String(1)
}

func (n *NetSMTPClient) Reset() error {
	args := n.Called()
	return args.Error(0)
}

func (n *NetSMTPClient) Noop() error {
	args := n.Called()
	return args.Error(0)
}

func (n *NetSMTPClient) Quit() error {
	args := n.Called()
	return args.Error(0)
}
