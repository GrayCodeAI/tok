package filter

import "sync"

// BufferPool manages reusable buffers
type BufferPool struct {
	pool *sync.Pool
}

func NewBufferPool(size int) *BufferPool {
	return &BufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return NewZeroCopyBuffer(size)
			},
		},
	}
}

func (bp *BufferPool) Get() *ZeroCopyBuffer {
	buf := bp.pool.Get().(*ZeroCopyBuffer)
	buf.Reset()
	return buf
}

func (bp *BufferPool) Put(buf *ZeroCopyBuffer) {
	bp.pool.Put(buf)
}

var globalBufferPool = NewBufferPool(4096)

func GetBuffer() *ZeroCopyBuffer {
	return globalBufferPool.Get()
}

func PutBuffer(buf *ZeroCopyBuffer) {
	globalBufferPool.Put(buf)
}
