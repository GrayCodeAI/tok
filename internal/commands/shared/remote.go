package shared

import (
	"context"
	"fmt"
	"os"
	"time"
)

// RemoteConfig holds remote client configuration.
type RemoteConfig struct {
	CompressionAddr string
	AnalyticsAddr   string
	Timeout         time.Duration
}

// RemoteClient is a stub for the remote client (functionality removed).
type RemoteClient struct{}

// CompressionClient is a stub for compression operations.
type CompressionClient struct{}

// Compress performs compression (stub - returns input unchanged).
func (c *CompressionClient) Compress(ctx context.Context, input, mode string, budget int) (*CompressionResult, error) {
	return &CompressionResult{
		Output:           input,
		OriginalTokens:   len(input) / 4,
		CompressedTokens: len(input) / 4,
		SavingsPercent:   0,
	}, nil
}

// Compression returns a compression client (stub).
func (c *RemoteClient) Compression() *CompressionClient {
	return &CompressionClient{}
}

// CompressionResult holds compression results.
type CompressionResult struct {
	Output           string
	OriginalTokens   int
	CompressedTokens int
	SavingsPercent   float64
}

// NewRemoteClient creates a new remote client (stub - returns nil).
func NewRemoteClient(cfg *RemoteConfig) (*RemoteClient, error) {
	return nil, fmt.Errorf("remote mode not implemented")
}

// Remote executor for gRPC-based compression and analytics.
// This is used when --remote flag is enabled to offload processing
// to tok microservices.

var (
	globalRemoteClient *RemoteClient
	remoteClientErr    error
)

// GetRemoteClient returns the global gRPC client (lazy initialization).
// Returns nil if remote mode is not enabled or connection failed.
func GetRemoteClient() *RemoteClient {
	if !IsRemoteMode() {
		return nil
	}

	if globalRemoteClient != nil {
		return globalRemoteClient
	}

	cfg := &RemoteConfig{
		CompressionAddr: GetCompressionAddr(),
		AnalyticsAddr:   GetAnalyticsAddr(),
		Timeout:         time.Duration(GetRemoteTimeout()) * time.Second,
	}

	// Set defaults if not specified
	if cfg.CompressionAddr == "" {
		cfg.CompressionAddr = "localhost:50051"
	}
	if cfg.AnalyticsAddr == "" {
		cfg.AnalyticsAddr = "localhost:50053"
	}

	globalRemoteClient, remoteClientErr = NewRemoteClient(cfg)
	if remoteClientErr != nil {
		if IsVerbose() {
			fmt.Fprintf(stderrWriter(), "[remote] failed to connect: %v\n", remoteClientErr)
		}
		return nil
	}

	return globalRemoteClient
}

// RemoteCompress performs compression via gRPC service.
// Returns the compressed output and tokens saved, or an error.
func RemoteCompress(input, mode string, budget int) (output string, tokensSaved int, err error) {
	c := GetRemoteClient()
	if c == nil {
		return "", 0, fmt.Errorf("remote client not available")
	}

	comp := c.Compression()
	if comp == nil {
		return "", 0, fmt.Errorf("remote compression client not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(GetRemoteTimeout())*time.Second)
	defer cancel()

	result, err := comp.Compress(ctx, input, mode, budget)
	if err != nil {
		return "", 0, fmt.Errorf("remote compression failed: %w", err)
	}

	tokensSaved = result.OriginalTokens - result.CompressedTokens
	if tokensSaved < 0 {
		tokensSaved = 0
	}

	if IsVerbose() {
		fmt.Fprintf(stderrWriter(), "[remote] compress done: %d -> %d tokens (%.1f%% saved)\n",
			result.OriginalTokens, result.CompressedTokens, result.SavingsPercent)
	}

	return result.Output, tokensSaved, nil
}

// RemoteRecordAnalytics sends command metrics to the analytics service.
func RemoteRecordAnalytics(command string, originalTokens, filteredTokens int, execTimeMs int64, success bool) error {
	// Analytics recording via remote service is optional
	// The analytics service aggregates from its own tracking
	// This is a placeholder for future direct push analytics
	return nil
}

// stderrWriter returns the stderr writer.
func stderrWriter() *os.File {
	return os.Stderr
}
