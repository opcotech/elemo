package mock

import (
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	mock.Mock
}

func (m *Logger) Sugar() *zap.SugaredLogger {
	args := m.Called()
	return args.Get(0).(*zap.SugaredLogger)
}

func (m *Logger) Named(s string) *zap.Logger {
	args := m.Called(s)
	return args.Get(0).(*zap.Logger)
}

func (m *Logger) WithOptions(opts ...zap.Option) *zap.Logger {
	args := m.Called(opts)
	return args.Get(0).(*zap.Logger)
}

func (m *Logger) With(fields ...zap.Field) *zap.Logger {
	args := m.Called(fields)
	return args.Get(0).(*zap.Logger)
}

func (m *Logger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	args := m.Called(lvl, msg)
	return args.Get(0).(*zapcore.CheckedEntry)
}

func (m *Logger) Log(lvl zapcore.Level, msg string, fields ...zap.Field) {
	m.Called(lvl, msg, fields)
}

func (m *Logger) Debug(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *Logger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *Logger) Warn(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *Logger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *Logger) DPanic(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *Logger) Panic(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *Logger) Fatal(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *Logger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Logger) Core() zapcore.Core {
	args := m.Called()
	return args.Get(0).(zapcore.Core)
}
