package httpproxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type ProxyConfig struct {
	ListenAddr   string
	UpstreamURL  string
	CompressFunc func(string) string
	Timeout      time.Duration
	MaxBodySize  int64
}

type HTTPProxy struct {
	config *ProxyConfig
	proxy  *httputil.ReverseProxy
	server *http.Server
}

func NewHTTPProxy(config *ProxyConfig) *HTTPProxy {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 10 * 1024 * 1024
	}

	upstream, _ := url.Parse(config.UpstreamURL)
	proxy := httputil.NewSingleHostReverseProxy(upstream)

	return &HTTPProxy{
		config: config,
		proxy:  proxy,
	}
}

func (p *HTTPProxy) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleRequest)

	p.server = &http.Server{
		Addr:    p.config.ListenAddr,
		Handler: mux,
	}

	return p.server.ListenAndServe()
}

func (p *HTTPProxy) Stop() error {
	if p.server != nil {
		return p.server.Close()
	}
	return nil
}

func (p *HTTPProxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	if p.config.CompressFunc != nil {
		wrapped := &responseWriter{ResponseWriter: w, compressFunc: p.config.CompressFunc}
		p.proxy.ServeHTTP(wrapped, r)
	} else {
		p.proxy.ServeHTTP(w, r)
	}
}

type responseWriter struct {
	http.ResponseWriter
	compressFunc func(string) string
	written      strings.Builder
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.written.Write(b)
	compressed := w.compressFunc(w.written.String())
	return w.ResponseWriter.Write([]byte(compressed))
}

type AdaptiveScalingConfig struct {
	ShortThreshold int
	LongThreshold  int
	ShortMode      string
	LongMode       string
}

type AdaptiveScaler struct {
	config *AdaptiveScalingConfig
}

func NewAdaptiveScaler() *AdaptiveScaler {
	return &AdaptiveScaler{
		config: &AdaptiveScalingConfig{
			ShortThreshold: 500,
			LongThreshold:  5000,
			ShortMode:      "surface",
			LongMode:       "core",
		},
	}
}

func (a *AdaptiveScaler) GetMode(input string) string {
	tokens := len(input) / 4
	if tokens < a.config.ShortThreshold {
		return a.config.ShortMode
	}
	if tokens > a.config.LongThreshold {
		return a.config.LongMode
	}
	return "trim"
}

type ModelFallbackManager struct {
	primary   string
	fallbacks []string
}

func NewModelFallbackManager(primary string, fallbacks ...string) *ModelFallbackManager {
	return &ModelFallbackManager{
		primary:   primary,
		fallbacks: fallbacks,
	}
}

func (m *ModelFallbackManager) GetModel(statusCode int) string {
	if statusCode == 429 || statusCode == 503 {
		for _, fb := range m.fallbacks {
			return fb
		}
	}
	return m.primary
}

func (m *ModelFallbackManager) GetAll() []string {
	return append([]string{m.primary}, m.fallbacks...)
}

type OpenTelemetryConfig struct {
	ServiceName string
	Endpoint    string
	SampleRate  float64
}

type OpenTelemetryCollector struct {
	config *OpenTelemetryConfig
}

func NewOpenTelemetryCollector(config *OpenTelemetryConfig) *OpenTelemetryCollector {
	if config.SampleRate == 0 {
		config.SampleRate = 1.0
	}
	return &OpenTelemetryCollector{config: config}
}

func (o *OpenTelemetryCollector) RecordMetric(name string, value float64, labels map[string]string) {
	metric := fmt.Sprintf("%s=%v labels=%v", name, value, labels)
	_ = metric
}

func (o *OpenTelemetryCollector) ExportMetrics() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Service: %s\n", o.config.ServiceName))
	sb.WriteString(fmt.Sprintf("# Endpoint: %s\n", o.config.Endpoint))
	return sb.String()
}
