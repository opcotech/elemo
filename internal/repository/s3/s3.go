package s3

import (
	"context"
	"errors"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	awsCredentials "github.com/aws/aws-sdk-go-v2/credentials"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
)

type Client interface {
	CreateBucket(ctx context.Context, params *awsS3.CreateBucketInput, optFns ...func(*awsS3.Options)) (*awsS3.CreateBucketOutput, error)
	HeadBucket(ctx context.Context, params *awsS3.HeadBucketInput, optFns ...func(*awsS3.Options)) (*awsS3.HeadBucketOutput, error)
	DeleteBucket(ctx context.Context, params *awsS3.DeleteBucketInput, optFns ...func(*awsS3.Options)) (*awsS3.DeleteBucketOutput, error)
	PutObject(ctx context.Context, params *awsS3.PutObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.PutObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *awsS3.ListObjectsV2Input, optFns ...func(*awsS3.Options)) (*awsS3.ListObjectsV2Output, error)
	DeleteObject(ctx context.Context, params *awsS3.DeleteObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.DeleteObjectOutput, error)
}

// NewClient creates a new S3 storage client.
func NewClient(ctx context.Context, conf *config.S3StorageConfig) (Client, error) {
	if conf == nil {
		return nil, config.ErrNoConfig
	}

	sdkConfig, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.Join(repository.ErrInvalidConfig, err)
	}

	if conf.BaseEndpoint != "" {
		sdkConfig.BaseEndpoint = &conf.BaseEndpoint
	}

	return awsS3.NewFromConfig(sdkConfig, func(o *awsS3.Options) {
		o.UsePathStyle = true
		o.Region = conf.Region
		o.Credentials = awsCredentials.NewStaticCredentialsProvider(
			conf.AccessKeyID,
			conf.SecretAccessKey,
			"",
		)
	}), nil
}

// StorageOption configures a Postgres database.
type StorageOption func(*Storage) error

// WithStorageClient sets the S3 client on the Storage.
func WithStorageClient(client Client) StorageOption {
	return func(storage *Storage) error {
		if client == nil {
			return repository.ErrNoClient
		}

		storage.client = client
		return nil
	}
}

// WithStorageBucket sets the S3 bucket on the Storage.
func WithStorageBucket(bucket string) StorageOption {
	return func(storage *Storage) error {
		if bucket == "" {
			return repository.ErrNoBucket
		}

		storage.bucket = bucket
		return nil
	}
}

// WithStorageLogger sets the logger for a Neo4j database.
func WithStorageLogger(logger log.Logger) StorageOption {
	return func(storage *Storage) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		storage.logger = logger
		return nil
	}
}

// WithStorageTracer sets the tracer for a Neo4j database.
func WithStorageTracer(tracer tracing.Tracer) StorageOption {
	return func(storage *Storage) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		storage.tracer = tracer
		return nil
	}
}

// Storage defines the interface for S3 storage.
type Storage struct {
	client Client         `validate:"required"`
	bucket string         `validate:"required"`
	logger log.Logger     `validate:"required"`
	tracer tracing.Tracer `validate:"required"`
}

// Ping checks the database connection.
func (s *Storage) Ping(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &awsS3.HeadBucketInput{Bucket: &s.bucket})
	return err
}

// GetClient returns the S3 client.
func (s *Storage) GetClient() Client {
	return s.client
}

// NewStorage creates a new Postgres database.
func NewStorage(opts ...StorageOption) (*Storage, error) {
	storage := &Storage{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(storage); err != nil {
			return nil, err
		}
	}

	return storage, nil
}

type RepositoryOption func(*baseRepository) error

// WithStorage sets the baseRepository for a baseRepository.
func WithStorage(storage *Storage) RepositoryOption {
	return func(r *baseRepository) error {
		if storage == nil {
			return repository.ErrNoDriver
		}
		r.storage = storage

		return nil
	}
}

// WithRepositoryLogger sets the logger for a baseRepository.
func WithRepositoryLogger(logger log.Logger) RepositoryOption {
	return func(r *baseRepository) error {
		if logger == nil {
			return log.ErrNoLogger
		}
		r.logger = logger

		return nil
	}
}

// WithRepositoryTracer sets the tracer for a baseRepository.
func WithRepositoryTracer(tracer tracing.Tracer) RepositoryOption {
	return func(r *baseRepository) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}
		r.tracer = tracer

		return nil
	}
}

// baseRepository represents an S3 static file storage.
type baseRepository struct {
	storage *Storage
	logger  log.Logger
	tracer  tracing.Tracer
}

// newBaseRepository creates a new baseRepository.
func newBaseRepository(opts ...RepositoryOption) (*baseRepository, error) {
	r := &baseRepository{
		logger: log.DefaultLogger(),
		tracer: tracing.NoopTracer(),
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func isNotFoundError(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchKey"
}
