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
	bp := &BufferPool{}
	bp.Pool = sync.Pool{New: func() interface{} {
		b := bytes.NewBuffer(make([]byte, 256))
		b.Reset()
		return &PooledBuffer{b, bp}
	}}
	return bp
}

// Get acquires a buffer from pool.
func (bp *BufferPool) Get() *PooledBuffer {
	return bp.Pool.Get().(*PooledBuffer)
}

// Put returns buffer back.
func (bp *BufferPool) Put(pb *PooledBuffer) {
	pb.Reset()
	bp.Pool.Put(pb)
}

// PooledBuffer is a *bytes.Buffer which belongs to a pool, it must Released
// after use.
type PooledBuffer struct {
	*bytes.Buffer
	pool *BufferPool
}

// Release puts pooled buffer back in the pool
func (pb *PooledBuffer) Release() {
	pb.Reset()
	pb.pool.Put(pb)
}
