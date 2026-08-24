package repository

import (
	"testing"

	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/pkg/log"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
)

func testS3Storage() *S3Storage {
	return &S3Storage{
		bucket: "test-bucket",
		logger: mocklog.NewMockLogger(nil),
		tracer: mocktrace.NewMockTracer(nil),
	}
}

func TestWithStorage(t *testing.T) {
	type args struct {
		storage *S3Storage
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "create new option with storage",
			args: args{
				storage: testS3Storage(),
			},
		},
		{
			name: "create new option with nil storage",
			args: args{
				storage: nil,
			},
			wantErr: ErrNoDriver,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := new(s3BaseRepository)
			err := WithS3Storage(tt.args.storage)(repo)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.args.storage, repo.storage)
			}
		})
	}
}

func TestWithRepositoryLogger(t *testing.T) {
	type args struct {
		logger log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    log.Logger
		wantErr error
	}{
		{
			name: "create new option with logger",
			args: args{
				logger: mocklog.NewMockLogger(nil),
			},
			want: mocklog.NewMockLogger(nil),
		},
		{
			name: "create new option with nil logger",
			args: args{
				logger: nil,
			},
			wantErr: log.ErrNoLogger,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := new(s3BaseRepository)
			err := WithS3RepositoryLogger(tt.args.logger)(repo)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.args.logger, repo.logger)
			}
		})
	}
}

func TestWithRepositoryTracer(t *testing.T) {
	type args struct {
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    tracing.Tracer
		wantErr error
	}{
		{
			name: "create new option with tracer",
			args: args{
				tracer: mocktrace.NewMockTracer(nil),
			},
			want: mocktrace.NewMockTracer(nil),
		},
		{
			name: "create new option with nil tracer",
			args: args{
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := new(s3BaseRepository)
			err := WithS3RepositoryTracer(tt.args.tracer)(repo)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.Equal(t, tt.args.tracer, repo.tracer)
			}
		})
	}
}

func TestNewBaseRepository(t *testing.T) {
	type args struct {
		storage *S3Storage
		logger  log.Logger
		tracer  tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "create new base repository",
			args: args{
				storage: testS3Storage(),
				logger:  mocklog.NewMockLogger(nil),
				tracer:  mocktrace.NewMockTracer(nil),
			},
		},
		{
			name: "create new base repository with nil storage",
			args: args{
				storage: nil,
				logger:  mocklog.NewMockLogger(nil),
				tracer:  mocktrace.NewMockTracer(nil),
			},
			wantErr: ErrNoDriver,
		},
		{
			name: "create new base repository with nil logger",
			args: args{
				storage: testS3Storage(),
				logger:  nil,
				tracer:  mocktrace.NewMockTracer(nil),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new base repository with nil tracer",
			args: args{
				storage: testS3Storage(),
				logger:  mocklog.NewMockLogger(nil),
				tracer:  nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, err := newS3BaseRepository(
				WithS3Storage(tt.args.storage),
				WithS3RepositoryLogger(tt.args.logger),
				WithS3RepositoryTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				require.NotNil(t, repo)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "not found error",
			err:  &mockAPIError{errorCode: "NoSuchKey"},
			want: true,
		},
		{
			name: "other error",
			err:  assert.AnError,
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isNotFoundError(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

type mockAPIError struct {
	errorCode string
}

func (m *mockAPIError) Error() string {
	return "mock API error"
}

func (m *mockAPIError) ErrorCode() string {
	return m.errorCode
}

func (m *mockAPIError) ErrorMessage() string {
	return "mock error message"
}

func (m *mockAPIError) ErrorFault() smithy.ErrorFault {
	return smithy.FaultUnknown
}
