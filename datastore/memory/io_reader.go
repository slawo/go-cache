package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type IOReader struct {
	mu sync.RWMutex
	p  int64
	d  *IOData
}

func (r *IOReader) GetPosition() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.p
}

func (r *IOReader) Read(ctx context.Context, p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.d == nil {
		return 0, errors.New("file writer: file is not open")
	}
	n, err = r.d.ReadAt(p, r.p)
	r.p += int64(n)
	return
}

func (r *IOReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.d == nil {
		return fmt.Errorf("file store: IOReader is not initialized")
	}
	r.d = nil
	r.p = 0
	return nil
}
