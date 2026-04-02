// Package testing provides mocks for testing.
package testing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// MockWriter is a mock io.Writer for testing.
type MockWriter struct {
	mu       sync.Mutex
	buffer   bytes.Buffer
	writeErr error
}

// NewMockWriter creates a new mock writer.
func NewMockWriter() *MockWriter {
	return &MockWriter{}
}

// Write implements io.Writer.
func (m *MockWriter) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.buffer.Write(p)
}

// SetError sets an error to return on write.
func (m *MockWriter) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeErr = err
}

// String returns the written content.
func (m *MockWriter) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.buffer.String()
}

// Bytes returns the written bytes.
func (m *MockWriter) Bytes() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.buffer.Bytes()
}

// Reset clears the buffer.
func (m *MockWriter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.buffer.Reset()
}

// MockReader is a mock io.Reader for testing.
type MockReader struct {
	data     []byte
	pos      int
	readErr  error
	delay    time.Duration
}

// NewMockReader creates a new mock reader.
func NewMockReader(data string) *MockReader {
	return &MockReader{data: []byte(data)}
}

// Read implements io.Reader.
func (m *MockReader) Read(p []byte) (n int, err error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.readErr != nil {
		return 0, m.readErr
	}
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

// SetError sets an error to return on read.
func (m *MockReader) SetError(err error) {
	m.readErr = err
}

// SetDelay sets a delay for each read.
func (m *MockReader) SetDelay(d time.Duration) {
	m.delay = d
}

// MockConn is a mock network connection.
type MockConn struct {
	readBuf  bytes.Buffer
	writeBuf bytes.Buffer
	closed   bool
	mu       sync.Mutex
}

// NewMockConn creates a new mock connection.
func NewMockConn() *MockConn {
	return &MockConn{}
}

// Read reads from the connection.
func (m *MockConn) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.EOF
	}
	return m.readBuf.Read(p)
}

// Write writes to the connection.
func (m *MockConn) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.writeBuf.Write(p)
}

// Close closes the connection.
func (m *MockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// FeedData adds data to be read.
func (m *MockConn) FeedData(data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readBuf.Write(data)
}

// WrittenData returns data that was written.
func (m *MockConn) WrittenData() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.writeBuf.Bytes()
}

// MockCache is a mock cache implementation.
type MockCache struct {
	mu      sync.RWMutex
	data    map[string][]byte
	hits    int
	misses  int
	getErr  error
	setErr  error
}

// NewMockCache creates a new mock cache.
func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string][]byte),
	}
}

// Get retrieves a value from the cache.
func (m *MockCache) Get(key string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getErr != nil {
		return nil, false
	}
	val, ok := m.data[key]
	if ok {
		m.hits++
	} else {
		m.misses++
	}
	return val, ok
}

// Set stores a value in the cache.
func (m *MockCache) Set(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.setErr != nil {
		return m.setErr
	}
	m.data[key] = value
	return nil
}

// Delete removes a value from the cache.
func (m *MockCache) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

// Clear clears all values.
func (m *MockCache) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string][]byte)
	m.hits = 0
	m.misses = 0
}

// Stats returns cache statistics.
func (m *MockCache) Stats() (hits, misses int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.hits, m.misses
}

// SetGetError sets the error to return on Get.
func (m *MockCache) SetGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getErr = err
}

// SetSetError sets the error to return on Set.
func (m *MockCache) SetSetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setErr = err
}

// MockCompressor is a mock compressor.
type MockCompressor struct {
	compressRatio   float64
	compressErr     error
	decompressErr   error
	callCount       int
	compressedSizes []int
}

// NewMockCompressor creates a new mock compressor.
func NewMockCompressor(ratio float64) *MockCompressor {
	return &MockCompressor{
		compressRatio:   ratio,
		compressedSizes: make([]int, 0),
	}
}

// Compress compresses content.
func (m *MockCompressor) Compress(content string) (string, int, error) {
	m.callCount++
	if m.compressErr != nil {
		return "", 0, m.compressErr
	}
	compressedLen := int(float64(len(content)) * m.compressRatio)
	saved := len(content) - compressedLen
	m.compressedSizes = append(m.compressedSizes, compressedLen)
	return content[:compressedLen], saved, nil
}

// Decompress decompresses content.
func (m *MockCompressor) Decompress(content string) (string, error) {
	if m.decompressErr != nil {
		return "", m.decompressErr
	}
	return content, nil
}

// SetCompressError sets the compression error.
func (m *MockCompressor) SetCompressError(err error) {
	m.compressErr = err
}

// SetDecompressError sets the decompression error.
func (m *MockCompressor) SetDecompressError(err error) {
	m.decompressErr = err
}

// CallCount returns the number of compression calls.
func (m *MockCompressor) CallCount() int {
	return m.callCount
}

// MockLogger is a mock logger.
type MockLogger struct {
	mu      sync.Mutex
	logs    []LogEntry
	enabled bool
}

// LogEntry represents a log entry.
type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]interface{}
	Time    time.Time
}

// NewMockLogger creates a new mock logger.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		logs:    make([]LogEntry, 0),
		enabled: true,
	}
}

// Debug logs a debug message.
func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.log("DEBUG", msg, fields...)
}

// Info logs an info message.
func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.log("INFO", msg, fields...)
}

// Warn logs a warning message.
func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.log("WARN", msg, fields...)
}

// Error logs an error message.
func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.log("ERROR", msg, fields...)
}

func (m *MockLogger) log(level, msg string, fields ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.enabled {
		return
	}
	entry := LogEntry{
		Level:   level,
		Message: msg,
		Time:    time.Now(),
		Fields:  make(map[string]interface{}),
	}
	// Simple field parsing (key, value pairs)
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			entry.Fields[key] = fields[i+1]
		}
	}
	m.logs = append(m.logs, entry)
}

// Logs returns all logged entries.
func (m *MockLogger) Logs() []LogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]LogEntry, len(m.logs))
	copy(result, m.logs)
	return result
}

// HasLog checks if a log message exists.
func (m *MockLogger) HasLog(level, contains string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.logs {
		if entry.Level == level && containsStr(entry.Message, contains) {
			return true
		}
	}
	return false
}

// Clear clears all logs.
func (m *MockLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = m.logs[:0]
}

// SetEnabled enables/disables logging.
func (m *MockLogger) SetEnabled(enabled bool) {
	m.enabled = enabled
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstr(s, substr)))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// MockHTTPClient is a mock HTTP client.
type MockHTTPClient struct {
	responses map[string]MockResponse
	requests  []MockRequest
	mu        sync.Mutex
}

// MockResponse is a mock HTTP response.
type MockResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
	Err        error
}

// MockRequest captures request details.
type MockRequest struct {
	Method string
	URL    string
	Body   []byte
}

// NewMockHTTPClient creates a new mock HTTP client.
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses: make(map[string]MockResponse),
		requests:  make([]MockRequest, 0),
	}
}

// AddResponse adds a response for a URL.
func (m *MockHTTPClient) AddResponse(method, url string, resp MockResponse) {
	key := method + " " + url
	m.responses[key] = resp
}

// Do executes an HTTP request.
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	body, _ := io.ReadAll(req.Body)
	m.requests = append(m.requests, MockRequest{
		Method: req.Method,
		URL:    req.URL.String(),
		Body:   body,
	})

	key := req.Method + " " + req.URL.String()
	resp, ok := m.responses[key]
	if !ok {
		return nil, fmt.Errorf("no mock response for %s", key)
	}
	if resp.Err != nil {
		return nil, resp.Err
	}

	httpResp := &http.Response{
		StatusCode: resp.StatusCode,
		Body:       io.NopCloser(bytes.NewReader(resp.Body)),
		Header:     make(http.Header),
	}
	for k, v := range resp.Headers {
		httpResp.Header.Set(k, v)
	}
	return httpResp, nil
}

// Requests returns captured requests.
func (m *MockHTTPClient) Requests() []MockRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockRequest, len(m.requests))
	copy(result, m.requests)
	return result
}

// MockCommandRunner mocks command execution.
type MockCommandRunner struct {
	commands map[string]MockCommandResult
	mu       sync.Mutex
}

// MockCommandResult is the result of a command.
type MockCommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// NewMockCommandRunner creates a new mock command runner.
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		commands: make(map[string]MockCommandResult),
	}
}

// AddCommand adds a command result.
func (m *MockCommandRunner) AddCommand(name string, result MockCommandResult) {
	m.commands[name] = result
}

// Run runs a command.
func (m *MockCommandRunner) Run(ctx context.Context, name string, args ...string) (string, string, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd := name
	if len(args) > 0 {
		cmd += " " + strings.Join(args, " ")
	}

	result, ok := m.commands[cmd]
	if !ok {
		return "", "", 127, fmt.Errorf("command not found: %s", cmd)
	}
	return result.Stdout, result.Stderr, result.ExitCode, result.Err
}

// MockContext provides a mock context for testing.
type MockContext struct {
	context.Context
	values map[interface{}]interface{}
	mu     sync.RWMutex
	done   chan struct{}
}

// NewMockContext creates a new mock context.
func NewMockContext() *MockContext {
	return &MockContext{
		Context: context.Background(),
		values:  make(map[interface{}]interface{}),
		done:    make(chan struct{}),
	}
}

// Value returns a value from the context.
func (m *MockContext) Value(key interface{}) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.values[key]
}

// SetValue sets a value in the context.
func (m *MockContext) SetValue(key, val interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[key] = val
}

// Done returns the done channel.
func (m *MockContext) Done() <-chan struct{} {
	return m.done
}

// Cancel cancels the context.
func (m *MockContext) Cancel() {
	close(m.done)
}
