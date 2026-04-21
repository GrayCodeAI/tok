package filter

import "github.com/GrayCodeAI/tok/internal/simd"

// SIMDEntropyFilter uses SIMD-optimized entropy calculation
type SIMDEntropyFilter struct {
	dispatcher *simd.Dispatcher
	threshold  float64
}

func NewSIMDEntropyFilter(threshold float64) *SIMDEntropyFilter {
	return &SIMDEntropyFilter{
		dispatcher: simd.NewDispatcher(),
		threshold:  threshold,
	}
}

func (f *SIMDEntropyFilter) Apply(input string, mode Mode) (string, int) {
	if len(input) < 50 {
		return input, 0
	}

	lines := splitLinesSimd(input)
	kept := make([]string, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		freq := make([]float64, 256)
		for i := 0; i < len(line); i++ {
			freq[line[i]]++
		}

		total := float64(len(line))
		for i := range freq {
			if freq[i] > 0 {
				freq[i] /= total
			}
		}

		entropy := f.dispatcher.EntropyFilter(freq)
		if entropy >= f.threshold {
			kept = append(kept, line)
		}
	}

	result := joinLinesSimd(kept)
	saved := len(input) - len(result)
	return result, saved / 4
}

func splitLinesSimd(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLinesSimd(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	total := 0
	for _, l := range lines {
		total += len(l) + 1
	}
	buf := make([]byte, 0, total)
	for _, l := range lines {
		buf = append(buf, l...)
		buf = append(buf, '\n')
	}
	return string(buf)
}
