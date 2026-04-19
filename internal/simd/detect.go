package simd

import (
	"runtime"
	"sync"
)

// CPUFeatures holds detected CPU capabilities
type CPUFeatures struct {
	HasAVX2   bool
	HasAVX512 bool
	HasNEON   bool
}

var (
	features     CPUFeatures
	featuresOnce sync.Once
)

// Detect returns CPU SIMD capabilities
func Detect() CPUFeatures {
	featuresOnce.Do(func() {
		features = detectCPUFeatures()
	})
	return features
}

func detectCPUFeatures() CPUFeatures {
	switch runtime.GOARCH {
	case "amd64":
		return CPUFeatures{
			HasAVX2:   detectAVX2(),
			HasAVX512: detectAVX512(),
		}
	case "arm64":
		return CPUFeatures{
			HasNEON: true, // ARM64 always has NEON
		}
	}
	return CPUFeatures{}
}

func detectAVX2() bool {
	// SIMD detection requires CGO and platform-specific assembly.
	// Disabled by default for portability; falls through to scalar.
	// Set CGO_ENABLED=1 and provide platform-specific cpuid bindings to enable.
	return false
}

func detectAVX512() bool {
	// SIMD detection requires CGO and platform-specific assembly.
	// Disabled by default for portability; falls through to scalar.
	// Set CGO_ENABLED=1 and provide platform-specific cpuid bindings to enable.
	return false
}

// Dispatcher selects optimal SIMD implementation
type Dispatcher struct {
	features CPUFeatures
}

// NewDispatcher creates SIMD dispatcher
func NewDispatcher() *Dispatcher {
	return &Dispatcher{features: Detect()}
}

// EntropyFilter dispatches to SIMD or scalar
func (d *Dispatcher) EntropyFilter(data []float64) float64 {
	if d.features.HasAVX2 {
		return entropyAVX2(data)
	}
	if d.features.HasNEON {
		return entropyNEON(data)
	}
	return entropyScalar(data)
}

func entropyAVX2(data []float64) float64 {
	// AVX2 implementation requires CGO and platform-specific assembly.
	// Disabled by default for portability; falls through to scalar.
	return entropyScalar(data)
}

func entropyNEON(data []float64) float64 {
	// NEON implementation requires CGO and platform-specific assembly.
	// Disabled by default for portability; falls through to scalar.
	return entropyScalar(data)
}

func entropyScalar(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		if v > 0 {
			sum -= v * fastLog2(v)
		}
	}
	return sum
}

func fastLog2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return float64(63 - clz(uint64(x)))
}

func clz(x uint64) int {
	if x == 0 {
		return 64
	}
	n := 0
	if x <= 0x00000000FFFFFFFF {
		n += 32
		x <<= 32
	}
	if x <= 0x0000FFFFFFFFFFFF {
		n += 16
		x <<= 16
	}
	if x <= 0x00FFFFFFFFFFFFFF {
		n += 8
		x <<= 8
	}
	if x <= 0x0FFFFFFFFFFFFFFF {
		n += 4
		x <<= 4
	}
	if x <= 0x3FFFFFFFFFFFFFFF {
		n += 2
		x <<= 2
	}
	if x <= 0x7FFFFFFFFFFFFFFF {
		n += 1
	}
	return n
}
