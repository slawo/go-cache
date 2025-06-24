package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/slawo/go-cache/datastore"
)

// NewSynchroniser creates a new Synchroniser instance.
func NewSynchroniser() (*Synchroniser, error) {
	return &Synchroniser{
		locks: make(map[string]*MutexWriteLock),
	}, nil
}

type Synchroniser struct {
	mu    sync.Mutex
	locks map[string]*MutexWriteLock
}

func (r *Synchroniser) GetWriteLock(ctx context.Context, lockID string) (datastore.DataWriteLock, error) {
	if strings.TrimSpace(lockID) == "" {
		return nil, datastore.ErrInvalidLockID
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.locks[lockID]; exists {
		return nil, fmt.Errorf("%w: %s", datastore.ErrLockAlreadyHeld, lockID)
	}
	lock := &MutexWriteLock{
		s:        r,
		lockID:   lockID,
		unlocked: make(chan struct{}),
	}
	r.locks[lockID] = lock
	return lock, nil
}

func (r *Synchroniser) removeWriteLock(l *MutexWriteLock) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	s        *Synchroniser
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
