package filter

import (
	"bytes"
	"strings"
	"sync"
)

// BytePool provides a pool of reusable byte slices to reduce allocations.
type BytePool struct {
	pools []*sync.Pool // One pool per size tier
}

var (
	// Global byte pool instance
	globalBytePool *BytePool
	initBytePool   sync.Once
)

// GetBytePool returns the global byte pool instance.
func GetBytePool() *BytePool {
	initBytePool.Do(func() {
		globalBytePool = NewBytePool()
	})
	return globalBytePool
}

// NewBytePool creates a new byte pool with multiple size tiers.
func NewBytePool() *BytePool {
	// Size tiers: 1KB, 4KB, 16KB, 64KB, 256KB, 1MB
	sizes := []int{1024, 4096, 16384, 65536, 262144, 1048576}
	pools := make([]*sync.Pool, len(sizes))

	for i, size := range sizes {
		s := size // Capture for closure
		pools[i] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, s)
			},
		}
	}

	return &BytePool{pools: pools}
}

// Get retrieves a byte slice from the pool.
// The slice will have at least the requested capacity.
func (bp *BytePool) Get(minCapacity int) []byte {
	poolIdx := bp.selectPool(minCapacity)
	if poolIdx < 0 {
		// Too large for pool, allocate directly
		return make([]byte, 0, minCapacity)
	}

	buf := bp.pools[poolIdx].Get().([]byte)
	return buf[:0] // Reset length but keep capacity
}

// Put returns a byte slice to the pool.
func (bp *BytePool) Put(buf []byte) {
	if cap(buf) == 0 {
		return
	}

	poolIdx := bp.selectPool(cap(buf))
	if poolIdx < 0 {
		// Don't return oversized buffers to pool
		return
	}

	// Clear the slice to avoid retaining references
	for i := range buf[:cap(buf)] {
		buf[i] = 0
	}

	bp.pools[poolIdx].Put(buf[:0])
}

// selectPool chooses the appropriate pool based on size.
func (bp *BytePool) selectPool(size int) int {
	switch {
	case size <= 1024:
		return 0
	case size <= 4096:
		return 1
	case size <= 16384:
		return 2
	case size <= 65536:
		return 3
	case size <= 262144:
		return 4
	case size <= 1048576:
		return 5
	default:
		return -1 // Too large
	}
}

// StringBuilderPool provides a pool of reusable strings.Builder.
type StringBuilderPool struct {
	pool sync.Pool
}

var (
	globalStringBuilderPool *StringBuilderPool
	initStringBuilderPool   sync.Once
)

// GetStringBuilderPool returns the global string builder pool.
func GetStringBuilderPool() *StringBuilderPool {
	initStringBuilderPool.Do(func() {
		globalStringBuilderPool = NewStringBuilderPool()
	})
	return globalStringBuilderPool
}

// NewStringBuilderPool creates a new string builder pool.
func NewStringBuilderPool() *StringBuilderPool {
	return &StringBuilderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
	}
}

// Get retrieves a strings.Builder from the pool.
func (sbp *StringBuilderPool) Get() *strings.Builder {
	return sbp.pool.Get().(*strings.Builder)
}

// Put returns a strings.Builder to the pool.
func (sbp *StringBuilderPool) Put(sb *strings.Builder) {
	sb.Reset()
	sbp.pool.Put(sb)
}

// PooledBuffer represents a buffer that can be returned to the pool.
type PooledBuffer struct {
	*bytes.Buffer
	pool *BufferPool
}

// Release returns the buffer to the pool.
func (pb *PooledBuffer) Release() {
	if pb.pool != nil {
		pb.pool.Put(pb.Buffer)
	}
}

// BufferPool provides a pool of reusable buffers with automatic sizing.
type BufferPool struct {
	pool sync.Pool
}

var (
	globalBufferPool *BufferPool
	initBufferPool   sync.Once
)

// GetBufferPool returns the global buffer pool.
func GetBufferPool() *BufferPool {
	initBufferPool.Do(func() {
		globalBufferPool = &BufferPool{
			pool: sync.Pool{
				New: func() interface{} {
					return new(bytes.Buffer)
				},
			},
		}
	})
	return globalBufferPool
}

// Get retrieves a buffer from the pool.
func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

// Put returns a buffer to the pool.
func (bp *BufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	bp.pool.Put(buf)
}
