package datastore_test

import (
	"context"
	"testing"

	"github.com/slawo/go-cache/datastore"
	"github.com/stretchr/testify/assert"
)

func TestNewMutexSynchroniser(t *testing.T) {
	s, err := datastore.NewMutexSynchroniser()
	assert.NoError(t, err)
	assert.IsType(t, &datastore.MutexSynchroniser{}, s)
}

func TestMutexSynchroniserGetLockWithEmptyKey(t *testing.T) {
	s, err := datastore.NewMutexSynchroniser()
	assert.NoError(t, err)

	lock, err := s.GetWriteLock(context.Background(), "")
	assert.EqualError(t, err, "invalid lock ID")
	assert.Nil(t, lock)
}

func TestMutexSynchroniserGetLockWithSpacesKey(t *testing.T) {
	s, err := datastore.NewMutexSynchroniser()
	assert.NoError(t, err)

	lock, err := s.GetWriteLock(context.Background(), "  ")
	assert.EqualError(t, err, "invalid lock ID")
	assert.Nil(t, lock)
}

func TestMutexSynchroniserGetLock(t *testing.T) {
	s, err := datastore.NewMutexSynchroniser()
	assert.NoError(t, err)

	lock, err := s.GetWriteLock(context.Background(), "testKey")
	assert.NoError(t, err)
	assert.NotNil(t, lock)
}

func TestMutexSynchroniserGetLockErrorOnSecondCall(t *testing.T) {
	s, err := datastore.NewMutexSynchroniser()
	assert.NoError(t, err)

	lock1, err := s.GetWriteLock(context.Background(), "testKey2")
	assert.NoError(t, err)
	assert.NotNil(t, lock1)

	lock2, err := s.GetWriteLock(context.Background(), "testKey2")
	assert.EqualError(t, err, "lock is already held")
	assert.Nil(t, lock2)

	err = lock1.Unlock()
	assert.NoError(t, err)

	lock3, err := s.GetWriteLock(context.Background(), "testKey2")
	assert.NoError(t, err)
	assert.NotNil(t, lock3)
}
