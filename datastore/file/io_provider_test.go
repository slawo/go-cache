package file_test

import (
	"context"
	"os"
	"testing"

	"github.com/slawo/go-cache/datastore"
	"github.com/slawo/go-cache/datastore/file"
	"github.com/slawo/go-cache/datastore/tests"
	"github.com/stretchr/testify/assert"
)

func TestNewSimpleFileStoreFailsOnMissingFolder(t *testing.T) {
	store, err := file.NewIOProvider("")
	assert.EqualError(t, err, "file io provider: missing path")
	assert.Nil(t, store)
}

func TestNewSimpleFileStoreFailsOnNonExistingFolder(t *testing.T) {
	d := t.TempDir() + "/non-existing"
	store, err := file.NewIOProvider(d)
	assert.EqualError(t, err, "file io provider: path does not exist")
	assert.Nil(t, store)
}

func TestNewSimpleFileStoreFailsOn(t *testing.T) {
	d := t.TempDir() + "*@:"
	store, err := file.NewIOProvider(d)
	assert.EqualError(t, err, "file io provider: path does not exist")
	assert.Nil(t, store)
}

func TestNewSimpleFileStoreFailsOnPathIsFile(t *testing.T) {
	d := t.TempDir() + "/a-file.bin"
	os.WriteFile(d, []byte("test"), 0644)
	store, err := file.NewIOProvider(d)
	assert.EqualError(t, err, "file io provider: path is not a directory")
	assert.Nil(t, store)
}

func TestSimpleFileFailsOnMoreThanOneWriter(t *testing.T) {
	d := t.TempDir()
	store, err := file.NewIOProvider(d)
	assert.NoError(t, err)
	assert.NotNil(t, store)

	w, err := store.GetWriterAt(t.Context(), "newfile.bin", 0)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	w2, err := store.GetWriterAt(t.Context(), "newfile.bin", 0)
	assert.NoError(t, err)
	assert.NotNil(t, w2)
}

func TestRunBaseIOProviderTests(t *testing.T) {
	d := t.TempDir()
	newIOProvider := func(ctx context.Context, t *testing.T) (datastore.DataIOProvider, error) {
		store, err := file.NewIOProvider(d)
		assert.NoError(t, err)
		assert.NotNil(t, store)
		return store, nil
	}
	opts := tests.BaseIOProviderTestsOpts{
		NewIOProvider: newIOProvider,
	}
	tests.RunBaseIOProviderTests(t, opts)
}
