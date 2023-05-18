package mock

import "github.com/stretchr/testify/mock"

type Buffer struct {
	mock.Mock
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	args := b.Called(p)
	return args.Int(0), args.Error(1)
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	args := b.Called(p)
	return args.Int(0), args.Error(1)
}

func (b *Buffer) Close() error {
	args := b.Called()
	return args.Error(0)
}
