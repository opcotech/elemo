package mock

import (
	"net/smtp"

	"github.com/stretchr/testify/mock"
)

type SMTPAuth struct {
	mock.Mock
}

func (S *SMTPAuth) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	args := S.Called(server)
	return args.String(0), args.Get(1).([]byte), args.Error(2)
}

func (S *SMTPAuth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	args := S.Called(fromServer, more)
	return args.Get(0).([]byte), args.Error(1)
}
