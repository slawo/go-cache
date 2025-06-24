package memory_test

import (
	"io"
	"testing"

	"github.com/slawo/go-cache/datastore/memory"
	"github.com/stretchr/testify/assert"
)

func TestIODataWrite(t *testing.T) {
	data := memory.NewIOData()

	// Test writing data
	n, err := data.WriteAt([]byte("Hello, World!"), 0)
	assert.NoError(t, err)
	assert.Equal(t, 13, n, "Expected to write 13 bytes")

	// Test reading back the data
	buf := make([]byte, 13)
	n, err = data.ReadAt(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, 13, n, "Expected to read 13 bytes")
	assert.Equal(t, "Hello, World!", string(buf[:n]), "Expected 'Hello, World!'")

	buf = make([]byte, 13)
	n, err = data.ReadAt(buf, 7)
	assert.EqualError(t, err, io.EOF.Error(), "Expected EOF error")
	assert.Equal(t, 6, n, "Expected to read 6 bytes")
	assert.Equal(t, "World!", string(buf[:n]), "Expected 'World!'")
}

func TestIODataReadAtNegativeIndex(t *testing.T) {
	data := memory.NewIOData()

	// Test reading at a negative index
	buf := make([]byte, 10)
	n, err := data.ReadAt(buf, -1)
	assert.EqualError(t, err, "read at negative offset")
	assert.Equal(t, 0, n, "Expected no bytes read")
}

func TestIODataWriteAtNegativeIndex(t *testing.T) {
	data := memory.NewIOData()

	// Test reading at a negative index
	buf := []byte("Hello, World!")
	n, err := data.WriteAt(buf, -1)
	assert.EqualError(t, err, "write at negative offset")
	assert.Equal(t, 0, n, "Expected no bytes read")
}
