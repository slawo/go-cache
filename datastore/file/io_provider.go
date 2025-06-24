package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	cache "github.com/slawo/go-cache"
)

const (
	// DefaultFileStorePath is the default path for the file store.
	DefaultFileStorePath = "/var/lib/filestore"
	DefaultMaxReaders    = 16
)

func NewIOProvider(path string) (*IOProvider, error) {
	if path == "" {
		return nil, errors.New("file io provider: missing path")
	}
	f, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return nil, errors.New("file io provider: path does not exist")
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.New("file io provider: unable to access path")
	}
	if !f.IsDir() {
		return nil, errors.New("file io provider: path is not a directory")
	}
	return &IOProvider{
		path: path,
	}, nil
}

type IOProvider struct {
	path string
}

func (s *IOProvider) GetReaderAt(ctx context.Context, fileId string, position int64) (cache.ReadCloser, error) {
	p := path.Join(s.path, fileId)
	file, err := os.Open(p)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("file store: unable to open file: %w", err)
		}
	}
	if position > 0 {
		if _, err := file.Seek(position, io.SeekStart); err != nil {
			file.Close()
			return nil, fmt.Errorf("file store: unable to seek in file: %w", err)
		}
	}
	return &SimpleFileReader{
		file: file,
		p:    position,
	}, nil
}

func (s *IOProvider) GetWriterAt(ctx context.Context, fileId string, position int64) (cache.WriteCloser, error) {
	fileName := path.Join(s.path, fileId)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, errors.New("file store: unable to open file for writing: " + err.Error())
	}
	if position > 0 {
		if _, err := file.Seek(position, io.SeekStart); err != nil {
			file.Close()
			return nil, errors.New("file store: unable to seek in file: " + err.Error())
		}
	}
	return &SimpleFileWriter{
		file: file,
		p:    position,
	}, nil
}

type SimpleFileWriter struct {
	mu   sync.Mutex
	file *os.File
	p    int64
}

type SimpleFileReader struct {
	mu   sync.Mutex
	file *os.File
	p    int64
}

func (r *SimpleFileReader) GetPosition() int64 {
	return r.p
}

func (r *SimpleFileReader) Read(ctx context.Context, p []byte) (n int, err error) {
	if r.file == nil {
		return 0, errors.New("file writer: file is not open")
	}
	n, err = r.file.ReadAt(p, r.p)
	r.p += int64(n) // Update the position after reading
	return
}

func (r *SimpleFileReader) Close() error {
	if r.file != nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.file != nil {
			if err := r.file.Close(); err != nil {
				return fmt.Errorf("file store: unable to close file reader: %w", err)
			}
		}
	}
	return nil
}

func (w *SimpleFileWriter) GetPosition() int64 {
	return w.p
}

func (w *SimpleFileWriter) Write(ctx context.Context, p []byte) (n int, err error) {
	if w.file == nil {
		return 0, errors.New("file writer: file is not open")
	}
	n, err = w.file.Write(p)
	w.p += int64(n) // Update the position after reading
	return
}

func (w *SimpleFileWriter) Close() error {
	if w.file != nil {
		w.mu.Lock()
		defer w.mu.Unlock()
		if w.file != nil {
			if err := w.file.Close(); err != nil {
				return fmt.Errorf("file store: unable to close file writer: %w", err)
			}
		}
	}
	return nil
}
