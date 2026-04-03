package web

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/httpmw"
)

const maxResponseBodySize = 10 * 1024 * 1024 // 10MB

var (
	apiProxyPort     int
	apiProxyUpstream string
	apiProxyAPIKey   string
)

// sharedPipeline is created once and reused across all API response compressions.
var sharedPipeline = filter.NewPipelineCoordinator(filter.PipelineConfig{
	Mode: filter.ModeMinimal, NgramEnabled: true,
	EnableCompaction: true, EnableAttribution: true,
})

var apiProxyCmd = &cobra.Command{
	Use:   "api-proxy",
	Short: "HTTP reverse proxy that compresses LLM API responses",
	Long: `Start a reverse proxy that sits between your app and LLM API.
Compresses tool output in API responses before they reach the model.

Example:
  tokman api-proxy --port 7878 --upstream https://api.anthropic.com
  export ANTHROPIC_BASE_URL=http://localhost:7878`,
	RunE: runAPIProxy,
}

func init() {
	apiProxyCmd.Flags().IntVar(&apiProxyPort, "port", 7878, "proxy listen port")
	apiProxyCmd.Flags().StringVar(&apiProxyUpstream, "upstream", "", "upstream API URL")
	apiProxyCmd.Flags().StringVar(&apiProxyAPIKey, "api-key", "", "optional API key for authenticating proxy requests")
	registry.Add(func() { registry.Register(apiProxyCmd) })
}

func runAPIProxy(cmd *cobra.Command, args []string) error {
	if apiProxyUpstream == "" {
		return fmt.Errorf("--upstream required")
	}
	if apiProxyPort < 1 || apiProxyPort > 65535 {
		return fmt.Errorf("invalid --port %d: must be between 1 and 65535", apiProxyPort)
	}

	upstream, err := url.Parse(apiProxyUpstream)
	if err != nil {
		return fmt.Errorf("invalid upstream URL: %w", err)
	}
	if upstream.Scheme != "http" && upstream.Scheme != "https" {
		return fmt.Errorf("invalid upstream URL: scheme must be http or https")
	}
	if upstream.Host == "" {
		return fmt.Errorf("invalid upstream URL: host required")
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = upstream.Scheme
			req.URL.Host = upstream.Host
			req.Host = upstream.Host
		},
		ModifyResponse: func(resp *http.Response) error {
			ct := resp.Header.Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				return nil
			}
			body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
			if err != nil {
				return err
			}
			resp.Body.Close()

			var data any
			if json.Unmarshal(body, &data) == nil {
				data = compressAPIData(data)
				newBody, err := json.Marshal(data)
				if err != nil {
					resp.Body = io.NopCloser(strings.NewReader(string(body)))
					return nil
				}
				resp.Body = io.NopCloser(strings.NewReader(string(newBody)))
				resp.ContentLength = int64(len(newBody))
				resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
			} else {
				resp.Body = io.NopCloser(strings.NewReader(string(body)))
			}
			return nil
		},
	}

	addr := fmt.Sprintf(":%d", apiProxyPort)
	fmt.Fprintf(os.Stderr, "tokman api-proxy on %s → %s\n", addr, apiProxyUpstream)

	rl := httpmw.NewDefault()

	var handler http.Handler = proxy
	if apiProxyAPIKey != "" {
		handler = authMiddleware(apiProxyAPIKey, proxy)
	}
	handler = rl.Middleware(handler)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	errCh := make(chan error, 1)
	go func() {
		if serveErr := srv.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-cmd.Context().Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	}
}

func authMiddleware(key string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+key)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func compressAPIData(data any) any {
	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any)
		for k, val := range v {
			if s, ok := val.(string); ok && len(s) > 500 {
				c, _ := sharedPipeline.Process(s)
				result[k] = c
			} else {
				result[k] = compressAPIData(val)
			}
		}
		return result
	case []any:
		for i, val := range v {
			v[i] = compressAPIData(val)
		}
		return v
	default:
		return v
	}
}
