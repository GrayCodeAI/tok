package filter

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// LLMCompressor uses an external LLM for semantic compression.
// Inspired by claw-compactor's Nexus and tamp's textpress.
type LLMCompressor struct {
	mu      sync.Mutex
	binPath string
	timeout int
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
func NewLLMCompressor(binPath string) *LLMCompressor {
	return &LLMCompressor{
		binPath: binPath,
		timeout: 30,
		enabled: binPath != "" && fileExists(binPath),
	}
}

// Compress uses the external LLM to semantically compress content.
func (lc *LLMCompressor) Compress(content string, maxTokens int) (string, int, int) {
	if !lc.enabled {
		return content, 0, 0
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	req := LLMCompressRequest{
		Content:   content,
		MaxTokens: maxTokens,
		Mode:      "compress",
	}
	reqBytes, _ := json.Marshal(req)

	cmd := exec.Command(lc.binPath)
	cmd.Stdin = strings.NewReader(string(reqBytes))
	cmd.Env = append(os.Environ(), "TOKMAN_LLM_TIMEOUT=30")

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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
