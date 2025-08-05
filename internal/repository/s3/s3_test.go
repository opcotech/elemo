package s3

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestNewClient(t *testing.T) {
	type args struct {
		ctx  context.Context
		conf *config.S3StorageConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "create new S3 client",
			args: args{
				ctx: context.Background(),
				conf: &config.S3StorageConfig{
					Region:          "us-east-1",
					AccessKeyID:     "test-access-key",
					SecretAccessKey: "test-secret-key",
					BaseEndpoint:    "http://localhost:9000",
				},
			},
		},
		{
			name: "create new S3 client with no config",
			args: args{
				ctx:  context.Background(),
				conf: nil,
			},
			wantErr: config.ErrNoConfig,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewClient(tt.args.ctx, tt.args.conf)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestWithStorageClient(t *testing.T) {
	type args struct {
		client Client
	}
	tests := []struct {
		name    string
		args    args
		want    Client
		wantErr error
	}{
		{
			name: "create new option with client",
			args: args{
				client: mock.NewS3Client(nil),
			},
			want: mock.NewS3Client(nil),
		},
		{
			name: "create new option with nil client",
			args: args{
				client: nil,
			},
			wantErr: repository.ErrNoClient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := new(Storage)
			err := WithStorageClient(tt.args.client)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, storage.client)
		})
	}
}

func TestWithStorageBucket(t *testing.T) {
	type args struct {
		bucket string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "create new option with bucket",
			args: args{
				bucket: "test-bucket",
			},
			want: "test-bucket",
		},
		{
			name: "create new option with empty bucket",
			args: args{
				bucket: "",
			},
			wantErr: repository.ErrNoBucket,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := new(Storage)
			err := WithStorageBucket(tt.args.bucket)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, storage.bucket)
		})
	}
}

func TestWithStorageLogger(t *testing.T) {
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
				logger: new(mock.Logger),
			},
			want: new(mock.Logger),
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
			storage := new(Storage)
			err := WithStorageLogger(tt.args.logger)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, storage.logger)
		})
	}
}

func TestWithStorageTracer(t *testing.T) {
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
				tracer: new(mock.Tracer),
			},
			want: new(mock.Tracer),
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
			storage := new(Storage)
			err := WithStorageTracer(tt.args.tracer)(storage)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, storage.tracer)
		})
	}
}

func TestNewStorage(t *testing.T) {
	type args struct {
		client Client
		bucket string
		logger log.Logger
		tracer tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    *Storage
		wantErr error
	}{
		{
			name: "create new storage",
			args: args{
				client: mock.NewS3Client(nil),
				bucket: "test-bucket",
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			want: &Storage{
				client: mock.NewS3Client(nil),
				bucket: "test-bucket",
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
		},
		{
			name: "create new storage with nil client",
			args: args{
				client: nil,
				bucket: "test-bucket",
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			wantErr: repository.ErrNoClient,
		},
		{
			name: "create new storage with empty bucket",
			args: args{
				client: mock.NewS3Client(nil),
				bucket: "",
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			wantErr: repository.ErrNoBucket,
		},
		{
			name: "create new storage with nil logger",
			args: args{
				client: mock.NewS3Client(nil),
				bucket: "test-bucket",
				logger: nil,
				tracer: new(mock.Tracer),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new storage with nil tracer",
			args: args{
				client: mock.NewS3Client(nil),
				bucket: "test-bucket",
				logger: new(mock.Logger),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage, err := NewStorage(
				WithStorageClient(tt.args.client),
				WithStorageBucket(tt.args.bucket),
				WithStorageLogger(tt.args.logger),
				WithStorageTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, storage)
		})
	}
}

func TestStorage_GetClient(t *testing.T) {
	t.Parallel()

	client := mock.NewS3Client(nil)

	storage := &Storage{
		client: client,
	}

	require.Equal(t, client, storage.GetClient())
}

func TestStorage_Ping(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type fields struct {
		client func(ctx context.Context, ctrl *gomock.Controller) Client
		bucket string
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "ping storage",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				client: func(ctx context.Context, ctrl *gomock.Controller) Client {
					c := mock.NewS3Client(ctrl)
					c.EXPECT().HeadBucket(ctx, &awsS3.HeadBucketInput{Bucket: aws.String("test-bucket")}, gomock.Any()).Return(&awsS3.HeadBucketOutput{}, nil)
					return c
				},
				bucket: "test-bucket",
			},
		},
		{
			name: "ping storage with error",
			args: args{
				ctx: context.Background(),
			},
			fields: fields{
				client: func(ctx context.Context, ctrl *gomock.Controller) Client {
					c := mock.NewS3Client(ctrl)
					c.EXPECT().HeadBucket(ctx, &awsS3.HeadBucketInput{Bucket: aws.String("test-bucket")}, gomock.Any()).Return(&awsS3.HeadBucketOutput{}, assert.AnError)
					return c
				},
				bucket: "test-bucket",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storage := &Storage{
				client: tt.fields.client(tt.args.ctx, ctrl),
				bucket: tt.fields.bucket,
			}
			err := storage.Ping(tt.args.ctx)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWithStorage(t *testing.T) {
	type args struct {
		storage *Storage
	}
	tests := []struct {
		name    string
		args    args
		want    *Storage
		wantErr error
	}{
		{
			name: "create new option with storage",
			args: args{
				storage: &Storage{
					client: mock.NewS3Client(nil),
					bucket: "test-bucket",
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
			},
			want: &Storage{
				client: mock.NewS3Client(nil),
				bucket: "test-bucket",
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
		},
		{
			name: "create new option with nil storage",
			args: args{
				storage: nil,
			},
			wantErr: repository.ErrNoDriver,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := new(baseRepository)
			err := WithStorage(tt.args.storage)(repo)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, repo.storage)
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
				logger: new(mock.Logger),
			},
			want: new(mock.Logger),
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
			repo := new(baseRepository)
			err := WithRepositoryLogger(tt.args.logger)(repo)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, repo.logger)
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
				tracer: new(mock.Tracer),
			},
			want: new(mock.Tracer),
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
			repo := new(baseRepository)
			err := WithRepositoryTracer(tt.args.tracer)(repo)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, repo.tracer)
		})
	}
}

func TestNewBaseRepository(t *testing.T) {
	type args struct {
		storage *Storage
		logger  log.Logger
		tracer  tracing.Tracer
	}
	tests := []struct {
		name    string
		args    args
		want    *baseRepository
		wantErr error
	}{
		{
			name: "create new base repository",
			args: args{
				storage: &Storage{
					client: mock.NewS3Client(nil),
					bucket: "test-bucket",
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
			want: &baseRepository{
				storage: &Storage{
					client: mock.NewS3Client(nil),
					bucket: "test-bucket",
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
				logger: new(mock.Logger),
				tracer: new(mock.Tracer),
			},
		},
		{
			name: "create new base repository with nil storage",
			args: args{
				storage: nil,
				logger:  new(mock.Logger),
				tracer:  new(mock.Tracer),
			},
			wantErr: repository.ErrNoDriver,
		},
		{
			name: "create new base repository with nil logger",
			args: args{
				storage: &Storage{
					client: mock.NewS3Client(nil),
					bucket: "test-bucket",
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
				logger: nil,
				tracer: new(mock.Tracer),
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "create new base repository with nil tracer",
			args: args{
				storage: &Storage{
					client: mock.NewS3Client(nil),
					bucket: "test-bucket",
					logger: new(mock.Logger),
					tracer: new(mock.Tracer),
				},
				logger: new(mock.Logger),
				tracer: nil,
			},
			wantErr: tracing.ErrNoTracer,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo, err := newBaseRepository(
				WithStorage(tt.args.storage),
				WithRepositoryLogger(tt.args.logger),
				WithRepositoryTracer(tt.args.tracer),
			)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, repo)
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

// Mock APIError for testing isNotFoundError
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
