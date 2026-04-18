package filter

import "unsafe"

// ZeroCopyBuffer provides zero-copy string operations
type ZeroCopyBuffer struct {
	data []byte
}

func NewZeroCopyBuffer(capacity int) *ZeroCopyBuffer {
	return &ZeroCopyBuffer{data: make([]byte, 0, capacity)}
}

func (z *ZeroCopyBuffer) Append(s string) {
	z.data = append(z.data, s...)
}

func (z *ZeroCopyBuffer) AppendByte(b byte) {
	z.data = append(z.data, b)
}

func (z *ZeroCopyBuffer) String() string {
	return *(*string)(unsafe.Pointer(&z.data))
}

func (z *ZeroCopyBuffer) Reset() {
	z.data = z.data[:0]
}

func (z *ZeroCopyBuffer) Len() int {
	return len(z.data)
}

// StringToBytes converts string to []byte without allocation
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

// BytesToString converts []byte to string without allocation
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
