package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type IOWriter struct {
	mu sync.RWMutex
	p  int64
	d  *IOData
}

func (w *IOWriter) GetPosition(ctx context.Context) int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.p
}

func (w *IOWriter) Write(ctx context.Context, p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.d == nil {
		return 0, errors.New("file writer: file is not open")
	}
	n, err = w.d.WriteAt(p, w.p)
	w.p += int64(n) // Update the position after reading
	return
}

func (w *IOWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.d == nil {
		return fmt.Errorf("file store: IOReader is not initialized")
	}
	w.d = nil
	w.p = 0
	return nil
}
