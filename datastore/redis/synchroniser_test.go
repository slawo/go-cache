package redis_test

import (
	"context"
	"testing"

	"github.com/slawo/go-cache/datastore"
	"github.com/slawo/go-cache/datastore/redis"
	"github.com/slawo/go-cache/datastore/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewServer(t *testing.T) string {
	ctx := t.Context()
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)

	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, redisC)
		require.NoError(t, err)
	})

	host, err := redisC.Host(ctx)
	require.NoError(t, err)
	port, err := redisC.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)
	dsn := host + ":" + port.Port()

	return dsn
}

func NewSynchroniser(t *testing.T) *redis.RedisSynchroniser {
	dsn := NewServer(t)
	ctx := t.Context()
	s, err := redis.NewSynchroniser(ctx, dsn, "", 0)
	assert.NoError(t, err)
	assert.IsType(t, &redis.RedisSynchroniser{}, s)
	return s
}

// func TestNewMutexSynchroniser(t *testing.T) {
// 	s, err := datastore.NewMutexSynchroniser()
// 	assert.NoError(t, err)
// 	assert.IsType(t, &datastore.MutexSynchroniser{}, s)
// }

func TestSynchroniserGetLockWithEmptyKey(t *testing.T) {
	s := NewSynchroniser(t)

	lock, err := s.GetWriteLock(context.Background(), "")
	assert.EqualError(t, err, "invalid lock ID")
	assert.Nil(t, lock)
}

func TestMutexSynchroniserGetLockWithSpacesKey(t *testing.T) {
	s := NewSynchroniser(t)

	lock, err := s.GetWriteLock(context.Background(), "  ")
	assert.EqualError(t, err, "invalid lock ID")
	assert.Nil(t, lock)
}

func TestMutexSynchroniserGetLock(t *testing.T) {
	s := NewSynchroniser(t)

	lock, err := s.GetWriteLock(context.Background(), "testKey")
	assert.NoError(t, err)
	assert.NotNil(t, lock)
	assert.IsType(t, &redis.WriteLock{}, lock)

}

func TestMutexSynchroniserGetLockErrorOnSecondCall(t *testing.T) {
	s := NewSynchroniser(t)

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

	err = lock1.Unlock()
	assert.NoError(t, err)

	lock4, err := s.GetWriteLock(context.Background(), "testKey2")
	assert.EqualError(t, err, "lock is already held: testKey2")
	assert.Nil(t, lock4)
}

func TestMutexSynchroniserGetLockMultiTest(t *testing.T) {
	dsn := NewServer(t)
	create := func(ctx context.Context, t *testing.T) (datastore.DataSynchroniser, error) {
		s, err := redis.NewSynchroniser(context.Background(), dsn, "", 0)
		require.NoError(t, err)
		require.IsType(t, &redis.RedisSynchroniser{}, s)
		return s, err
	}
	tests.RunParallelLockTests(t, tests.ParallelLockTestsOpts{
		NewDataSynchroniser: create,
		MaxSyncs:            10,
		MaxTries:            3,
		MaxLocks:            50,
	})
}
