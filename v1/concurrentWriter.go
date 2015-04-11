package log

import (
	"io"
	"sync"
)

// ConcurrentWriter is a concurrent safe wrapper around io.Writer
type ConcurrentWriter struct {
	writer io.Writer
	sync.Mutex
}

// NewConcurrentWriter crates a new concurrent writer wrapper around existing writer.
func NewConcurrentWriter(writer io.Writer) io.Writer {
	return &ConcurrentWriter{writer: writer}
}

func (cw *ConcurrentWriter) Write(p []byte) (n int, err error) {
	cw.Lock()
	defer cw.Unlock()
	// this is basically the same logic as in go's log.Output()
	return cw.writer.Write(p)
}
