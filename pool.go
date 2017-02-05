package logxi

import (
	"bytes"
	"sync"
)

// BufferPool is a synced pool of byte buffers for building strings.
type BufferPool struct {
	sync.Pool
}

// NewBufferPool creates a BufferPool.
func NewBufferPool() *BufferPool {
	return &BufferPool{
		Pool: sync.Pool{New: func() interface{} {
			b := bytes.NewBuffer(make([]byte, 256))
			b.Reset()
			return b
		}},
	}
}

// Get acquires a buffer from pool.
func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.Pool.Get().(*bytes.Buffer)
}

// Put returns buffer back.
func (bp *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	bp.Pool.Put(b)
}
