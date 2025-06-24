package memory

import (
	"context"
	"errors"
	"sync"

	cache "github.com/slawo/go-cache"
)

const (
	// DefaultFileStorePath is the default path for the file store.
	DefaultFileStorePath = "/var/lib/filestore"
	DefaultMaxReaders    = 16
)

func NewIOProvider() (*IOProvider, error) {
	return &IOProvider{
		data: make(map[string]*IOData),
	}, nil
}

type IOProvider struct {
	mu   sync.RWMutex
	data map[string]*IOData
}

func (s *IOProvider) GetReaderAt(ctx context.Context, ID string, position int64) (cache.ReadCloser, error) {
	if ID == "" {
		return nil, errors.New("io provider: missing ID")
	}
	return &IOReader{
		p: position,
		d: s.getIOData(ID),
	}, nil
}

func (s *IOProvider) GetWriterAt(ctx context.Context, ID string, position int64) (cache.WriteCloser, error) {
	if ID == "" {
		return nil, errors.New("io provider: missing ID")
	}
	return &IOWriter{
		p: position,
		d: s.getIOData(ID),
	}, nil
}

func (s *IOProvider) getIOData(fileId string) *IOData {
	s.mu.RLock()
	data, exists := s.data[fileId]
	s.mu.RUnlock()
	if !exists {
		s.mu.Lock()
		defer s.mu.Unlock()
		data = &IOData{
			d: []byte{},
		}
		s.data[fileId] = data
	}
	return data
}
