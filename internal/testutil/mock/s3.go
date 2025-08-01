package mock

import (
	"context"

	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/mock"
)

// S3ClientOld is a mock implementation of the S3 Client interface
type S3ClientOld struct {
	mock.Mock
}

func (m *S3ClientOld) CreateBucket(ctx context.Context, params *awsS3.CreateBucketInput, optFns ...func(*awsS3.Options)) (*awsS3.CreateBucketOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsS3.CreateBucketOutput), args.Error(1)
}

func (m *S3ClientOld) HeadBucket(ctx context.Context, params *awsS3.HeadBucketInput, optFns ...func(*awsS3.Options)) (*awsS3.HeadBucketOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsS3.HeadBucketOutput), args.Error(1)
}

func (m *S3ClientOld) DeleteBucket(ctx context.Context, params *awsS3.DeleteBucketInput, optFns ...func(*awsS3.Options)) (*awsS3.DeleteBucketOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsS3.DeleteBucketOutput), args.Error(1)
}

func (m *S3ClientOld) PutObject(ctx context.Context, params *awsS3.PutObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.PutObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsS3.PutObjectOutput), args.Error(1)
}

func (m *S3ClientOld) ListObjectsV2(ctx context.Context, params *awsS3.ListObjectsV2Input, optFns ...func(*awsS3.Options)) (*awsS3.ListObjectsV2Output, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsS3.ListObjectsV2Output), args.Error(1)
}

func (m *S3ClientOld) DeleteObject(ctx context.Context, params *awsS3.DeleteObjectInput, optFns ...func(*awsS3.Options)) (*awsS3.DeleteObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsS3.DeleteObjectOutput), args.Error(1)
}
