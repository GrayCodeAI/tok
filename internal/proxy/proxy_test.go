package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewProxy(t *testing.T) {
	p := NewProxy(":8080", "https://api.openai.com")
	if p == nil {
		t.Fatal("expected non-nil proxy")
	}
	if p.listenAddr != ":8080" {
		t.Errorf("expected listen addr :8080, got %s", p.listenAddr)
	}
	if p.targetURL != "https://api.openai.com" {
		t.Errorf("expected target URL, got %s", p.targetURL)
	}
}

func TestValidTargetURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "valid https", raw: "https://api.openai.com"},
		{name: "valid http", raw: "http://localhost:8080"},
		{name: "empty", raw: "", wantErr: true},
		{name: "missing scheme", raw: "api.openai.com", wantErr: true},
		{name: "bad scheme", raw: "ftp://example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidTargetURL(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidTargetURL(%q) error = %v, wantErr=%v", tt.raw, err, tt.wantErr)
			}
		})
	}
}

func TestDetectAPIFormat(t *testing.T) {
	tests := []struct {
		path     string
		expected APIFormat
	}{
		{"/v1/chat/completions", APIFormatOpenAI},
		{"/chat/completions", APIFormatOpenAI},
		{"/v1/messages", APIFormatAnthropic},
		{"/messages", APIFormatAnthropic},
		{"/v1beta/models/gemini-pro:generateContent", APIFormatGemini},
		{"/generateContent", APIFormatGemini},
		{"/unknown", APIFormatUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			got := detectAPIFormat(req)
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestCompressOpenAI(t *testing.T) {
	p := NewProxy(":8080", "https://api.openai.com")

	reqBody := map[string]any{
		"model": "gpt-4",
		"messages": []map[string]any{
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": strings.Repeat("Hello world. ", 100)},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	compressed, input, output := p.compressOpenAI(bodyBytes)
	if len(compressed) == 0 {
		t.Error("expected non-empty compressed body")
	}
	if input == 0 {
		t.Error("expected non-zero input tokens")
	}
	if output == 0 {
		t.Error("expected non-zero output tokens")
	}
	if output >= input {
		t.Errorf("expected output < input, got input=%d output=%d", input, output)
	}
}

func TestCompressAnthropic(t *testing.T) {
	p := NewProxy(":8080", "https://api.anthropic.com")

	reqBody := map[string]any{
		"model":  "claude-3-opus",
		"system": "You are a coding assistant.",
		"messages": []map[string]any{
			{"role": "user", "content": strings.Repeat("Test message. ", 100)},
		},
		"max_tokens": 4096,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	compressed, input, output := p.compressAnthropic(bodyBytes)
	if len(compressed) == 0 {
		t.Error("expected non-empty compressed body")
	}
	if input == 0 {
		t.Error("expected non-zero input tokens")
	}
	if output >= input {
		t.Errorf("expected output < input, got input=%d output=%d", input, output)
	}
}

func TestCompressGemini(t *testing.T) {
	p := NewProxy(":8080", "https://generativelanguage.googleapis.com")

	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]any{
					{"text": strings.Repeat("Gemini test content. ", 100)},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	compressed, input, output := p.compressGemini(bodyBytes)
	if len(compressed) == 0 {
		t.Error("expected non-empty compressed body")
	}
	if input == 0 {
		t.Error("expected non-zero input tokens")
	}
	if output >= input {
		t.Errorf("expected output < input, got input=%d output=%d", input, output)
	}
}

func TestSetModelAlias(t *testing.T) {
	p := NewProxy(":8080", "https://api.openai.com")
	p.SetModelAlias("gpt-4", "gpt-4o-mini")

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions",
		bytes.NewReader([]byte(`{"model":"gpt-4","messages":[]}`)))
	req.Header.Set("Content-Type", "application/json")

	result := p.applyModelAlias(req)
	bodyBytes, _ := io.ReadAll(result.Body)
	var body map[string]any
	json.Unmarshal(bodyBytes, &body)

	if body["model"] != "gpt-4o-mini" {
		t.Errorf("expected model gpt-4o-mini, got %v", body["model"])
	}
}

func TestRequestCache(t *testing.T) {
	rc := &requestCache{
		items: make(map[string]*cachedResponse),
		ttl:   5 * time.Minute,
	}

	key := "test-key"
	body := []byte(`{"response":"ok"}`)
	headers := http.Header{"Content-Type": []string{"application/json"}}

	rc.set(key, body, headers, 200)

	cached := rc.get(key)
	if cached == nil {
		t.Fatal("expected cached response")
	}
	if !bytes.Equal(cached.body, body) {
		t.Errorf("expected body %s, got %s", body, cached.body)
	}
	if cached.status != 200 {
		t.Errorf("expected status 200, got %d", cached.status)
	}
}

func TestRequestCacheExpiry(t *testing.T) {
	rc := &requestCache{
		items: make(map[string]*cachedResponse),
		ttl:   -1 * time.Second, // Already expired
	}

	key := "test-key"
	rc.set(key, []byte("body"), nil, 200)

	cached := rc.get(key)
	if cached != nil {
		t.Error("expected expired cache to return nil")
	}
}

func TestExtractModelName(t *testing.T) {
	tests := []struct {
		format   APIFormat
		body     string
		expected string
	}{
		{APIFormatOpenAI, `{"model":"gpt-4"}`, "gpt-4"},
		{APIFormatAnthropic, `{"model":"claude-3-opus"}`, "claude-3-opus"},
		{APIFormatUnknown, `{"model":"unknown"}`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := extractModelName([]byte(tt.body), tt.format)
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestProxyStats(t *testing.T) {
	stats := &ProxyStats{
		ByFormat: make(map[APIFormat]int64),
		ByModel:  make(map[string]int64),
	}

	stats.record(APIFormatOpenAI, 1000, 500, 500, "gpt-4")
	stats.record(APIFormatOpenAI, 2000, 1000, 1000, "gpt-4")
	stats.record(APIFormatAnthropic, 1500, 750, 750, "claude-3")

	if stats.TotalRequests != 3 {
		t.Errorf("expected 3 requests, got %d", stats.TotalRequests)
	}
	if stats.TotalInputTokens != 4500 {
		t.Errorf("expected 4500 input tokens, got %d", stats.TotalInputTokens)
	}
	if stats.TotalSavedTokens != 2250 {
		t.Errorf("expected 2250 saved tokens, got %d", stats.TotalSavedTokens)
	}
	if stats.ByModel["gpt-4"] != 2 {
		t.Errorf("expected 2 gpt-4 requests, got %d", stats.ByModel["gpt-4"])
	}
}

func TestHealthEndpoint(t *testing.T) {
	p := NewProxy(":8080", "https://api.openai.com")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}
}

func TestMetricsEndpoint(t *testing.T) {
	p := NewProxy(":8080", "https://api.openai.com")
	// Record some stats first
	p.stats.record(APIFormatOpenAI, 1000, 500, 500, "gpt-4")

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if int(resp["total_requests"].(float64)) != 1 {
		t.Errorf("expected 1 request, got %v", resp["total_requests"])
	}
}

func TestMetricsEndpointZeroInputDoesNotNaN(t *testing.T) {
	p := NewProxy(":8080", "https://api.openai.com")

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if resp["savings_percent"] != float64(0) {
		t.Fatalf("expected savings_percent 0, got %v", resp["savings_percent"])
	}
}

func TestRequestCacheClonesHeadersAndBody(t *testing.T) {
	rc := &requestCache{
		items: make(map[string]*cachedResponse),
		ttl:   5 * time.Minute,
	}

	headers := http.Header{"X-Test": []string{"a"}}
	body := []byte("body")
	rc.set("k", body, headers, http.StatusOK)

	headers.Set("X-Test", "b")
	body[0] = 'B'

	cached := rc.get("k")
	if cached == nil {
		t.Fatal("expected cached response")
	}
	if got := cached.headers.Get("X-Test"); got != "a" {
		t.Fatalf("cached header = %q, want %q", got, "a")
	}
	if string(cached.body) != "body" {
		t.Fatalf("cached body = %q, want %q", string(cached.body), "body")
	}
}

func TestCompressOpenAI_InvalidJSON(t *testing.T) {
	p := NewProxy(":8080", "https://api.openai.com")

	bodyBytes := []byte(`{invalid json}`)
	compressed, input, output := p.compressOpenAI(bodyBytes)

	if !bytes.Equal(compressed, bodyBytes) {
		t.Error("expected unchanged body for invalid JSON")
	}
	if input != 0 || output != 0 {
		t.Error("expected zero tokens for invalid JSON")
	}
}
