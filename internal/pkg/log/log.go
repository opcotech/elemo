package log

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/opcotech/elemo/internal/pkg"
)

var (
	ErrNoLogger            = errors.New("no logger")             // the logger is missing
	ErrInvalidLogLevel     = errors.New("invalid log level")     // invalid log level
	ErrInvalidLoggerConfig = errors.New("invalid logger config") // invalid logger config

	globalLogger, _ = zap.NewProduction()
	loggerLock      sync.Mutex
)

// ctxKey is a type alias for context key.
type ctxKey string

// ZapLogger is a type alias for zap.Logger.
type ZapLogger = zap.Logger

// Logger defines the interface for the application logger.
type Logger interface {
	Sugar() *zap.SugaredLogger
	Named(s string) *zap.Logger
	WithOptions(opts ...zap.Option) *zap.Logger
	With(fields ...zap.Field) *zap.Logger
	Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry
	Log(lvl zapcore.Level, msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	DPanic(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Sync() error
	Core() zapcore.Core
}

// ConfigureLogger configures the logger then returns it.
func ConfigureLogger(level string) (Logger, error) {
	var err error

	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, errors.Join(ErrInvalidLogLevel, err)
	}

	zapConf := zap.NewProductionConfig()
	zapConf.Level = logLevel

	loggerLock.Lock()
	defer loggerLock.Unlock()

	if globalLogger, err = zapConf.Build(); err != nil {
		return nil, errors.Join(ErrInvalidLoggerConfig, err)
	}

	globalLogger = globalLogger.Named("root")

	return globalLogger, nil
}

// DefaultLogger returns the global logger.
func DefaultLogger() Logger {
	return globalLogger
}

// WithContext returns a new context with the logger. If the logger is not
// provided, it returns the context with the global logger assigned.
func WithContext(ctx context.Context, logger Logger) context.Context {
	ctxLogger := logger
	if ctxLogger == nil {
		ctxLogger = globalLogger
	}

	return context.WithValue(ctx, pkg.CtxKeyLogger, ctxLogger)
}

// FromContext returns the logger from the context. If the logger is not
// found in the context, it returns the global logger.
func FromContext(ctx context.Context) Logger {
	if ctxLogger, ok := ctx.Value(pkg.CtxKeyLogger).(Logger); ok {
		return ctxLogger
	}

	return globalLogger
}

// Log logs the message with the given level.
// NOTE: This may log sensitive information.
// TODO: Implement a log filter to filter out sensitive information.
func Log(ctx context.Context, level zapcore.Level, message string, fields ...zap.Field) {
	commonFields := make([]zap.Field, 0)
	logger := FromContext(ctx)
	logger.Log(level, message, append(fields, commonFields...)...)
}

// Debug logs the message with the debug level.
func Debug(ctx context.Context, message string, fields ...zap.Field) {
	Log(ctx, zapcore.DebugLevel, message, fields...)
}

// Info logs the message with the info level.
func Info(ctx context.Context, message string, fields ...zap.Field) {
	Log(ctx, zapcore.InfoLevel, message, fields...)
}

// Warn logs the message with the warn level.
func Warn(ctx context.Context, message string, fields ...zap.Field) {
	Log(ctx, zapcore.WarnLevel, message, fields...)
}

// Error logs the message with the error level.
func Error(ctx context.Context, err error, fields ...zap.Field) {
	Log(ctx, zapcore.ErrorLevel, err.Error(), append(fields, zap.Error(err))...)
}

// Fatal logs the message with the fatal level.
func Fatal(ctx context.Context, message string, fields ...zap.Field) {
	Log(ctx, zapcore.FatalLevel, message, fields...)
}

// Panic logs the message with the panic level.
func Panic(ctx context.Context, message string, fields ...zap.Field) {
	Log(ctx, zapcore.PanicLevel, message, fields...)
}
