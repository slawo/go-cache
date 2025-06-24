package tests

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/slawo/go-cache/datastore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type BaseIOProviderTestsOpts struct {
	NewIOProvider func(ctx context.Context, t *testing.T) (datastore.DataIOProvider, error)
}

func RunBaseIOProviderTests(t *testing.T, opts BaseIOProviderTestsOpts) {
	require.NotNil(t, opts.NewIOProvider, "NewIOProvider function must be provided")
	t.Run("NewProvider", func(t *testing.T) {
		p, err := opts.NewIOProvider(context.Background(), t)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})
	t.Run("GetRwiter", func(t *testing.T) {
		t.Parallel()
		p, err := opts.NewIOProvider(context.Background(), t)
		require.NoError(t, err)
		require.NotNil(t, p)
		fn := generateFileName()
		writer, err := p.GetWriterAt(context.Background(), fn, 0)
		assert.NoError(t, err)
		assert.NotNil(t, writer)
		assert.NoError(t, writer.Close())
	})
	t.Run("GetRwiterAtPosition", func(t *testing.T) {
		t.Parallel()
		p, err := opts.NewIOProvider(context.Background(), t)
		require.NoError(t, err)
		require.NotNil(t, p)
		fn := generateFileName()
		writer, err := p.GetWriterAt(context.Background(), fn, 1000000000) // 1GB offset
		assert.NoError(t, err)
		assert.NotNil(t, writer)
		assert.NoError(t, writer.Close())
	})
	t.Run("GetRwiterAtPositionSuccesfullyWrites", func(t *testing.T) {
		t.Parallel()
		p, err := opts.NewIOProvider(context.Background(), t)
		require.NoError(t, err)
		require.NotNil(t, p)
		fn := generateFileName()
		writer, err := p.GetWriterAt(context.Background(), fn, 0)
		require.NoError(t, err)
		require.NotNil(t, writer)
		t.Cleanup(func() {
			assert.NoError(t, writer.Close())
			// Clean up the file after the test
		})
		data := []byte("test data for writing to file")
		n, err := writer.Write(t.Context(), data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
	})
	t.Run("GetRwiterMultiCloseErrors", func(t *testing.T) {
		t.Parallel()
		p, err := opts.NewIOProvider(context.Background(), t)
		require.NoError(t, err)
		require.NotNil(t, p)
		fn := generateFileName()
		writer, err := p.GetWriterAt(context.Background(), fn, 0)
		assert.NoError(t, err)
		assert.NotNil(t, writer)
		assert.NoError(t, writer.Close())
		assert.Error(t, writer.Close())
	})
	t.Run("GetReader", func(t *testing.T) {
		t.Parallel()
		p, err := opts.NewIOProvider(context.Background(), t)
		require.NoError(t, err)
		require.NotNil(t, p)
		fn := generateFileName()
		reader, err := p.GetReaderAt(context.Background(), fn, 0)
		assert.NoError(t, err)
		assert.NotNil(t, reader)
		assert.Equal(t, int64(0), reader.GetPosition())
		assert.NoError(t, reader.Close())
	})
	t.Run("GetReaderAtPosition", func(t *testing.T) {
		t.Parallel()
		p, err := opts.NewIOProvider(context.Background(), t)
		require.NoError(t, err)
		require.NotNil(t, p)
		fn := generateFileName()
		reader, err := p.GetWriterAt(context.Background(), fn, 1000000000) // 1GB offset
		assert.NoError(t, err)
		assert.NotNil(t, reader)
		assert.Equal(t, int64(1000000000), reader.GetPosition())
		assert.NoError(t, reader.Close())
	})
	t.Run("GetReaderMultiCloseError", func(t *testing.T) {
		t.Parallel()
		p, err := opts.NewIOProvider(context.Background(), t)
		require.NoError(t, err)
		require.NotNil(t, p)
		fn := generateFileName()
		reader, err := p.GetWriterAt(context.Background(), fn, 0)
		assert.NoError(t, err)
		assert.NotNil(t, reader)
		assert.Equal(t, int64(0), reader.GetPosition())
		assert.NoError(t, reader.Close())
		assert.Error(t, reader.Close())
	})
	t.Run("GetReaderAtPositionSuccesfullyReadsWrittenData", func(t *testing.T) {
		t.Parallel()
		p, err := opts.NewIOProvider(context.Background(), t)
		require.NoError(t, err)
		require.NotNil(t, p)
		fn := generateFileName()
		writer, err := p.GetWriterAt(context.Background(), fn, 0)
		require.NoError(t, err)
		require.NotNil(t, writer)
		t.Cleanup(func() {
			assert.NoError(t, writer.Close())
		})

		data := []byte("test data for writing to file")

		t.Run("WriteData", func(t *testing.T) {
			assert.Equal(t, int64(0), writer.GetPosition(), "Initial position should be 0")
			n, err := writer.Write(t.Context(), data)
			require.NoError(t, err)
			require.Equal(t, len(data), n)
			require.Equal(t, int64(len(data)), writer.GetPosition())
		})
		reader, err := p.GetReaderAt(context.Background(), fn, 0)
		assert.NoError(t, err)
		require.NotNil(t, reader)
		t.Cleanup(func() {
			assert.NoError(t, reader.Close())
		})
		t.Run("ReadData", func(t *testing.T) {
			assert.Equal(t, int64(0), reader.GetPosition(), "Initial position should be 0")
			readData := make([]byte, 1024)
			n, err := reader.Read(t.Context(), readData)
			assert.EqualError(t, err, io.EOF.Error())
			assert.Equal(t, len(data), n)
			assert.Equal(t, data, readData[:n])
			assert.Equal(t, int64(len(data)), reader.GetPosition(), "Position should be equal to data length")
		})
		t.Run("ReadDataEOF", func(t *testing.T) {
			readData := make([]byte, 1024)
			n, err := reader.Read(t.Context(), readData)
			assert.EqualError(t, err, io.EOF.Error())
			assert.Equal(t, 0, n, "No more data should be read after EOF")
			assert.Equal(t, int64(len(data)), reader.GetPosition(), "Position should remain at the end of the data")
		})
		// Now let's read from a specific offset

		offset := 4
		reader2, err := p.GetReaderAt(context.Background(), fn, int64(offset))
		assert.NoError(t, err)
		require.NotNil(t, reader2)
		assert.Equal(t, int64(4), reader2.GetPosition())
		t.Cleanup(func() {
			assert.NoError(t, reader2.Close())
		})
		t.Run("GetReaderAtOffset", func(t *testing.T) {
			assert.Equal(t, int64(4), reader2.GetPosition(), "Position should be equal to offset")
			readData := make([]byte, 1024)
			n, err := reader2.Read(t.Context(), readData)
			assert.EqualError(t, err, io.EOF.Error())
			assert.Equal(t, len(data)-offset, n)
			assert.Equal(t, data[offset:], readData[:n])
			assert.Equal(t, int64(len(data)), reader2.GetPosition())
		})

		t.Run("GetReaderAtOffsetEOF", func(t *testing.T) {
			readData := make([]byte, 1024)
			n, err := reader2.Read(t.Context(), readData)
			assert.EqualError(t, err, io.EOF.Error())
			assert.Equal(t, 0, n)
			assert.Equal(t, []byte{}, readData[:n])
			assert.Equal(t, int64(len(data)), reader2.GetPosition())
		})
	})

	t.Run("GetReaderAtPositionSuccesfullyReadsDataSequence", func(t *testing.T) {
		t.Parallel()
		p, err := opts.NewIOProvider(context.Background(), t)
		require.NoError(t, err)
		require.NotNil(t, p)
		fn := generateFileName()
		writer, err := p.GetWriterAt(context.Background(), fn, 0)
		require.NoError(t, err)
		require.NotNil(t, writer)
		t.Cleanup(func() {
			assert.NoError(t, writer.Close())
		})

		data := []byte("test data for writing to file")

		t.Run("WriteData", func(t *testing.T) {
			assert.Equal(t, int64(0), writer.GetPosition(), "Initial position should be 0")
			n, err := writer.Write(t.Context(), data)
			require.NoError(t, err)
			require.Equal(t, len(data), n)
			require.Equal(t, int64(len(data)), writer.GetPosition())
		})
		reader, err := p.GetReaderAt(context.Background(), fn, 0)
		assert.NoError(t, err)
		require.NotNil(t, reader)
		t.Cleanup(func() {
			assert.NoError(t, reader.Close())
		})

		t.Run("ReadData", func(t *testing.T) {
			for writer.GetPosition() < int64(len(data)) {
				readData := make([]byte, 8)
				n, err := reader.Read(t.Context(), readData)
				if reader.GetPosition() >= int64(len(data)) {
					assert.EqualError(t, err, io.EOF.Error(), "Expected EOF when reading beyond data length")
				} else {
					assert.NoError(t, err, "Expected no error when reading data")
				}
				assert.Greater(t, n, 0, "Expected to read some data")
				assert.Equal(t, data[reader.GetPosition()-int64(n):reader.GetPosition()], readData[:n])
			}
		})
	})

	t.Run("GetWriterAtPositionSuccesfullyWritesDataSequence", func(t *testing.T) {
		t.Parallel()
		p, err := opts.NewIOProvider(context.Background(), t)
		require.NoError(t, err)
		require.NotNil(t, p)
		fn := generateFileName()

		data := []byte("test data for writing to file")

		for idx := 0; idx < 10; idx++ {
			t.Run(fmt.Sprintf("WriteData %d", idx), func(t *testing.T) {
				writer, err := p.GetWriterAt(context.Background(), fn, 0)
				require.NoError(t, err)
				require.NotNil(t, writer)
				t.Cleanup(func() {
					assert.NoError(t, writer.Close())
				})

				t.Run("WriteData", func(t *testing.T) {
					for i := 0; i < len(data); i += 8 {
						require.Equal(t, int64(i), writer.GetPosition())
						end := i + 8
						if end > len(data) {
							end = len(data)
						}
						n, err := writer.Write(t.Context(), data[i:end])
						require.NoError(t, err)
						require.Equal(t, end-i, n)
						require.Equal(t, int64(end), writer.GetPosition())
					}
				})

				require.Equal(t, int64(len(data)), writer.GetPosition())
				reader, err := p.GetReaderAt(context.Background(), fn, 0)
				assert.NoError(t, err)
				require.NotNil(t, reader)
				t.Cleanup(func() {
					assert.NoError(t, reader.Close())
				})

				t.Run("ReadData", func(t *testing.T) {
					assert.Equal(t, int64(0), reader.GetPosition(), "Initial position should be 0")
					readData := make([]byte, 1024)
					n, err := reader.Read(t.Context(), readData)
					assert.EqualError(t, err, io.EOF.Error())
					assert.Equal(t, len(data), n)
					assert.Equal(t, data, readData[:n])
					assert.Equal(t, int64(len(data)), reader.GetPosition(), "Position should be equal to data length")
				})
			})
		}
	})
}

func generateFileName() string {
	// This function generates a random file name with 20 random chars for testing purposes.
	return "testfile_" + randomString(20) + ".bin"
}

func randomString(length int) string {
	// This function generates a random string of the specified length.
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
