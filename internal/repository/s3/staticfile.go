package s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
)

// StaticFileRepository represents an S3 static file storage.
type StaticFileRepository struct {
	*baseRepository
}

func (r *StaticFileRepository) Create(ctx context.Context, path string, data []byte) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.StaticFileRepository/Create")
	defer span.End()

	_, err := r.storage.client.PutObject(ctx, &awsS3.PutObjectInput{
		Bucket:      &r.storage.bucket,
		Key:         &path,
		Body:        bytes.NewReader(data),
		ContentType: convert.ToPointer(http.DetectContentType(data)),
	})
	if err != nil {
		r.logger.Error(
			repository.ErrFileCreate.Error(),
			log.WithPath(path),
			log.WithAction(log.ActionFilePut),
			log.WithTraceID(tracing.GetTraceIDFromCtx(ctx)),
			log.WithError(err),
		)
		return errors.Join(repository.ErrFileCreate, err)
	}

	r.logger.Info(
		"new file created",
		log.WithPath(path),
		log.WithAction(log.ActionFilePut),
		log.WithTraceID(tracing.GetTraceIDFromCtx(ctx)),
	)

	return nil
}

func (r *StaticFileRepository) Get(ctx context.Context, path string) ([]byte, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.StaticFileRepository/Get")
	defer span.End()

	res, err := r.storage.client.(*awsS3.Client).GetObject(ctx, &awsS3.GetObjectInput{
		Bucket: &r.storage.bucket,
		Key:    &path,
	})
	if err != nil {
		if isNotFoundError(err) {
			return nil, errors.Join(repository.ErrFileGet, repository.ErrNotFound)
		}
		r.logger.Error(
			repository.ErrFileGet.Error(),
			log.WithPath(path),
			log.WithAction(log.ActionFileGet),
			log.WithTraceID(tracing.GetTraceIDFromCtx(ctx)),
			log.WithError(err),
		)
		return nil, errors.Join(repository.ErrFileGet, err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		r.logger.Error(
			repository.ErrFileGet.Error(),
			log.WithPath(path),
			log.WithAction(log.ActionFileGet),
			log.WithTraceID(tracing.GetTraceIDFromCtx(ctx)),
			log.WithError(err),
		)
		return nil, errors.Join(repository.ErrFileGet, err)
	}

	return body, nil
}

func (r *StaticFileRepository) Update(ctx context.Context, path string, data []byte) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.StaticFileRepository/Update")
	defer span.End()

	_, err := r.storage.client.PutObject(ctx, &awsS3.PutObjectInput{
		Bucket:      &r.storage.bucket,
		Key:         &path,
		Body:        bytes.NewReader(data),
		ContentType: convert.ToPointer(http.DetectContentType(data)),
	})
	if err != nil {
		r.logger.Error(
			repository.ErrFileUpdate.Error(),
			log.WithPath(path),
			log.WithAction(log.ActionFileUpdate),
			log.WithTraceID(tracing.GetTraceIDFromCtx(ctx)),
			log.WithError(err),
		)
		return errors.Join(repository.ErrFileUpdate, err)
	}

	r.logger.Info(
		"file updated",
		log.WithPath(path),
		log.WithAction(log.ActionFileUpdate),
		log.WithTraceID(tracing.GetTraceIDFromCtx(ctx)),
	)

	return nil
}

func (r *StaticFileRepository) Delete(ctx context.Context, path string) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.StaticFileRepository/DeleteByWorkspaceID")
	defer span.End()

	_, err := r.storage.client.DeleteObject(ctx, &awsS3.DeleteObjectInput{
		Bucket: &r.storage.bucket,
		Key:    &path,
	})
	if err != nil {
		if isNotFoundError(err) {
			return errors.Join(repository.ErrFileGet, repository.ErrNotFound)
		}
		r.logger.Error(
			repository.ErrFileDelete.Error(),
			log.WithPath(path),
			log.WithAction(log.ActionFileDelete),
			log.WithTraceID(tracing.GetTraceIDFromCtx(ctx)),
			log.WithError(err),
		)
		return errors.Join(repository.ErrFileDelete, err)
	}

	r.logger.Info(
		"file deleted",
		log.WithPath(path),
		log.WithAction(log.ActionFileDelete),
		log.WithTraceID(tracing.GetTraceIDFromCtx(ctx)),
	)

	return nil
}

// NewStaticFileRepository creates a new StaticFileRepository to store files on
// Amazon S3.
func NewStaticFileRepository(opts ...RepositoryOption) (*StaticFileRepository, error) {
	baseRepo, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &StaticFileRepository{
		baseRepository: baseRepo,
	}, nil
}
