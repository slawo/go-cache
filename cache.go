package data

import (
	"context"
	"errors"
)

var (
	ErrFileTooSmall = errors.New("file too small")
)

//go:generate mockery --name ReadCloser --output mocks
type ReadCloser interface {
	GetPosition() int64
	Read(ctx context.Context, p []byte) (n int, err error)
	Close() error
}

//go:generate mockery --name WriteCloser --output mocks
type WriteCloser interface {
	GetPosition() int64
	Write(ctx context.Context, p []byte) (n int, err error)
	Close() error
}

type Cache interface {
	ReadDataAt(context.Context, string) (ReadCloser, error)
}

// SourceRepository serves ReaderCloser instanced to access data from a source file.
// Usually I expect HTTP to be a primary source, but additional sources could come in
// handy (ie FTP)
//
//go:generate mockery --name SourceRepository
type SourceRepository interface {
	GetReaderAt(ctx context.Context, uri string, position int64) (ReadCloser, error)
}

// DataRepository is an interface for a data repository that provides methods to
// get read and write access to data at specific positions. This is typically
// used for data storage systems that allow random access to data, such as files
// or objct stores.
//
//go:generate mockery --name DataRepository --output mocks
type DataRepository interface {
	GetWriterAt(ctx context.Context, uri string, position int64) (WriteCloser, error)
	GetReaderAt(ctx context.Context, uri string, position int64) (ReadCloser, error)
}
