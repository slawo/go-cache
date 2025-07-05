package memory_test

import (
	"testing"

	"github.com/slawo/go-cache/datastore"
	"github.com/slawo/go-cache/datastore/memory"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryMetaDataStore(t *testing.T) {
	store := memory.NewMetaDataStore()
	assert.NotNil(t, store)
}

func TestMemoryMetaDataStoreSaveFileMetaErrorOnNilData(t *testing.T) {
	store := memory.NewMetaDataStore()
	err := store.SaveFileMeta(t.Context(), nil)
	assert.EqualError(t, err, "file metadata cannot be nil")
}

func TestMemoryMetaDataStoreSaveFileMetaErrorOnInvalidID(t *testing.T) {
	store := memory.NewMetaDataStore()
	err := store.SaveFileMeta(t.Context(), &datastore.FileMeta{
		FileId:   " 	",
		FileSize: 1024,
		Checksum: "abc123",
	})
	assert.ErrorIs(t, err, datastore.ErrInvalidFileID)
}

func TestMemoryMetaDataStoreGetFileMetaErrorOnInvalidID(t *testing.T) {
	store := memory.NewMetaDataStore()
	m, err := store.GetFileMeta(t.Context(), " 	")
	assert.ErrorIs(t, err, datastore.ErrInvalidFileID)
	assert.Nil(t, m)
}

func TestMemoryMetaDataStoreGetFileMetaErrorOnMissingMeta(t *testing.T) {
	store := memory.NewMetaDataStore()
	m, err := store.GetFileMeta(t.Context(), "nonexistent-file-id")
	assert.ErrorIs(t, err, datastore.ErrFileNotFound)
	assert.Nil(t, m)
}

func TestMemoryMetaDataStoreGetFileMetaSuccess(t *testing.T) {
	store := &memory.MetaDataStore{}
	fileId := "test-file-id"
	completionData := &datastore.FileMeta{
		FileId:   fileId,
		FileSize: 1024,
		Checksum: "abc123",
	}
	err := store.SaveFileMeta(t.Context(), completionData)
	assert.NoError(t, err)

	m, err := store.GetFileMeta(t.Context(), fileId)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, fileId, m.FileId)
	assert.Equal(t, int64(1024), m.FileSize)
	assert.Equal(t, "abc123", m.Checksum)
}

func TestMemoryMetaDataStoreSaveFileCompletionDataOnNilData(t *testing.T) {
	store := memory.NewMetaDataStore()
	err := store.SaveFileCompletionData(t.Context(), nil)
	assert.EqualError(t, err, "completion data cannot be nil")
}

func TestMemoryMetaDataStoreSaveFileCompletionDataOnInvalidID(t *testing.T) {
	store := memory.NewMetaDataStore()
	err := store.SaveFileCompletionData(t.Context(), &datastore.FileCompletionData{
		FileId:         " 	",
		PartSize:       1024,
		PartsCompleted: 5,
	})
	assert.ErrorIs(t, err, datastore.ErrInvalidFileID)
}

func TestMemoryMetaDataStoreGetFileCompletionDataErrorOnInvalidID(t *testing.T) {
	store := memory.NewMetaDataStore()
	m, err := store.GetFileCompletionData(t.Context(), " 	")
	assert.ErrorIs(t, err, datastore.ErrInvalidFileID)
	assert.Nil(t, m)
}

func TestMemoryMetaDataStoreGetFileCompletionDataErrorOnMissingMeta(t *testing.T) {
	store := memory.NewMetaDataStore()
	m, err := store.GetFileCompletionData(t.Context(), "nonexistent-file-id")
	assert.ErrorIs(t, err, datastore.ErrFileNotFound)
	assert.Nil(t, m)
}

func TestMemoryMetaDataStoreGetFileCompletionDataSuccess(t *testing.T) {
	store := &memory.MetaDataStore{}
	fileId := "test-file-id"
	completionData := &datastore.FileCompletionData{
		FileId:         fileId,
		PartSize:       1024,
		PartsCompleted: 5,
	}
	err := store.SaveFileCompletionData(t.Context(), completionData)
	assert.NoError(t, err)

	m, err := store.GetFileCompletionData(t.Context(), fileId)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, fileId, m.FileId)
	assert.Equal(t, int64(1024), m.PartSize)
	assert.Equal(t, 5, m.PartsCompleted)
}
