package datastore

import (
	"context"
	"errors"
)

const (
	InvalidLockIDError   = "invalid lock ID"
	LockAlreadyHeldError = "lock is already held"
)

var (
	// ErrInvalidFileID is returned when an invalid file ID is provided.
	ErrInvalidLockID = errors.New(InvalidLockIDError)
	// ErrLockAlreadyHeld is returned when a lock is already held.
	ErrLockAlreadyHeld = errors.New(LockAlreadyHeldError)
)

//go:generate mockery --name DataSynchroniser --output mocks
type DataSynchroniser interface {
	GetWriteLock(ctx context.Context, lockID string) (DataWriteLock, error)
}

//go:generate mockery --name DataWriteLock --output mocks
type DataWriteLock interface {
	// Unlock releases the lock. A new lock will need to be created.
	Unlock() error
	// Unlocked checks if the lock is currently held.
	Unlocked() bool
	// WaitUnlocked returns a channel that will be closed when the lock is released.
	WaitUnlocked() <-chan struct{}
}
