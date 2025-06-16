package datastore

import (
	"context"
	"errors"
	"strings"
	"sync"
)

const (
	InvalidLockIDError = "invalid lock ID"
)

var (
	// ErrInvalidFileID is returned when an invalid file ID is provided.
	ErrInvalidLockID = errors.New(InvalidLockIDError)
)

func NewMutexSynchroniser() (*MutexSynchroniser, error) {
	return &MutexSynchroniser{
		locks: make(map[string]*MutexWriteLock),
	}, nil
}

type MutexSynchroniser struct {
	mu    sync.Mutex
	locks map[string]*MutexWriteLock
}

func (r *MutexSynchroniser) GetWriteLock(ctx context.Context, lockID string) (DataWriteLock, error) {
	if strings.TrimSpace(lockID) == "" {
		return nil, ErrInvalidLockID
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.locks[lockID]; exists {
		return nil, errors.New("lock is already held")
	}
	lock := &MutexWriteLock{
		s:        r,
		lockID:   lockID,
		unlocked: make(chan struct{}),
	}
	r.locks[lockID] = lock
	return lock, nil
}

func (r *MutexSynchroniser) removeWriteLock(l *MutexWriteLock) (bool, error) {
	if l == nil {
		return false, errors.New("lock is nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if l.Unlocked() {
		return false, nil
	}
	lock, exists := r.locks[l.lockID]
	if !exists {
		return false, errors.New("lock is not held")
	}
	if l != lock {
		return false, errors.New("lock does not match the held lock")
	}
	delete(r.locks, l.lockID)
	close(l.unlocked)
	return true, nil
}

type MutexWriteLock struct {
	s        *MutexSynchroniser
	lockID   string
	unlocked chan struct{}
}

func (l *MutexWriteLock) Unlock() error {
	_, err := l.s.removeWriteLock(l)
	return err
}

func (l *MutexWriteLock) Unlocked() bool {
	select {
	case <-l.unlocked:
		return true
	default:
		return false
	}
}

func (l *MutexWriteLock) WaitUnlocked() <-chan struct{} {
	return l.unlocked
}
