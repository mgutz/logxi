package log

import (
	"bytes"
	"sync"
)

type bufferPool struct {
	sync.Pool
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		Pool: sync.Pool{New: func() interface{} {
			b := bytes.NewBuffer(make([]byte, 128))
			b.Reset()
			return b
		}},
	}
}

func (bp *bufferPool) get() *bytes.Buffer {
	return bp.Pool.Get().(*bytes.Buffer)
}

func (bp *bufferPool) put(b *bytes.Buffer) {
	b.Reset()
	bp.Pool.Put(b)
}
