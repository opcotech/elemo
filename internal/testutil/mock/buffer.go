package mock

import "github.com/stretchr/testify/mock"

type BufferOld struct {
	mock.Mock
}

func (b *BufferOld) Read(p []byte) (n int, err error) {
	args := b.Called(p)
	return args.Int(0), args.Error(1)
}

func (b *BufferOld) Write(p []byte) (n int, err error) {
	args := b.Called(p)
	return args.Int(0), args.Error(1)
}

func (b *BufferOld) Close() error {
	args := b.Called()
	return args.Error(0)
}
