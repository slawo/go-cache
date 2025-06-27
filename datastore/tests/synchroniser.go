package tests

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/slawo/go-cache/datastore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ParallelLockTestsOpts struct {
	MaxSyncs            int
	MaxTries            int
	MaxLocks            int
	NewDataSynchroniser func(ctx context.Context, t *testing.T) (datastore.DataSynchroniser, error)
}

func RunParallelLockTests(t *testing.T, opts ParallelLockTestsOpts) {
	require.NotNil(t, opts.NewDataSynchroniser, "NewDataSynchroniser function must be provided")
	if opts.MaxSyncs <= 0 {
		opts.MaxSyncs = 10
	}
	if opts.MaxLocks <= 0 {
		opts.MaxLocks = 50
	}
	if opts.MaxTries <= 0 {
		opts.MaxTries = 3
	}
	syncs := make([]datastore.DataSynchroniser, 0, opts.MaxSyncs)
	for i := 0; i < opts.MaxSyncs; i++ {
		s, err := opts.NewDataSynchroniser(context.Background(), t)
		require.NoError(t, err)
		syncs = append(syncs, s)
	}

	errs := make(chan error)
	rets := make(chan struct {
		syncIdx int
		idx     int
		lock    datastore.DataWriteLock
	})

	errsCt := make(map[int]int, opts.MaxLocks)
	validCt := make(map[int]int, opts.MaxLocks)
	locks := make([]datastore.DataWriteLock, 0, opts.MaxLocks)

	wgr := sync.WaitGroup{}
	wgr.Add(1)
	go func() {
		defer wgr.Done()
		for err := range errs {
			if err != nil {
				msg := err.Error()
				if errors.Is(err, datastore.ErrLockAlreadyHeld) || strings.Contains(err.Error(), datastore.LockAlreadyHeldError) {
					stid := err.Error()[24:29]
					idx, err := strconv.Atoi(stid)
					require.NoError(t, err)
					errsCt[idx]++
					fmt.Printf("Error on %s: %s, ct: %d\n", stid, msg, errsCt[idx])
				}
			}
		}
	}()

	wgr.Add(1)
	go func() {
		defer wgr.Done()
		for ret := range rets {
			if ret.lock != nil {
				fmt.Printf("Success on synchroniser: %02d, lock: %06d\n", ret.syncIdx, ret.idx)
				validCt[ret.idx]++
				locks = append(locks, ret.lock)
			}
		}
	}()

	wg := sync.WaitGroup{}
	st := sync.RWMutex{}
	st.Lock()
	for i, s := range syncs {
		wg.Add(1)
		go func(s datastore.DataSynchroniser, index int) {
			defer wg.Done()
			st.RLock()
			defer st.RUnlock()
			for j := 0; j < opts.MaxLocks; j++ {
				for i := 0; i < opts.MaxTries; i++ {
					lock, err := s.GetWriteLock(context.Background(), fmt.Sprintf("testKey%d", j))
					if err != nil {
						errs <- fmt.Errorf("synchroniser: %02d, lock: %05d: %w", index, j, err)
					} else {
						rets <- struct {
							syncIdx int
							idx     int
							lock    datastore.DataWriteLock
						}{
							syncIdx: index,
							idx:     j,
							lock:    lock,
						}
						time.Sleep(1 * time.Millisecond)
					}
				}
			}
		}(s, i)
	}
	st.Unlock()
	wg.Wait()
	close(rets)
	close(errs)
	wgr.Wait()

	for i := 0; i < opts.MaxLocks; i++ {
		assert.Equal(t, 1, validCt[i], "Lock %d was not acquired", i)
	}
	mt := (opts.MaxTries * len(syncs)) - 1
	for i := 0; i < opts.MaxLocks; i++ {
		assert.Equal(t, mt, errsCt[i], "Lock %d has failed %d times should have failed %d times", i, errsCt[i], mt)
	}

	st.Lock()
	wg = sync.WaitGroup{}
	for _, l := range locks {
		wg.Add(1)
		go func(l datastore.DataWriteLock) {
			defer wg.Done()
			st.RLock()
			defer st.RUnlock()
			assert.False(t, l.Unlocked())
			l.Unlock()
			assert.True(t, l.Unlocked())
		}(l)

		wg.Add(1)
		go func(l datastore.DataWriteLock) {
			defer wg.Done()
			<-l.WaitUnlocked()
		}(l)
	}
	st.Unlock()
	wg.Wait()
}
