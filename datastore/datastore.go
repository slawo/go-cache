package datastore

import (
	"context"
	"errors"

	cache "github.com/slawo/go-cache"
)

var (
	// ErrFileNotFound is returned when a file is not found in the data store.
	ErrFileNotFound = errors.New("file not found")
	// ErrInvalidFileID is returned when an invalid file ID is provided.
	ErrInvalidFileID = errors.New("invalid file ID")
)

//go:generate mockery --name DataIOProvider --output mocks
type DataIOProvider interface {
	// GetFileReader returns a reader for the file with the given ID.
	GetReaderAt(ctx context.Context, dataID string, position int64) (cache.ReadCloser, error)
	// GetFileWriter returns a writer for the file with the given ID.
	GetWriterAt(ctx context.Context, dataID string, position int64) (cache.WriteCloser, error)
}
