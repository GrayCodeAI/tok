// Package main demonstrates TokMan integration with LangChain
// 
// This example shows how to use TokMan as a context compressor
// before sending prompts to LLMs via LangChain.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// MockLLMClient simulates a LangChain LLM client
type MockLLMClient struct {
	MaxContextTokens int
	CallCount        int
}

// Call simulates an LLM call with token counting
func (m *MockLLMClient) Call(ctx context.Context, prompt string) (string, error) {
	m.CallCount++
	
	// Count tokens (simplified - real implementation would use tiktoken)
	tokenCount := len(strings.Fields(prompt))
	
	if tokenCount > m.MaxContextTokens {
		return "", fmt.Errorf("context length exceeded: %d > %d", tokenCount, m.MaxContextTokens)
	}
	
	return fmt.Sprintf("Response to: %s...", prompt[:min(50, len(prompt))]), nil
}

// TokManCompressor wraps TokMan for LangChain integration
type TokManCompressor struct {
	pipeline *filter.Pipeline
	mode     string
}

// NewTokManCompressor creates a new compressor with the given mode
func NewTokManCompressor(mode string) *TokManCompressor {
	return &TokManCompressor{
		pipeline: filter.NewPipeline(),
		mode:     mode,
	}
}

// Compress reduces the token count of the input text
func (c *TokManCompressor) Compress(ctx context.Context, text string) (string, int, int, error) {
	originalTokens := filter.EstimateTokens(text)
	
	result, err := c.pipeline.Process(text, filter.PipelineConfig{
		Mode: c.mode,
	})
	if err != nil {
		return "", 0, 0, err
	}
	
	return result.Output, originalTokens, result.FinalTokens, nil
}

// Chain represents a simple LLM chain
type Chain struct {
	LLM       *MockLLMClient
	Compressor *TokManCompressor
}

// Run executes the chain with automatic compression
func (c *Chain) Run(ctx context.Context, input string) (string, error) {
	// Estimate original tokens
	originalTokens := filter.EstimateTokens(input)
	
	// Compress if needed
	compressed := input
	finalTokens := originalTokens
	
	if originalTokens > c.LLM.MaxContextTokens {
		var err error
		compressed, _, finalTokens, err = c.Compressor.Compress(ctx, input)
		if err != nil {
			return "", fmt.Errorf("compression failed: %w", err)
		}
		log.Printf("Compressed: %d -> %d tokens (%.1f%% reduction)",
			originalTokens, finalTokens,
			float64(originalTokens-finalTokens)/float64(originalTokens)*100)
	}
	
	// Call LLM with compressed context
	return c.LLM.Call(ctx, compressed)
}

func main() {
	fmt.Println("=== TokMan + LangChain Integration Example ===\n")
	
	// Create a mock LLM with limited context
	llm := &MockLLMClient{MaxContextTokens: 100}
	
	// Create compressor
	compressor := NewTokManCompressor("balanced")
	
	// Create chain
	chain := &Chain{
		LLM:        llm,
		Compressor: compressor,
	}
	
	// Example 1: Small prompt (no compression needed)
	smallPrompt := "What is 2 + 2?"
	fmt.Printf("Small prompt: %q\n", smallPrompt)
	response, err := chain.Run(context.Background(), smallPrompt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n\n", response)
	
	// Example 2: Large document (compression applied)
	largeDocument := generateLargeDocument()
	fmt.Printf("Large document: %d tokens\n", filter.EstimateTokens(largeDocument))
	response, err = chain.Run(context.Background(), largeDocument)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n\n", response)
	
	// Example 3: Code compression
	codeContext := generateCodeContext()
	fmt.Printf("Code context: %d tokens\n", filter.EstimateTokens(codeContext))
	response, err = chain.Run(context.Background(), codeContext)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Response: %s\n", response)
	
	fmt.Printf("\nTotal LLM calls: %d\n", llm.CallCount)
}

func generateLargeDocument() string {
	var sb strings.Builder
	sb.WriteString("DOCUMENT: Software Architecture Best Practices\n\n")
	
	for i := 1; i <= 50; i++ {
		sb.WriteString(fmt.Sprintf(`
Section %d: Design Principles
When designing software systems, it is important to consider multiple factors
including scalability, maintainability, and performance. The following principles
should guide your architectural decisions:

1. Separation of Concerns - Each module should have a single responsibility
2. Don't Repeat Yourself (DRY) - Avoid code duplication
3. Keep It Simple, Stupid (KISS) - Simplicity should be a key goal
4. You Ain't Gonna Need It (YAGNI) - Don't implement unused features

These principles help create maintainable and scalable software systems.
`, i))
	}
	
	return sb.String()
}

func generateCodeContext() string {
	return `
// UserService handles user-related operations
type UserService struct {
    db     *Database
    cache  *Cache
    logger *Logger
}

// CreateUser creates a new user in the system
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Check if user exists
    exists, err := s.db.UserExists(ctx, req.Email)
    if err != nil {
        s.logger.Error("failed to check user existence", "error", err)
        return nil, err
    }
    if exists {
        return nil, ErrUserAlreadyExists
    }
    
    // Create user
    user := &User{
        ID:        uuid.New(),
        Email:     req.Email,
        Name:      req.Name,
        CreatedAt: time.Now(),
    }
    
    if err := s.db.CreateUser(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    // Cache the user
    s.cache.Set(ctx, user.ID, user, time.Hour)
    
    return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    // Check cache first
    if user, ok := s.cache.Get(ctx, id); ok {
        return user.(*User), nil
    }
    
    // Fetch from database
    user, err := s.db.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache for future requests
    s.cache.Set(ctx, id, user, time.Hour)
    
    return user, nil
}

Question: What design patterns are used in this code?
`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
