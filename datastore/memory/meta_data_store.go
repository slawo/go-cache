package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/slawo/go-cache/datastore"
)

func NewMetaDataStore() *MetaDataStore {
	return &MetaDataStore{
		mapFileMeta:           make(map[string]*datastore.FileMeta),
		mapFileCompletionData: make(map[string]*datastore.FileCompletionData),
	}
}

type MetaDataStore struct {
	muFileMeta            sync.RWMutex
	muFileCompletionData  sync.RWMutex
	mapFileMeta           map[string]*datastore.FileMeta
	mapFileCompletionData map[string]*datastore.FileCompletionData
}

// GetFileMeta retrieves metadata for a file by its ID.
func (s *MetaDataStore) GetFileMeta(ctx context.Context, fileId string) (*datastore.FileMeta, error) {
	if strings.TrimSpace(fileId) == "" {
		return nil, fmt.Errorf("%w: empty file ID", datastore.ErrInvalidFileID)
	}
	s.muFileMeta.RLock()
	defer s.muFileMeta.RUnlock()
	if fileMeta, exists := s.mapFileMeta[fileId]; exists {
		metaCopy := *fileMeta // Create a copy to avoid external modifications
		return &metaCopy, nil
	}
	return nil, fmt.Errorf("%w: %s", datastore.ErrFileNotFound, fileId)
}

// SaveFileMeta saves metadata for a file.
func (s *MetaDataStore) SaveFileMeta(ctx context.Context, fileMeta *datastore.FileMeta) error {
	if fileMeta == nil {
		return errors.New("file metadata cannot be nil")
	} else if strings.TrimSpace(fileMeta.FileId) == "" {
		return fmt.Errorf("%w: empty file ID", datastore.ErrInvalidFileID)
	}
	s.muFileMeta.Lock()
	defer s.muFileMeta.Unlock()
	if s.mapFileMeta == nil {
		s.mapFileMeta = make(map[string]*datastore.FileMeta)
	}
	// Create a copy of the fileMeta to avoid external modifications
	metaCopy := *fileMeta
	s.mapFileMeta[fileMeta.FileId] = &metaCopy

	return nil
}

// GetFileCompletionData retrieves completion data for a file by its ID.
func (s *MetaDataStore) GetFileCompletionData(ctx context.Context, fileId string) (*datastore.FileCompletionData, error) {
	if strings.TrimSpace(fileId) == "" {
		return nil, fmt.Errorf("%w: empty file ID", datastore.ErrInvalidFileID)
	}
	s.muFileCompletionData.RLock()
	defer s.muFileCompletionData.RUnlock()
	if fileMeta, exists := s.mapFileCompletionData[fileId]; exists {
		completionCopy := *fileMeta // Create a copy to avoid external modifications
		return &completionCopy, nil
	}
	return nil, fmt.Errorf("%w: %s", datastore.ErrFileNotFound, fileId)
}

// SaveFileCompletionData saves completion data for a file.
func (s *MetaDataStore) SaveFileCompletionData(ctx context.Context, completionData *datastore.FileCompletionData) error {
	if completionData == nil {
		return errors.New("completion data cannot be nil")
	} else if strings.TrimSpace(completionData.FileId) == "" {
		return fmt.Errorf("%w: empty file ID", datastore.ErrInvalidFileID)
	}
	s.muFileCompletionData.Lock()
	defer s.muFileCompletionData.Unlock()
	if s.mapFileCompletionData == nil {
		s.mapFileCompletionData = make(map[string]*datastore.FileCompletionData)
	}
	// Create a copy of the fileMeta to avoid external modifications
	completionCopy := *completionData
	s.mapFileCompletionData[completionData.FileId] = &completionCopy

	return nil
}
