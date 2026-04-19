package compressor

import (
	"testing"
)

func BenchmarkCompressLite(b *testing.B) {
	text := "Please really just utilize the configuration file and actually just do it simply"
	for i := 0; i < b.N; i++ {
		compressLite(text)
	}
}

func BenchmarkCompressFull(b *testing.B) {
	text := "The database connection is actually just working fine and really doing well"
	for i := 0; i < b.N; i++ {
		compressFull(text)
	}
}

func BenchmarkCompressUltra(b *testing.B) {
	text := "The database authentication configuration needs to be updated and fixed properly"
	for i := 0; i < b.N; i++ {
		compressUltra(text)
	}
}

func BenchmarkCompressParallel(b *testing.B) {
	texts := []string{
		"Please really just utilize the configuration",
		"The database connection is actually just working fine",
		"The database authentication configuration needs to be updated",
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			Compress(texts[i%len(texts)], "ultra")
			i++
		}
	})
}
