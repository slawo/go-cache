package memory_test

import (
	"context"
	"testing"

	"github.com/slawo/go-cache/datastore"
	"github.com/slawo/go-cache/datastore/memory"
	"github.com/slawo/go-cache/datastore/tests"
	"github.com/stretchr/testify/assert"
)

func TestNewSynchroniser(t *testing.T) {
	s, err := memory.NewSynchroniser()
	assert.NoError(t, err)
	assert.IsType(t, &memory.Synchroniser{}, s)
}

func TestSynchroniserGetLockWithEmptyKey(t *testing.T) {
	s, err := memory.NewSynchroniser()
	assert.NoError(t, err)

	lock, err := s.GetWriteLock(context.Background(), "")
	assert.EqualError(t, err, "invalid lock ID")
	assert.Nil(t, lock)
}

func TestSynchroniserGetLockWithSpacesKey(t *testing.T) {
	s, err := memory.NewSynchroniser()
	assert.NoError(t, err)

	lock, err := s.GetWriteLock(context.Background(), "  ")
	assert.EqualError(t, err, "invalid lock ID")
	assert.Nil(t, lock)
}

func TestSynchroniserGetLock(t *testing.T) {
	s, err := memory.NewSynchroniser()
	assert.NoError(t, err)

	lock, err := s.GetWriteLock(context.Background(), "testKey")
	assert.NoError(t, err)
	assert.NotNil(t, lock)
}

func TestSynchroniserGetLockErrorOnSecondCall(t *testing.T) {
	s, err := memory.NewSynchroniser()
	assert.NoError(t, err)

	lock1, err := s.GetWriteLock(context.Background(), "testKey2")
	assert.NoError(t, err)
	assert.NotNil(t, lock1)

	lock2, err := s.GetWriteLock(context.Background(), "testKey2")
	assert.EqualError(t, err, "lock is already held: testKey2")
	assert.Nil(t, lock2)

	err = lock1.Unlock()
	assert.NoError(t, err)

	lock3, err := s.GetWriteLock(context.Background(), "testKey2")
	assert.NoError(t, err)
	assert.NotNil(t, lock3)
}

func TestSynchroniserLockTests(t *testing.T) {
	create := func(ctx context.Context, t *testing.T) (datastore.DataSynchroniser, error) {
		return memory.NewSynchroniser()
	}
	tests.RunParallelLockTests(t, tests.ParallelLockTestsOpts{
		MaxSyncs:            1,
		MaxTries:            30,
		MaxLocks:            500,
		NewDataSynchroniser: create,
	})
}
