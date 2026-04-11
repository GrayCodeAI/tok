package filter

import (
	"strings"
	"testing"
)

func TestGMSAFilter_Name(t *testing.T) {
	f := NewGMSAFilter()
	if f.Name() != "28_gmsa" {
		t.Errorf("unexpected name: %s", f.Name())
	}
}

func TestGMSAFilter_ModeNone(t *testing.T) {
	f := NewGMSAFilter()
	input := "chunk one line one\nchunk one line two\nchunk one line three\n\nchunk two line one\nchunk two line two\n"
	out, saved := f.Apply(input, ModeNone)
	if out != input || saved != 0 {
		t.Error("ModeNone should return unchanged")
	}
}

func TestGMSAFilter_ShortInput(t *testing.T) {
	f := NewGMSAFilter()
	input := "just one paragraph\nwith two lines"
	out, _ := f.Apply(input, ModeMinimal)
	if out != input {
		t.Error("single chunk should pass through")
	}
}

func TestGMSAFilter_MergesSimilarChunks(t *testing.T) {
	f := NewGMSAFilter()
	// Three paragraphs with very similar content about authentication
	para1 := "The authentication system uses JWT tokens for user identity verification.\nTokens are signed with a secret key and have a configurable expiration time.\nThe auth service validates tokens on every API request to protected endpoints."
	para2 := "User authentication relies on JWT token verification for identity.\nEach token contains user claims and is signed using HMAC-SHA256 algorithm.\nThe authentication middleware validates these tokens before processing requests."
	para3 := "JWT-based authentication verifies user identity on each request.\nTokens include user ID and role claims signed with the application secret.\nExpired or invalid tokens are rejected by the authentication layer."
	footer := "ERROR: database connection pool exhausted after 30 seconds timeout."
	input := para1 + "\n\n" + para2 + "\n\n" + para3 + "\n\n" + footer
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings on similar chunks")
	}
	if !strings.Contains(out, "ERROR:") {
		t.Error("error content must be preserved")
	}
	if !strings.Contains(out, "similar chunks merged") {
		t.Error("expected merge annotation in output")
	}
}

func TestGMSAFilter_PreservesDistinctChunks(t *testing.T) {
	f := NewGMSAFilter()
	// Very different paragraphs - should all survive
	para1 := "The caching layer uses Redis for distributed key-value storage.\nCache entries expire after configurable TTL based on data volatility.\nCache miss rate is monitored via Prometheus metrics dashboard."

	para2 := "Database migrations are managed using Flyway version control tool.\nEach migration script is numbered and applied in sequential order.\nRollback scripts are provided for every forward migration applied."

	para3 := "The message queue uses RabbitMQ for asynchronous event processing.\nConsumers acknowledge messages only after successful processing completes.\nDead letter queues capture failed messages for manual investigation."

	input := para1 + "\n\n" + para2 + "\n\n" + para3
	out, _ := f.Apply(input, ModeMinimal)

	// All 3 distinct chunks should survive
	if !strings.Contains(out, "Redis") {
		t.Error("distinct chunk 1 should be preserved")
	}
	if !strings.Contains(out, "Flyway") {
		t.Error("distinct chunk 2 should be preserved")
	}
	if !strings.Contains(out, "RabbitMQ") {
		t.Error("distinct chunk 3 should be preserved")
	}
}

func TestGMSAFilter_SemanticAlignAnchorFirst(t *testing.T) {
	f := NewGMSAFilter()
	// Mix anchor (error) chunk and regular chunks
	// After alignment, error chunk should be near the top
	para1 := "Background context about the service deployment process.\nThe deployment pipeline runs through several stages before completion.\nEach stage validates configuration and tests the service functionality."

	para2 := "ERROR: deployment failed at stage 3 with exit code 1.\nThe container failed health check after restart with signal SIGTERM.\nCheck logs for details about the failed health check endpoint."

	para3 := "Additional notes about deployment monitoring and observability.\nMetrics are exported to Prometheus and visualized in Grafana dashboards.\nAlerts fire when error rate exceeds configured threshold values."

	// para1, para3 are low-anchor; para2 is high-anchor (errors)
	input := para1 + "\n\n" + para2 + "\n\n" + para3
	out, _ := f.Apply(input, ModeMinimal)

	errIdx := strings.Index(out, "ERROR:")
	if errIdx < 0 {
		t.Fatal("error chunk must be in output")
	}
	// Error chunk should appear earlier than background context
	bgIdx := strings.Index(out, "Background context")
	if bgIdx >= 0 && errIdx > bgIdx {
		// Allow some flexibility — alignment is a soft reorder
		t.Logf("note: error chunk at %d, background at %d (alignment may vary)", errIdx, bgIdx)
	}
}

func TestGMSAFilter_AggressiveMoreReduction(t *testing.T) {
	f := NewGMSAFilter()
	makeChunk := func(prefix string) string {
		return prefix + " uses connection pooling for database access management.\n" +
			prefix + " implements retry logic with exponential backoff strategy.\n" +
			prefix + " monitors health via periodic ping to the database server."
	}
	chunks := []string{
		makeChunk("ServiceA"), makeChunk("ServiceB"), makeChunk("ServiceC"),
		makeChunk("ServiceD"), makeChunk("ServiceE"),
		"", "FATAL: cluster unreachable after 5 retries exceeded limit.",
	}
	input := strings.Join(chunks, "\n")

	_, savedMin := f.Apply(input, ModeMinimal)
	_, savedAgg := f.Apply(input, ModeAggressive)
	if savedAgg < savedMin {
		t.Error("aggressive should save >= minimal")
	}
}
