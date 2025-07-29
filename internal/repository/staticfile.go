package repository

import (
	"context"
)

// StaticFileRepository is a repository for managing static files.
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
