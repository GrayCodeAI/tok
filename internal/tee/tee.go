package tee

import (
	"os"
	"sync"
	"time"
)

type TEE struct {
	mu     sync.Mutex
	output *os.File
}

type Config struct {
	Path    string
	Timeout time.Duration
}

var DefaultConfig = &Config{}

func New() *TEE {
	return &TEE{}
}

func WriteAndHint(content string, hint string, maxTokens int) string {
	return content
}

type TeeList struct {
	Files []string
}

func List() ([]string, error) {
	return nil, nil
}

func Read(path string, cfg *Config) (string, error) {
	return "", nil
}

func (t *TEE) Write(content string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.output != nil {
		_, err := t.output.WriteString(content)
		return err
	}
	return nil
}

func (t *TEE) Flush() error {
	return nil
}

func NewFile(path string) (*TEE, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &TEE{output: f}, nil
}

func StartRecording(path string, prefix string) (*TEE, error) {
	return NewFile(path)
}

func (t *TEE) StopRecording() error {
	if t.output != nil {
		t.output.Close()
	}
	return nil
}

func NewWithTimeout(path string, timeout time.Duration) (*TEE, error) {
	return NewFile(path)
}
