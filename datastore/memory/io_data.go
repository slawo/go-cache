package memory

import (
	"errors"
	"io"
	"sync"
)

func NewIOData() *IOData {
	return &IOData{
		d: []byte{},
	}
}

type IOData struct {
	mu sync.RWMutex
	d  []byte
}

func (d *IOData) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("read at negative offset")
	}
	if len(p) == 0 {
		return 0, errors.New("read to empty array")
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if off >= int64(len(d.d)) {
		return 0, io.EOF
	}
	start := int(off)
	end := start + len(p)
	if end > len(d.d) {
		end = len(d.d)
	}
	if start >= end {
		return 0, io.EOF
	}
	n = copy(p, d.d[start:])
	if n < len(p) {
		err = io.EOF
	}
	return
}

func (d *IOData) WriteAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("write at negative offset")
	}
	if len(p) == 0 {
		return 0, nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for off > int64(len(d.d)) {
		d.d = append(d.d, 0)
	}
	for i := 0; i < len(p); i++ {
		if i+int(off) >= len(d.d) {
			d.d = append(d.d, p[i:]...)
			return len(p), nil
		}
		d.d[int(off)+i] = p[i]
		n++
	}
	return n, nil
}
