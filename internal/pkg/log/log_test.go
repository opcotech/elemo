package log

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/opcotech/elemo/internal/testutil/mock"
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
			name: "parse level",
			args: args{
				level: "debug",
			},
		},
		{
			name: "parse level error",
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
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ConfigureLogger(tt.args.level)
			assert.Equal(t, tt.wantErr, err != nil)
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
	}{
		{
			name: "context with logger",
			args: args{
				ctx:    context.Background(),
				logger: new(mock.Logger),
			},
		},
		{
			name: "context with global logger",
			args: args{
				ctx:    context.Background(),
				logger: nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := WithContext(tt.args.ctx, tt.args.logger)

			if tt.args.logger == nil {
				assert.Equal(t, globalLogger, got.Value(CtxKeyLogger).(Logger))
			} else {
				assert.Equal(t, tt.args.logger, got.Value(CtxKeyLogger).(Logger))
			}
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
				ctx: context.WithValue(context.Background(), CtxKeyLogger, new(mock.Logger)),
			},
			want: new(mock.Logger),
		},
		{
			name: "context with global logger",
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

			assert.Equal(t, tt.want, FromContext(tt.args.ctx))
		})
	}
}

func TestLog(t *testing.T) {
	type args struct {
		ctx     func(logger *mock.Logger) context.Context
		level   zapcore.Level
		message string
		fields  []zap.Field
	}

	tests := []struct {
		name   string
		args   args
		logger func(zapcore.Level, string, []zap.Field) *mock.Logger
	}{
		{
			name: "log message",
			args: args{
				ctx: func(logger *mock.Logger) context.Context {
					return context.WithValue(context.Background(), CtxKeyLogger, logger)
				},
				level:   zapcore.DebugLevel,
				message: "test",
				fields:  []zap.Field{},
			},
			logger: func(level zapcore.Level, message string, fields []zap.Field) *mock.Logger {
				logger := new(mock.Logger)
				logger.On("Log", level, message, fields).Return()
				return logger
			},
		},
		{
			name: "log with extra fields",
			args: args{
				ctx: func(logger *mock.Logger) context.Context {
					ctx := context.Background()
					return context.WithValue(ctx, CtxKeyLogger, logger)
				},
				level:   zapcore.DebugLevel,
				message: "test",
				fields:  []zap.Field{zap.String("test", "test")},
			},
			logger: func(level zapcore.Level, message string, fields []zap.Field) *mock.Logger {
				logger := new(mock.Logger)
				logger.On("Log", level, message, fields).Return()
				return logger
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := tt.logger(tt.args.level, tt.args.message, tt.args.fields)
			ctx := tt.args.ctx(logger)

			Log(ctx, tt.args.level, tt.args.message, tt.args.fields...)

			logger.AssertExpectations(t)
		})
	}
}

func TestDebug(t *testing.T) {
	type args struct {
		ctx     func(logger *mock.Logger) context.Context
		message string
		fields  []zap.Field
	}

	tests := []struct {
		name   string
		args   args
		logger func(string, []zap.Field) *mock.Logger
	}{
		{
			name: "log debug message",
			args: args{
				ctx: func(logger *mock.Logger) context.Context {
					return context.WithValue(context.Background(), CtxKeyLogger, logger)
				},
				message: "test",
				fields:  []zap.Field{zap.String("test", "test")},
			},
			logger: func(message string, fields []zap.Field) *mock.Logger {
				logger := new(mock.Logger)
				logger.On("Log", zapcore.DebugLevel, message, fields).Return()
				return logger
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := tt.logger(tt.args.message, tt.args.fields)
			ctx := tt.args.ctx(logger)

			Debug(ctx, tt.args.message, tt.args.fields...)

			logger.AssertExpectations(t)
		})
	}
}

func TestInfo(t *testing.T) {
	type args struct {
		ctx     func(logger *mock.Logger) context.Context
		message string
		fields  []zap.Field
	}

	tests := []struct {
		name   string
		args   args
		logger func(string, []zap.Field) *mock.Logger
	}{
		{
			name: "log info message",
			args: args{
				ctx: func(logger *mock.Logger) context.Context {
					return context.WithValue(context.Background(), CtxKeyLogger, logger)
				},
				message: "test",
				fields:  []zap.Field{zap.String("test", "test")},
			},
			logger: func(message string, fields []zap.Field) *mock.Logger {
				logger := new(mock.Logger)
				logger.On("Log", zapcore.InfoLevel, message, fields).Return()
				return logger
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := tt.logger(tt.args.message, tt.args.fields)
			ctx := tt.args.ctx(logger)

			Info(ctx, tt.args.message, tt.args.fields...)

			logger.AssertExpectations(t)
		})
	}
}

func TestWarn(t *testing.T) {
	type args struct {
		ctx     func(logger *mock.Logger) context.Context
		message string
		fields  []zap.Field
	}

	tests := []struct {
		name   string
		args   args
		logger func(string, []zap.Field) *mock.Logger
	}{
		{
			name: "log warn message",
			args: args{
				ctx: func(logger *mock.Logger) context.Context {
					return context.WithValue(context.Background(), CtxKeyLogger, logger)
				},
				message: "test",
				fields:  []zap.Field{zap.String("test", "test")},
			},
			logger: func(message string, fields []zap.Field) *mock.Logger {
				logger := new(mock.Logger)
				logger.On("Log", zapcore.WarnLevel, message, fields).Return()
				return logger
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := tt.logger(tt.args.message, tt.args.fields)
			ctx := tt.args.ctx(logger)

			Warn(ctx, tt.args.message, tt.args.fields...)

			logger.AssertExpectations(t)
		})
	}
}

func TestError(t *testing.T) {
	type args struct {
		ctx    func(logger *mock.Logger) context.Context
		err    error
		fields []zap.Field
	}

	tests := []struct {
		name   string
		args   args
		logger func(error, []zap.Field) *mock.Logger
	}{
		{
			name: "log error message",
			args: args{
				ctx: func(logger *mock.Logger) context.Context {
					return context.WithValue(context.Background(), CtxKeyLogger, logger)
				},
				err:    fmt.Errorf("test"),
				fields: []zap.Field{zap.String("test", "test")},
			},
			logger: func(err error, fields []zap.Field) *mock.Logger {
				logger := new(mock.Logger)
				logger.On("Log", zapcore.ErrorLevel, err.Error(), append(fields, WithError(err))).Return()
				return logger
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := tt.logger(tt.args.err, tt.args.fields)
			ctx := tt.args.ctx(logger)

			Error(ctx, tt.args.err, tt.args.fields...)

			logger.AssertExpectations(t)
		})
	}
}

func TestFatal(t *testing.T) {
	type args struct {
		ctx     func(logger *mock.Logger) context.Context
		message string
		fields  []zap.Field
	}

	tests := []struct {
		name   string
		args   args
		logger func(string, []zap.Field) *mock.Logger
	}{
		{
			name: "log fatal message",
			args: args{
				ctx: func(logger *mock.Logger) context.Context {
					return context.WithValue(context.Background(), CtxKeyLogger, logger)
				},
				message: "test",
				fields:  []zap.Field{zap.String("test", "test")},
			},
			logger: func(message string, fields []zap.Field) *mock.Logger {
				logger := new(mock.Logger)
				logger.On("Log", zapcore.FatalLevel, message, fields).Return()
				return logger
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := tt.logger(tt.args.message, tt.args.fields)
			ctx := tt.args.ctx(logger)

			Fatal(ctx, tt.args.message, tt.args.fields...)

			logger.AssertExpectations(t)
		})
	}
}

func TestPanic(t *testing.T) {
	type args struct {
		ctx     func(logger *mock.Logger) context.Context
		message string
		fields  []zap.Field
	}

	tests := []struct {
		name   string
		args   args
		logger func(string, []zap.Field) *mock.Logger
	}{
		{
			name: "log panic message",
			args: args{
				ctx: func(logger *mock.Logger) context.Context {
					return context.WithValue(context.Background(), CtxKeyLogger, logger)
				},
				message: "test",
				fields:  []zap.Field{zap.String("test", "test")},
			},
			logger: func(message string, fields []zap.Field) *mock.Logger {
				logger := new(mock.Logger)
				logger.On("Log", zapcore.PanicLevel, message, fields).Return()
				return logger
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := tt.logger(tt.args.message, tt.args.fields)
			ctx := tt.args.ctx(logger)

			Panic(ctx, tt.args.message, tt.args.fields...)

			logger.AssertExpectations(t)
		})
	}
}
