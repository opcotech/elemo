package repository

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
)

var (
	ErrFileCreate = errors.New("failed to create file") // the file could not be created
	ErrFileDelete = errors.New("failed to delete file") // the file could not be deleted
	ErrFileGet    = errors.New("failed to get file")    // the file could not be retrieved
	ErrFileUpdate = errors.New("failed to update file") // the file could not be updated
)

type StaticFileRepository interface {
	// Create puts a new file in the static storage for the given path, reading
	// its data from the reader. It returns an error if the operation failed.
	Create(ctx context.Context, path string, data []byte) error
	// Get retrieves an object and writes its data to the designated location.
	// It returns an error if the operation failed.
	Get(ctx context.Context, path string) ([]byte, error)
	// Update replaces the file at the given path with the new data. It returns
	// an error if the operation failed.
	Update(ctx context.Context, path string, data []byte) error
	// Delete removes a file from the static storage, and returns an error if
	// the operation failed.
	Delete(ctx context.Context, path string) error
}

// StaticFileRepository represents an S3 static file storage.
type S3StaticFileRepository struct {
	*s3BaseRepository
}

func (r *S3StaticFileRepository) Create(ctx context.Context, path string, data []byte) error {
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
			ctx,
			ErrFileCreate.Error(),
			log.WithPath(path),
			log.WithAction(log.ActionFilePut),
			log.WithError(err),
		)
		return errors.Join(ErrFileCreate, err)
	}

	r.logger.Info(
		ctx,
		"new file created",
		log.WithPath(path),
		log.WithAction(log.ActionFilePut),
	)

	return nil
}

func (r *S3StaticFileRepository) Get(ctx context.Context, path string) ([]byte, error) {
	ctx, span := r.tracer.Start(ctx, "repository.pg.StaticFileRepository/Get")
	defer span.End()

	res, err := r.storage.client.GetObject(ctx, &awsS3.GetObjectInput{
		Bucket: &r.storage.bucket,
		Key:    &path,
	})
	if err != nil {
		if isNotFoundError(err) {
			return nil, errors.Join(ErrFileGet, ErrNotFound)
		}
		r.logger.Error(
			ctx,
			ErrFileGet.Error(),
			log.WithPath(path),
			log.WithAction(log.ActionFileGet),
			log.WithError(err),
		)
		return nil, errors.Join(ErrFileGet, err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		r.logger.Error(
			ctx,
			ErrFileGet.Error(),
			log.WithPath(path),
			log.WithAction(log.ActionFileGet),
			log.WithError(err),
		)
		return nil, errors.Join(ErrFileGet, err)
	}

	return body, nil
}

func (r *S3StaticFileRepository) Update(ctx context.Context, path string, data []byte) error {
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
			ctx,
			ErrFileUpdate.Error(),
			log.WithPath(path),
			log.WithAction(log.ActionFileUpdate),
			log.WithError(err),
		)
		return errors.Join(ErrFileUpdate, err)
	}

	r.logger.Info(
		ctx,
		"file updated",
		log.WithPath(path),
		log.WithAction(log.ActionFileUpdate),
	)

	return nil
}

func (r *S3StaticFileRepository) Delete(ctx context.Context, path string) error {
	ctx, span := r.tracer.Start(ctx, "repository.pg.StaticFileRepository/DeleteByWorkspaceID")
	defer span.End()

	_, err := r.storage.client.DeleteObject(ctx, &awsS3.DeleteObjectInput{
		Bucket: &r.storage.bucket,
		Key:    &path,
	})
	if err != nil {
		if isNotFoundError(err) {
			return errors.Join(ErrFileGet, ErrNotFound)
		}
		r.logger.Error(
			ctx,
			ErrFileDelete.Error(),
			log.WithPath(path),
			log.WithAction(log.ActionFileDelete),
			log.WithError(err),
		)
		return errors.Join(ErrFileDelete, err)
	}

	r.logger.Info(
		ctx,
		"file deleted",
		log.WithPath(path),
		log.WithAction(log.ActionFileDelete),
	)

	return nil
}

// NewStaticFileRepository creates a new StaticFileRepository to store files on
// Amazon S3.
func NewStaticFileRepository(opts ...S3RepositoryOption) (*S3StaticFileRepository, error) {
	baseRepo, err := newS3BaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &S3StaticFileRepository{
		s3BaseRepository: baseRepo,
	}, nil
}
