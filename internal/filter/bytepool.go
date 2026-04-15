package filter

import (
	"bytes"
	"sync"
)

// BytePool provides a pool of reusable byte buffers to reduce GC pressure
// Used for high-frequency string operations in the compression pipeline

type BytePool struct {
	smallPool  sync.Pool // < 1KB
	mediumPool sync.Pool // 1KB - 10KB
	largePool  sync.Pool // 10KB - 100KB
	hugePool   sync.Pool // > 100KB
}

// bufferPools holds reusable bytes.Buffer pools by size category
type bufferPools struct {
	small  sync.Pool // < 1KB
	medium sync.Pool // 1KB - 10KB
	large  sync.Pool // 10KB - 100KB
}

var globalBufferPools = &bufferPools{
	small: sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 1024))
		},
	},
	medium: sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 10240))
		},
	},
	large: sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 102400))
		},
	},
}

// Global byte pool instance
var globalBytePool = NewBytePool()

// NewBytePool creates a new byte buffer pool
func NewBytePool() *BytePool {
	return &BytePool{
		smallPool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 1024)
				return &b
			},
		},
		mediumPool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 10240)
				return &b
			},
		},
		largePool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 102400)
				return &b
			},
		},
		hugePool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 0, 1048576)
				return &b
			},
		},
	}
}

// Get retrieves a buffer from the pool based on desired capacity
func (p *BytePool) Get(capacity int) *[]byte {
	switch {
	case capacity <= 1024:
		return p.smallPool.Get().(*[]byte)
	case capacity <= 10240:
		return p.mediumPool.Get().(*[]byte)
	case capacity <= 102400:
		return p.largePool.Get().(*[]byte)
	default:
		return p.hugePool.Get().(*[]byte)
	}
}

// Put returns a buffer to the pool
func (p *BytePool) Put(b *[]byte) {
	if b == nil {
		return
	}

	// Reset buffer to zero length (keep capacity)
	*b = (*b)[:0]

	capLen := cap(*b)
	switch {
	case capLen <= 1024:
		p.smallPool.Put(b)
	case capLen <= 10240:
		p.mediumPool.Put(b)
	case capLen <= 102400:
		p.largePool.Put(b)
	default:
		p.hugePool.Put(b)
	}
}

// AcquireStringBuilder gets a bytes.Buffer from the pool based on capacity needs.
// The returned buffer is reset and ready for use.
func (p *BytePool) AcquireStringBuilder(capacity int) *bytes.Buffer {
	var buf *bytes.Buffer

	switch {
	case capacity <= 1024:
		buf = globalBufferPools.small.Get().(*bytes.Buffer)
	case capacity <= 10240:
		buf = globalBufferPools.medium.Get().(*bytes.Buffer)
	default:
		buf = globalBufferPools.large.Get().(*bytes.Buffer)
	}

	buf.Reset()
	return buf
}

// ReleaseStringBuilder returns a buffer to the appropriate pool for reuse.
// Buffers are reset before pooling to prevent memory leaks.
func (p *BytePool) ReleaseStringBuilder(buf *bytes.Buffer) {
	if buf == nil {
		return
	}

	// Reset and return to pool based on capacity
	capLen := buf.Cap()
	switch {
	case capLen <= 1024:
		globalBufferPools.small.Put(buf)
	case capLen <= 10240:
		globalBufferPools.medium.Put(buf)
	default:
		globalBufferPools.large.Put(buf)
	}
}

// GetBytes is a convenience function to get bytes from the global pool
func GetBytes(capacity int) *[]byte {
	return globalBytePool.Get(capacity)
}

// PutBytes is a convenience function to return bytes to the global pool
func PutBytes(b *[]byte) {
	globalBytePool.Put(b)
}

// FastStringBuilder provides a high-performance string builder with pooling
type FastStringBuilder struct {
	buf []byte
}

// NewFastStringBuilder creates a new fast string builder
func NewFastStringBuilder(capacity int) *FastStringBuilder {
	return &FastStringBuilder{
		buf: make([]byte, 0, capacity),
	}
}

// WriteString appends a string
func (b *FastStringBuilder) WriteString(s string) {
	b.buf = append(b.buf, s...)
}

// WriteByte appends a byte
func (b *FastStringBuilder) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// Write appends bytes
func (b *FastStringBuilder) Write(p []byte) {
	b.buf = append(b.buf, p...)
}

// String returns the built string
func (b *FastStringBuilder) String() string {
	return string(b.buf)
}

// Reset clears the buffer (keeps capacity)
func (b *FastStringBuilder) Reset() {
	b.buf = b.buf[:0]
}

// Len returns the current length
func (b *FastStringBuilder) Len() int {
	return len(b.buf)
}

// Cap returns the capacity
func (b *FastStringBuilder) Cap() int {
	return cap(b.buf)
}

// Grow grows the buffer capacity
func (b *FastStringBuilder) Grow(n int) {
	if n > cap(b.buf)-len(b.buf) {
		newBuf := make([]byte, len(b.buf), cap(b.buf)+n)
		copy(newBuf, b.buf)
		b.buf = newBuf
	}
}
