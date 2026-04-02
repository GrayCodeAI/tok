package teecommand

import (
	"os/exec"
	"strings"
	"sync"
)

type TeeMode string

const (
	TeeCapture     TeeMode = "capture"
	TeePassthrough TeeMode = "passthrough"
)

type TeeResult struct {
	ExitCode    int    `json:"exit_code"`
	Original    string `json:"original"`
	Compressed  string `json:"compressed"`
	SavedTokens int    `json:"saved_tokens"`
	Error       error  `json:"error,omitempty"`
}

type TeeExecutor struct {
	mode       TeeMode
	maxHistory int
	mu         sync.Mutex
	history    []TeeResult
}

func NewTeeExecutor() *TeeExecutor {
	return &TeeExecutor{
		mode:       TeeCapture,
		maxHistory: 100,
	}
}

func (t *TeeExecutor) Execute(args []string, compressFunc func(string) string) *TeeResult {
	result := &TeeResult{}

	cmd := exec.Command(args[0], args[1:]...)
	rawOutput, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}

	original := string(rawOutput)
	result.Original = original
	result.Compressed = compressFunc(original)
	result.SavedTokens = (len(original) - len(result.Compressed)) / 4
	result.Error = err

	t.mu.Lock()
	t.history = append(t.history, *result)
	if len(t.history) > t.maxHistory {
		t.history = t.history[1:]
	}
	t.mu.Unlock()

	return result
}

func (t *TeeExecutor) Recover(index int) string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if index >= 0 && index < len(t.history) {
		return t.history[index].Original
	}
	return ""
}

func (t *TeeExecutor) History() []TeeResult {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]TeeResult{}, t.history...)
}

type SavingsTracker struct {
	daily map[string]int64
	total int64
	mu    sync.Mutex
}

func NewSavingsTracker() *SavingsTracker {
	return &SavingsTracker{
		daily: make(map[string]int64),
	}
}

func (st *SavingsTracker) Record(savedTokens int) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.total += int64(savedTokens)
}

func (st *SavingsTracker) Total() int64 {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.total
}

func (st *SavingsTracker) RenderASCIIGraph(width int) string {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.total == 0 {
		return "No savings recorded yet.\n"
	}

	barLen := int(st.total) / 100
	if barLen > width-10 {
		barLen = width - 10
	}
	if barLen < 1 {
		barLen = 1
	}

	var sb strings.Builder
	sb.WriteString("TokMan Savings History\n")
	sb.WriteString(strings.Repeat("─", width) + "\n")
	sb.WriteString("│" + strings.Repeat("█", barLen))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", width) + "\n")
	sb.WriteString("Total: " + formatNumber(st.total) + " tokens saved\n")
	return sb.String()
}

func formatNumber(n int64) string {
	s := strings.Builder{}
	digits := []rune{}
	for n > 0 {
		digits = append([]rune{rune('0' + n%10)}, digits...)
		n /= 10
	}
	if len(digits) == 0 {
		return "0"
	}
	count := 0
	for i := len(digits) - 1; i >= 0; i-- {
		s.WriteRune(digits[i])
		count++
		if count%3 == 0 && i > 0 {
			s.WriteRune(',')
		}
	}
	return s.String()
}
