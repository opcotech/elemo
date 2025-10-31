package log

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"

	"github.com/opcotech/elemo/internal/pkg"
)

// Level is an alias for slog.Level.
type Level = slog.Level

// Attr is an alias for slog.Attr.
type Attr = slog.Attr

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
	LevelPanic = slog.Level(12) // Higher than Error
	LevelFatal = slog.Level(13) // Higher than Panic
)

var (
	ErrNoLogger            = errors.New("no logger")             // the logger is missing
	ErrInvalidLogLevel     = errors.New("invalid log level")     // invalid log level
	ErrInvalidLoggerConfig = errors.New("invalid logger config") // invalid logger config

	globalLogger Logger = &slogLogger{logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))}
	loggerLock   sync.Mutex
)

// Logger defines the interface for the application logger.
//
//go:generate mockgen -destination ../../testutil/mock/logger_gen.go -package mock github.com/opcotech/elemo/internal/pkg/log Logger
type Logger interface {
	Named(s string) Logger
	With(fields ...Attr) Logger
	Log(ctx context.Context, level Level, msg string, fields ...Attr)
	Debug(ctx context.Context, msg string, fields ...Attr)
	Info(ctx context.Context, msg string, fields ...Attr)
	Warn(ctx context.Context, msg string, fields ...Attr)
	Error(ctx context.Context, msg string, fields ...Attr)
	Panic(ctx context.Context, msg string, fields ...Attr)
	Fatal(ctx context.Context, msg string, fields ...Attr)
}

// slogLogger wraps *slog.Logger to implement the Logger interface.
type slogLogger struct {
	logger *slog.Logger
}

// Named returns a new logger with the specified name added as an attribute.
func (l *slogLogger) Named(s string) Logger {
	return &slogLogger{
		logger: l.logger.With(slog.String("logger", s)),
	}
}

// With returns a new logger with the specified attributes.
func (l *slogLogger) With(fields ...Attr) Logger {
	attrs := make([]slog.Attr, len(fields))
	copy(attrs, fields)
	return &slogLogger{
		logger: l.logger.With(attrsToArgs(attrs)...),
	}
}

// Log logs a message at the specified level.
func (l *slogLogger) Log(ctx context.Context, level Level, msg string, fields ...Attr) {
	mergedFields := extractStandardFields(ctx, fields...)
	attrs := make([]slog.Attr, len(mergedFields))
	copy(attrs, mergedFields)
	l.logger.Log(ctx, level, msg, attrsToArgs(attrs)...)
}

// Debug logs a message at debug level.
func (l *slogLogger) Debug(ctx context.Context, msg string, fields ...Attr) {
	mergedFields := extractStandardFields(ctx, fields...)
	l.logger.Debug(msg, attrsToArgs(mergedFields)...)
}

// Info logs a message at info level.
func (l *slogLogger) Info(ctx context.Context, msg string, fields ...Attr) {
	mergedFields := extractStandardFields(ctx, fields...)
	l.logger.Info(msg, attrsToArgs(mergedFields)...)
}

// Warn logs a message at warn level.
func (l *slogLogger) Warn(ctx context.Context, msg string, fields ...Attr) {
	mergedFields := extractStandardFields(ctx, fields...)
	l.logger.Warn(msg, attrsToArgs(mergedFields)...)
}

// Error logs a message at error level.
func (l *slogLogger) Error(ctx context.Context, msg string, fields ...Attr) {
	mergedFields := extractStandardFields(ctx, fields...)
	l.logger.Error(msg, attrsToArgs(mergedFields)...)
}

// Panic logs a message at panic level and then panics.
func (l *slogLogger) Panic(ctx context.Context, msg string, fields ...Attr) {
	mergedFields := extractStandardFields(ctx, fields...)
	l.logger.Log(ctx, LevelPanic, msg, attrsToArgs(mergedFields)...)
	panic(msg)
}

// Fatal logs a message at fatal level and then calls os.Exit(1).
func (l *slogLogger) Fatal(ctx context.Context, msg string, fields ...Attr) {
	mergedFields := extractStandardFields(ctx, fields...)
	l.logger.Log(ctx, LevelFatal, msg, attrsToArgs(mergedFields)...)
	os.Exit(1)
}

// attrsToArgs converts slog.Attr slice to ...any for slog methods that don't accept Attr directly.
func attrsToArgs(attrs []slog.Attr) []any {
	args := make([]any, len(attrs)*2)
	for i, attr := range attrs {
		args[i*2] = attr.Key
		args[i*2+1] = attr.Value.Any()
	}
	return args
}

// extractStandardFields extracts standard fields from context and merges them with provided fields.
// Explicit fields take precedence over auto-extracted fields.
func extractStandardFields(ctx context.Context, fields ...Attr) []Attr {
	// Create a map to track which fields are explicitly provided
	explicitFields := make(map[string]bool)
	for _, field := range fields {
		explicitFields[field.Key] = true
	}

	// Extract standard fields from context
	extracted := make([]Attr, 0, 4)

	// Extract request_id
	if !explicitFields[FieldRequestID] {
		if requestID := middleware.GetReqID(ctx); requestID != "" {
			extracted = append(extracted, WithRequestID(requestID))
		}
	}

	// Extract user_id
	if !explicitFields[FieldUserID] {
		if userID := pkg.CtxUserID(ctx); userID != "" {
			extracted = append(extracted, WithUserID(userID))
		}
	}

	// Extract trace_id
	if !explicitFields[FieldTraceID] {
		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			if traceID := span.SpanContext().TraceID().String(); traceID != "" && traceID != "00000000000000000000000000000000" {
				extracted = append(extracted, WithTraceID(traceID))
			}
		}
	}

	// TODO: Extract session_id (if available in context in the future)

	return append(fields, extracted...)
}

// ConfigureLogger configures the logger then returns it.
func ConfigureLogger(level string) (Logger, error) {
	var slogLevel slog.Level

	if level == "" {
		slogLevel = slog.LevelInfo
	} else {
		if err := slogLevel.UnmarshalText([]byte(level)); err != nil {
			return nil, errors.Join(ErrInvalidLogLevel, err)
		}
	}

	loggerLock.Lock()
	defer loggerLock.Unlock()

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slogLevel,
	})

	globalLogger = &slogLogger{logger: slog.New(handler)}
	globalLogger = globalLogger.Named("root")

	return globalLogger, nil
}

// DefaultLogger returns the global logger.
func DefaultLogger() Logger {
	return globalLogger
}

// SimpleLogger is used to log the message where only arguments are available.
type SimpleLogger struct {
	logger Logger
}

func (l *SimpleLogger) log(ctx context.Context, level Level, args ...interface{}) {
	message := args[len(args)-1].(string)

	logArgs := make([]Attr, (len(args)-1)/2)
	for i, j := 1, 0; i < len(args)-1; i += 2 {
		logArgs[j] = slog.Any(args[i].(string), args[i+1])
		j++
	}

	l.logger.Log(ctx, level, message, logArgs...)
}

func (l *SimpleLogger) Debug(args ...interface{}) {
	l.log(context.Background(), LevelDebug, args...)
}

func (l *SimpleLogger) Info(args ...interface{}) {
	l.log(context.Background(), LevelInfo, args...)
}

func (l *SimpleLogger) Warn(args ...interface{}) {
	l.log(context.Background(), LevelWarn, args...)
}

func (l *SimpleLogger) Error(args ...interface{}) {
	l.log(context.Background(), LevelError, args...)
}

func (l *SimpleLogger) Fatal(args ...interface{}) {
	l.log(context.Background(), LevelFatal, args...)
}

// NewSimpleLogger returns a new SimpleLogger.
func NewSimpleLogger(logger Logger) *SimpleLogger {
	return &SimpleLogger{logger: logger}
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
func Log(ctx context.Context, level Level, message string, fields ...Attr) {
	logger := FromContext(ctx)
	logger.Log(ctx, level, message, fields...)
}

// Debug logs the message with the debug level.
func Debug(ctx context.Context, message string, fields ...Attr) {
	Log(ctx, LevelDebug, message, fields...)
}

// Info logs the message with the info level.
func Info(ctx context.Context, message string, fields ...Attr) {
	Log(ctx, LevelInfo, message, fields...)
}

// Warn logs the message with the warn level.
func Warn(ctx context.Context, message string, fields ...Attr) {
	Log(ctx, LevelWarn, message, fields...)
}

// Error logs the message with the error level.
func Error(ctx context.Context, err error, fields ...Attr) {
	Log(ctx, LevelError, err.Error(), append(fields, slog.Any("error", err))...)
}

// Fatal logs the message with the fatal level.
func Fatal(ctx context.Context, message string, fields ...Attr) {
	Log(ctx, LevelFatal, message, fields...)
}

// Panic logs the message with the panic level.
func Panic(ctx context.Context, message string, fields ...Attr) {
	Log(ctx, LevelPanic, message, fields...)
}
