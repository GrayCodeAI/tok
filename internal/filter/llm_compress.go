package filter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LLMCompressor uses an external LLM for semantic compression.
// Inspired by claw-compactor's Nexus and tamp's textpress.
type LLMCompressor struct {
	mu      sync.Mutex
	binPath string
	timeout time.Duration
	enabled bool
}

// LLMCompressRequest is the JSON input for LLM compression.
type LLMCompressRequest struct {
	Content   string `json:"content"`
	MaxTokens int    `json:"max_tokens"`
	Mode      string `json:"mode"`
}

// LLMCompressResponse is the JSON output from LLM compression.
type LLMCompressResponse struct {
	Compressed string `json:"compressed"`
	TokensIn   int    `json:"tokens_in"`
	TokensOut  int    `json:"tokens_out"`
}

// NewLLMCompressor creates a new LLM-based compressor.
// binPath must be an absolute path to an executable file.
func NewLLMCompressor(binPath string) *LLMCompressor {
	return &LLMCompressor{
		binPath: binPath,
		timeout: llmTimeoutFromEnv(),
		enabled: isExecutable(binPath),
	}
}

// Compress uses the external LLM to semantically compress content.
func (lc *LLMCompressor) Compress(content string, maxTokens int) (string, int, int) {
	if !lc.enabled {
		return content, 0, 0
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	reqBytes, err := json.Marshal(LLMCompressRequest{
		Content:   content,
		MaxTokens: maxTokens,
		Mode:      "compress",
	})
	if err != nil {
		return content, 0, 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), lc.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, lc.binPath)
	cmd.Stdin = strings.NewReader(string(reqBytes))
	cmd.Env = append(os.Environ(), fmt.Sprintf("TOK_LLM_TIMEOUT=%d", int(lc.timeout.Seconds())))

	out, err := cmd.Output()
	if err != nil {
		return content, 0, 0
	}

	var resp LLMCompressResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return content, 0, 0
	}

	saved := resp.TokensIn - resp.TokensOut
	return resp.Compressed, saved, resp.TokensIn
}

// IsEnabled returns whether LLM compression is available.
func (lc *LLMCompressor) IsEnabled() bool {
	return lc.enabled
}

// SetEnabled toggles LLM compression.
func (lc *LLMCompressor) SetEnabled(enabled bool) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.enabled = enabled
}

// isExecutable reports whether path is an absolute path to a regular executable file.
func isExecutable(path string) bool {
	if path == "" || !filepath.IsAbs(path) {
		return false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode()&0111 != 0
}

// llmTimeoutFromEnv reads TOK_LLM_TIMEOUT (seconds); defaults to 30s.
func llmTimeoutFromEnv() time.Duration {
	if v := os.Getenv("TOK_LLM_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 30 * time.Second
}
