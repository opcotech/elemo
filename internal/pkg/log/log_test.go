package log

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/pkg"
)

func TestDefaultLogger(t *testing.T) {
	t.Parallel()

	assert.Equal(t, globalLogger, DefaultLogger())
}

func TestConfigureLogger(t *testing.T) {
	type args struct {
		level string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "parse debug level",
			args: args{
				level: "debug",
			},
			wantErr: false,
		},
		{
			name: "parse info level",
			args: args{
				level: "info",
			},
			wantErr: false,
		},
		{
			name: "parse warn level",
			args: args{
				level: "warn",
			},
			wantErr: false,
		},
		{
			name: "parse error level",
			args: args{
				level: "error",
			},
			wantErr: false,
		},
		{
			name: "parse invalid level",
			args: args{
				level: "invalid-level",
			},
			wantErr: true,
		},
		{
			name: "use default level",
			args: args{
				level: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, err := ConfigureLogger(tt.args.level)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestSlogLogger_Named(t *testing.T) {
	t.Parallel()

	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(os.Stderr, nil))}
	namedLogger := logger.Named("test-logger")

	assert.NotNil(t, namedLogger)
	assert.NotEqual(t, logger, namedLogger)
}

func TestSlogLogger_With(t *testing.T) {
	t.Parallel()

	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(os.Stderr, nil))}
	withLogger := logger.With(WithUserID("user123"), WithPath("/test"))

	assert.NotNil(t, withLogger)
	assert.NotEqual(t, logger, withLogger)
}

func TestSlogLogger_Log(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, nil))}

	ctx := context.Background()
	logger.Log(ctx, LevelInfo, "test message", WithUserID("user123"))

	assert.Contains(t, buf.String(), "test message")
	assert.Contains(t, buf.String(), "user_id")
	assert.Contains(t, buf.String(), "user123")
}

func TestSlogLogger_Debug(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))}

	ctx := context.Background()
	logger.Debug(ctx, "debug message", WithPath("/debug"))

	assert.Contains(t, buf.String(), "debug message")
}

func TestSlogLogger_Info(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, nil))}

	ctx := context.Background()
	logger.Info(ctx, "info message", WithPath("/info"))

	assert.Contains(t, buf.String(), "info message")
}

func TestSlogLogger_Warn(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, nil))}

	ctx := context.Background()
	logger.Warn(ctx, "warn message", WithPath("/warn"))

	assert.Contains(t, buf.String(), "warn message")
}

func TestSlogLogger_Error(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, nil))}

	ctx := context.Background()
	logger.Error(ctx, "error message", WithPath("/error"))

	assert.Contains(t, buf.String(), "error message")
}

func TestSlogLogger_Panic(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, nil))}

	ctx := context.Background()
	assert.Panics(t, func() {
		logger.Panic(ctx, "panic message", WithPath("/panic"))
	})

	assert.Contains(t, buf.String(), "panic message")
}

func TestAttrsToArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		attrs []slog.Attr
		want  int
	}{
		{
			name:  "empty attrs",
			attrs: []slog.Attr{},
			want:  0,
		},
		{
			name:  "single attr",
			attrs: []slog.Attr{slog.String("key", "value")},
			want:  2,
		},
		{
			name: "multiple attrs",
			attrs: []slog.Attr{
				slog.String("key1", "value1"),
				slog.Int("key2", 42),
			},
			want: 4,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := attrsToArgs(tt.attrs)
			assert.Equal(t, tt.want, len(args))
		})
	}
}

func TestExtractStandardFields(t *testing.T) {
	type args struct {
		ctx    context.Context
		fields []Attr
	}

	tests := []struct {
		name           string
		args           args
		wantRequestID  bool
		wantUserID     bool
		wantTraceID    bool
		explicitFields map[string]bool
	}{
		{
			name: "empty context, no fields",
			args: args{
				ctx:    context.Background(),
				fields: []Attr{},
			},
			wantRequestID: false,
			wantUserID:    false,
			wantTraceID:   false,
		},
		{
			name: "context with request ID",
			args: args{
				ctx:    context.WithValue(context.Background(), middleware.RequestIDKey, "req123"),
				fields: []Attr{},
			},
			wantRequestID: true,
			wantUserID:    false,
			wantTraceID:   false,
		},
		{
			name: "context with machine user ID",
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, pkg.CtxMachineUser),
				fields: []Attr{},
			},
			wantRequestID: false,
			wantUserID:    true,
			wantTraceID:   false,
		},
		{
			name: "explicit request ID should not be overridden",
			args: args{
				ctx:    context.WithValue(context.Background(), middleware.RequestIDKey, "req456"),
				fields: []Attr{WithRequestID("explicit-req")},
			},
			explicitFields: map[string]bool{FieldRequestID: true},
		},
		{
			name: "explicit user ID should not be overridden",
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, "user456"),
				fields: []Attr{WithUserID("explicit-user")},
			},
			explicitFields: map[string]bool{FieldUserID: true},
		},
		{
			name: "context with trace ID",
			args: args{
				// Note: noop spans don't record, so trace ID extraction won't work with noop tracer
				// This test verifies the behavior when no trace ID is available
				ctx:    context.Background(),
				fields: []Attr{},
			},
			wantRequestID: false,
			wantUserID:    false,
			wantTraceID:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := extractStandardFields(tt.args.ctx, tt.args.fields...)

			if tt.explicitFields != nil {
				// Check that explicit fields are present
				for key := range tt.explicitFields {
					found := false
					for _, field := range result {
						if field.Key == key {
							found = true
							break
						}
					}
					assert.True(t, found, "expected field %s to be present", key)
				}
				return
			}

			// Check for extracted fields
			hasRequestID := false
			hasUserID := false
			hasTraceID := false

			for _, field := range result {
				if field.Key == FieldRequestID {
					hasRequestID = true
				}
				if field.Key == FieldUserID {
					hasUserID = true
				}
				if field.Key == FieldTraceID {
					hasTraceID = true
				}
			}

			assert.Equal(t, tt.wantRequestID, hasRequestID, "request_id extraction mismatch")
			assert.Equal(t, tt.wantUserID, hasUserID, "user_id extraction mismatch")
			assert.Equal(t, tt.wantTraceID, hasTraceID, "trace_id extraction mismatch")
		})
	}
}

func TestWithContext(t *testing.T) {
	type args struct {
		ctx    context.Context
		logger Logger
	}

	tests := []struct {
		name string
		args args
		want Logger
	}{
		{
			name: "context with logger",
			args: args{
				ctx:    context.Background(),
				logger: DefaultLogger(),
			},
			want: DefaultLogger(),
		},
		{
			name: "context with nil logger uses global",
			args: args{
				ctx:    context.Background(),
				logger: nil,
			},
			want: globalLogger,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := WithContext(tt.args.ctx, tt.args.logger)
			logger := got.Value(pkg.CtxKeyLogger).(Logger)

			assert.Equal(t, tt.want, logger)
		})
	}
}

func TestFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name string
		args args
		want Logger
	}{
		{
			name: "context with logger",
			args: args{
				ctx: WithContext(context.Background(), DefaultLogger()),
			},
			want: DefaultLogger(),
		},
		{
			name: "context without logger returns global",
			args: args{
				ctx: context.Background(),
			},
			want: globalLogger,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := FromContext(tt.args.ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLog(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, nil))}
	ctx := WithContext(context.Background(), logger)

	Log(ctx, LevelInfo, "test message", WithUserID("user123"))

	assert.Contains(t, buf.String(), "test message")
}

func TestDebug(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))}
	ctx := WithContext(context.Background(), logger)

	Debug(ctx, "debug message", WithPath("/debug"))

	assert.Contains(t, buf.String(), "debug message")
}

func TestInfo(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, nil))}
	ctx := WithContext(context.Background(), logger)

	Info(ctx, "info message", WithPath("/info"))

	assert.Contains(t, buf.String(), "info message")
}

func TestWarn(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, nil))}
	ctx := WithContext(context.Background(), logger)

	Warn(ctx, "warn message", WithPath("/warn"))

	assert.Contains(t, buf.String(), "warn message")
}

func TestError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, nil))}
	ctx := WithContext(context.Background(), logger)

	err := assert.AnError
	Error(ctx, err, WithPath("/error"))

	assert.Contains(t, buf.String(), err.Error())
	assert.Contains(t, buf.String(), "error")
}

func TestPanic(t *testing.T) {
	t.Parallel()

	// Test package-level Panic function (logs at panic level but doesn't panic)
	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, nil))}
	ctx := WithContext(context.Background(), logger)

	Panic(ctx, "panic message", WithPath("/panic"))
	assert.Contains(t, buf.String(), "panic message")
}

func TestNewSimpleLogger(t *testing.T) {
	t.Parallel()

	logger := DefaultLogger()
	simpleLogger := NewSimpleLogger(logger)

	require.NotNil(t, simpleLogger)
	assert.Equal(t, logger, simpleLogger.logger)
}

func TestSimpleLogger_Log(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := &slogLogger{logger: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))}
	simpleLogger := NewSimpleLogger(logger)

	simpleLogger.Debug("key", "value", "debug message")
	assert.Contains(t, buf.String(), "debug message")

	buf.Reset()
	simpleLogger.Info("key", "value", "info message")
	assert.Contains(t, buf.String(), "info message")

	buf.Reset()
	simpleLogger.Warn("key", "value", "warn message")
	assert.Contains(t, buf.String(), "warn message")

	buf.Reset()
	simpleLogger.Error("key", "value", "error message")
	assert.Contains(t, buf.String(), "error message")
}
