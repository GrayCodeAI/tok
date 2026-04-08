package client

import (
	"context"
	"time"
)

type CompressionResult struct {
	Output           string
	OriginalTokens   int
	CompressedTokens int
	SavingsPercent   float64
}

type Client struct {
	config      *Config
	compression *CompressionClient
}

type Config struct {
	Address         string
	Timeout         time.Duration
	RetryCount      int
	CompressionAddr string
	AnalyticsAddr   string
}

type CompressionClient struct {
	client *Client
}

func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	return &Client{config: cfg}, nil
}

func NewWithConfig(cfg *Config) (*Client, error) {
	return New(cfg)
}

func (c *Client) Compression() *CompressionClient {
	if c.compression == nil {
		c.compression = &CompressionClient{client: c}
	}
	return c.compression
}

func (c *CompressionClient) Compress(ctx context.Context, input, mode string, budget int) (*CompressionResult, error) {
	return &CompressionResult{
		Output:           input,
		OriginalTokens:   len(input) / 4,
		CompressedTokens: len(input) / 4,
		SavingsPercent:   0,
	}, nil
}

func (c *CompressionClient) Decompress(data []byte) ([]byte, error) {
	return data, nil
}
