package memory_test

import (
	"context"
	"testing"

	"github.com/slawo/go-cache/datastore"
	"github.com/slawo/go-cache/datastore/memory"
	"github.com/slawo/go-cache/datastore/tests"
	"github.com/stretchr/testify/assert"
)

func TestNewIOProvider(t *testing.T) {
	p, err := memory.NewIOProvider()
	assert.NoError(t, err, "Expected no error when creating IOProvider")
	assert.NotNil(t, p, "Expected IOProvider to be non-nil")
}

func TestIOProviderGetReaderAt(t *testing.T) {
	p, err := memory.NewIOProvider()
	assert.NoError(t, err, "Expected no error when creating IOProvider")

	// Test with empty ID
	reader, err := p.GetReaderAt(t.Context(), "", 0)
	assert.Error(t, err, "Expected error for empty ID")
	assert.Nil(t, reader)

	// Test with valid ID
	reader, err = p.GetReaderAt(t.Context(), "test-file", 0)
	assert.NoError(t, err, "Expected no error for valid ID")
	assert.NotNil(t, reader, "Expected reader to be non-nil")
	assert.Equal(t, int64(0), reader.GetPosition(t.Context()), "Expected reader ID to match")

	err = reader.Close()
	assert.NoError(t, err, "Expected no error when closing reader")

	// Test with large offset
	reader, err = p.GetReaderAt(t.Context(), "test-file", 100000)
	assert.NoError(t, err, "Expected no error for valid ID")
	assert.NotNil(t, reader, "Expected reader to be non-nil")
	assert.Equal(t, int64(100000), reader.GetPosition(t.Context()), "Expected reader ID to match")

	err = reader.Close()
	assert.NoError(t, err, "Expected no error when closing reader")
}

func TestIOProviderGetWriterAt(t *testing.T) {
	p, err := memory.NewIOProvider()
	assert.NoError(t, err, "Expected no error when creating IOProvider")

	// Test with empty ID
	reader, err := p.GetWriterAt(t.Context(), "", 0)
	assert.Error(t, err, "Expected error for empty ID")
	assert.Nil(t, reader)

	// Test with valid ID
	reader, err = p.GetWriterAt(t.Context(), "test-file", 0)
	assert.NoError(t, err, "Expected no error for valid ID")
	assert.NotNil(t, reader, "Expected reader to be non-nil")
	assert.Equal(t, int64(0), reader.GetPosition(t.Context()), "Expected reader ID to match")

	err = reader.Close()
	assert.NoError(t, err, "Expected no error when closing reader")

	// Test with large offset
	reader, err = p.GetWriterAt(t.Context(), "test-file", 100000)
	assert.NoError(t, err, "Expected no error for valid ID")
	assert.NotNil(t, reader, "Expected reader to be non-nil")
	assert.Equal(t, int64(100000), reader.GetPosition(t.Context()), "Expected reader ID to match")

	err = reader.Close()
	assert.NoError(t, err, "Expected no error when closing reader")
}

func TestBaseIOProviderTests(t *testing.T) {
	tests.RunBaseIOProviderTests(t, tests.BaseIOProviderTestsOpts{
		NewIOProvider: func(ctx context.Context, t *testing.T) (datastore.DataIOProvider, error) {
			return memory.NewIOProvider()
		},
	})
	assert.NotNil(t, t, "TestBaseIOProviderTests should not be nil")
}
