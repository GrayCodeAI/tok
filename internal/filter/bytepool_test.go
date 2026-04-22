package filter

import (
	"testing"
)

func TestNewBytePool(t *testing.T) {
	bp := NewBytePool()
	if bp == nil {
		t.Fatal("expected non-nil BytePool")
	}
}

func TestBytePool_GetPut(t *testing.T) {
	bp := NewBytePool()

	sizes := []int{100, 5000, 50000, 500000}
	for _, size := range sizes {
		b := bp.Get(size)
		if b == nil {
			t.Fatalf("expected non-nil buffer for capacity %d", size)
		}
		if cap(*b) < size {
			t.Errorf("capacity %d < requested %d", cap(*b), size)
		}

		// Write some data
		*b = append(*b, []byte("test data")...)
		bp.Put(b)

		// Get again — should be reset
		b2 := bp.Get(size)
		if len(*b2) != 0 {
			t.Errorf("expected zero-length buffer after Put, got %d", len(*b2))
		}
	}
}

func TestBytePool_PutNil(t *testing.T) {
	bp := NewBytePool()
	// Should not panic
	bp.Put(nil)
}

func TestGetBytesPutBytes(t *testing.T) {
	b := GetBytes(100)
	if b == nil {
		t.Fatal("expected non-nil bytes")
	}
	*b = append(*b, []byte("hello")...)
	PutBytes(b)
}

func TestFastStringBuilder(t *testing.T) {
	fb := NewFastStringBuilder(64)
	if fb == nil {
		t.Fatal("expected non-nil FastStringBuilder")
	}

	fb.WriteString("hello ")
	fb.WriteString("world")
	fb.WriteByte('!')
	fb.Write([]byte(" test"))

	if fb.Len() != 17 {
		t.Errorf("expected len 17, got %d", fb.Len())
	}
	if fb.String() != "hello world! test" {
		t.Errorf("unexpected string: %q", fb.String())
	}

	fb.Reset()
	if fb.Len() != 0 {
		t.Errorf("expected len 0 after reset, got %d", fb.Len())
	}
	if fb.Cap() == 0 {
		t.Error("expected non-zero capacity after reset")
	}
}

func TestFastStringBuilder_Grow(t *testing.T) {
	fb := NewFastStringBuilder(10)
	fb.WriteString("hello")
	fb.Grow(100)
	if fb.Cap() < 105 {
		t.Errorf("expected capacity >= 105, got %d", fb.Cap())
	}
}

func TestBytePool_AcquireReleaseStringBuilder(t *testing.T) {
	bp := NewBytePool()

	// Small
	buf := bp.AcquireStringBuilder(100)
	buf.WriteString("hello")
	bp.ReleaseStringBuilder(buf)

	// Medium
	buf2 := bp.AcquireStringBuilder(5000)
	buf2.WriteString("world")
	bp.ReleaseStringBuilder(buf2)

	// Large
	buf3 := bp.AcquireStringBuilder(50000)
	buf3.WriteString("big data")
	bp.ReleaseStringBuilder(buf3)

	// Nil
	bp.ReleaseStringBuilder(nil)
}
