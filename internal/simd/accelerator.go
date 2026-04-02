package simd

import "runtime"

type SIMDAccelerator struct {
	avx2   bool
	avx512 bool
	neon   bool
}

func NewSIMDAccelerator() *SIMDAccelerator {
	return &SIMDAccelerator{
		avx2:   runtime.GOARCH == "amd64",
		avx512: runtime.GOARCH == "amd64",
		neon:   runtime.GOARCH == "arm64",
	}
}

func (s *SIMDAccelerator) HasSIMD() bool {
	return s.avx2 || s.avx512 || s.neon
}

func (s *SIMDAccelerator) Features() map[string]bool {
	return map[string]bool{
		"avx2":   s.avx2,
		"avx512": s.avx512,
		"neon":   s.neon,
	}
}

func (s *SIMDAccelerator) CountBytes(data []byte, target byte) int {
	count := 0
	for _, b := range data {
		if b == target {
			count++
		}
	}
	return count
}

func (s *SIMDAccelerator) FindByte(data []byte, target byte) int {
	for i, b := range data {
		if b == target {
			return i
		}
	}
	return -1
}

func (s *SIMDAccelerator) CompareBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
