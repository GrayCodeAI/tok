package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"time"
)

// CompressionComparison holds comparison results between different algorithms
type CompressionComparison struct {
	Algorithm        string        `json:"algorithm"`
	OriginalSize     int           `json:"original_size"`
	CompressedSize   int           `json:"compressed_size"`
	CompressionRatio float64       `json:"compression_ratio"`
	SpaceSaved       int           `json:"space_saved"`
	Percentage       float64       `json:"percentage"`
	Duration         time.Duration `json:"duration"`
	Speed            float64       `json:"speed_mbps"`
}

// CompareResult holds the full comparison
type CompareResult struct {
	OriginalData []byte                  `json:"-"`
	Algorithms   []CompressionComparison `json:"algorithms"`
	Winner       string                  `json:"winner"`
	BestRatio    float64                 `json:"best_ratio"`
}

// CompareAlgorithms compares multiple compression algorithms
func CompareAlgorithms(data []byte) (*CompareResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to compress")
	}

	result := &CompareResult{
		OriginalData: data,
		Algorithms:   []CompressionComparison{},
	}

	// Test Brotli at different levels
	for _, level := range []int{1, 4, 8, 11} {
		comp := NewBrotliCompressorWithConfig(BrotliConfig{Quality: level, LGWin: 22})

		start := time.Now()
		compressed, err := comp.Compress(data)
		duration := time.Since(start)

		if err != nil {
			continue
		}

		speed := float64(len(data)) / duration.Seconds() / 1024 / 1024 // MB/s

		comparison := CompressionComparison{
			Algorithm:        fmt.Sprintf("brotli-%d", level),
			OriginalSize:     len(data),
			CompressedSize:   len(compressed),
			CompressionRatio: float64(len(compressed)) / float64(len(data)),
			SpaceSaved:       len(data) - len(compressed),
			Percentage:       (1.0 - float64(len(compressed))/float64(len(data))) * 100,
			Duration:         duration,
			Speed:            speed,
		}

		result.Algorithms = append(result.Algorithms, comparison)
	}

	// Test Gzip at different levels
	for _, level := range []int{1, 6, 9} {
		start := time.Now()
		var buf bytes.Buffer
		w, _ := gzip.NewWriterLevel(&buf, level)
		w.Write(data)
		w.Close()
		duration := time.Since(start)

		compressed := buf.Bytes()
		speed := float64(len(data)) / duration.Seconds() / 1024 / 1024

		comparison := CompressionComparison{
			Algorithm:        fmt.Sprintf("gzip-%d", level),
			OriginalSize:     len(data),
			CompressedSize:   len(compressed),
			CompressionRatio: float64(len(compressed)) / float64(len(data)),
			SpaceSaved:       len(data) - len(compressed),
			Percentage:       (1.0 - float64(len(compressed))/float64(len(data))) * 100,
			Duration:         duration,
			Speed:            speed,
		}

		result.Algorithms = append(result.Algorithms, comparison)
	}

	// Find winner
	bestRatio := 1.0
	for _, comp := range result.Algorithms {
		if comp.CompressionRatio < bestRatio {
			bestRatio = comp.CompressionRatio
			result.Winner = comp.Algorithm
		}
	}
	result.BestRatio = bestRatio

	return result, nil
}

// PrintComparison prints a formatted comparison table
func (cr *CompareResult) PrintComparison() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("\nCompression Comparison (%d bytes original)\n", len(cr.OriginalData)))
	buf.WriteString(string(bytes.Repeat([]byte("="), 70)) + "\n")
	buf.WriteString(fmt.Sprintf("%-12s %12s %12s %10s %10s %12s\n",
		"Algorithm", "Original", "Compressed", "Ratio", "Saved %", "Speed (MB/s)"))
	buf.WriteString(string(bytes.Repeat([]byte("-"), 70)) + "\n")

	for _, comp := range cr.Algorithms {
		winner := ""
		if comp.Algorithm == cr.Winner {
			winner = "*"
		}
		buf.WriteString(fmt.Sprintf("%-12s %12s %12s %9.2f%% %9.1f%% %12.2f%s\n",
			comp.Algorithm,
			formatBytes(comp.OriginalSize),
			formatBytes(comp.CompressedSize),
			comp.CompressionRatio*100,
			comp.Percentage,
			comp.Speed,
			winner))
	}

	buf.WriteString(string(bytes.Repeat([]byte("-"), 70)) + "\n")
	buf.WriteString(fmt.Sprintf("Winner: %s (%.2f%% of original size)\n", cr.Winner, cr.BestRatio*100))

	return buf.String()
}

func formatBytes(b int) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
