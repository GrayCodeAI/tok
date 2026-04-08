package shared

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/GrayCodeAI/tokman/pkg/client"
)

// Remote executor for gRPC-based compression and analytics.
// This is used when --remote flag is enabled to offload processing
// to TokMan microservices.

var (
	globalRemoteClient *client.Client
	remoteClientErr    error
)

// GetRemoteClient returns the global gRPC client (lazy initialization).
// Returns nil if remote mode is not enabled or connection failed.
func GetRemoteClient() *client.Client {
	if !IsRemoteMode() {
		return nil
	}

	if globalRemoteClient != nil {
		return globalRemoteClient
	}

	cfg := &client.Config{
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

	globalRemoteClient, remoteClientErr = client.New(cfg)
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
