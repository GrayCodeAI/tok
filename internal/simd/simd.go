package simd

import "strings"

func HasANSI(data string) bool {
	return strings.Contains(data, "\x1b")
}

func ContainsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func ContainsWord(s, word string) bool {
	return strings.Contains(s, word)
}

func IsWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func SplitWords(s string) []string {
	var words []string
	var current strings.Builder
	for i := 0; i < len(s); i++ {
		if IsWordChar(s[i]) {
			current.WriteByte(s[i])
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

func StripANSI(data string) string {
	var result strings.Builder
	result.Grow(len(data))
	inEscape := false
	for i := 0; i < len(data); i++ {
		if data[i] == 0x1b && i+1 < len(data) && data[i+1] == '[' {
			inEscape = true
			i++
			continue
		}
		if inEscape {
			if (data[i] >= 'A' && data[i] <= 'Z') || (data[i] >= 'a' && data[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteByte(data[i])
	}
	return result.String()
}

func Process(data string) string {
	return StripANSI(data)
}
