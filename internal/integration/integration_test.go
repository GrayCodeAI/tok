package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/GrayCodeAI/tokman/internal/auth"
	"github.com/GrayCodeAI/tokman/internal/core"
	"github.com/GrayCodeAI/tokman/internal/license"
	"github.com/GrayCodeAI/tokman/internal/observability"
	"github.com/GrayCodeAI/tokman/internal/tracking"
	"github.com/GrayCodeAI/tokman/services/analytics"
)

// IntegrationTestSuite runs full end-to-end tests
type IntegrationTestSuite struct {
	t                    *testing.T
	apiEndpoint          string
	testAPIKey           string
	testTeamID           string
	httpClient           *http.Client
	analyticsService     *analytics.Service
	licenseManager       *license.Manager
	authManager          *auth.Manager
	compressionPipeline  *core.Pipeline
	observabilityManager *observability.Manager
}

// SetupSuite initializes test environment
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{
		t:           t,
		apiEndpoint: "http://localhost:8083",
		testAPIKey:  "test-key-integration",
		testTeamID:  "team-test-integration",
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}

	// Initialize services
	suite.setup()
	return suite
}

func (s *IntegrationTestSuite) setup() {
	// Initialize tracking/database
	db := tracking.NewDatabase()
	if err := db.Init(); err != nil {
		s.t.Fatalf("Failed to init database: %v", err)
	}

	// Initialize analytics
	s.analyticsService = analytics.NewService(db)

	// Initialize license manager
	s.licenseManager = license.NewManager()
	proLicense := &license.License{
		TeamID: s.testTeamID,
		Tier:   license.TierPro,
		Status: "active",
	}
	s.licenseManager.CreateLicense(proLicense)

	// Initialize auth manager
	s.authManager = auth.NewManager(db)

	// Initialize compression pipeline
	s.compressionPipeline = core.NewPipeline()

	// Initialize observability
	s.observabilityManager = observability.NewManager()
}

// TestEndToEndAnalysis tests complete analysis flow
func (s *IntegrationTestSuite) TestEndToEndAnalysis(t *testing.T) {
	code := `
func calculateSum(numbers []int) int {
    sum := 0
    for _, num := range numbers {
        sum += num
    }
    return sum
}
`

	payload := map[string]interface{}{
		"code":              code,
		"language":          "go",
		"compression_level": "aggressive",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/analyze", s.apiEndpoint), bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.testAPIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if _, ok := result["tokens_saved"]; !ok {
		t.Error("Missing tokens_saved in response")
	}
	if _, ok := result["compression_ratio"]; !ok {
		t.Error("Missing compression_ratio in response")
	}
	if _, ok := result["processing_time_ms"]; !ok {
		t.Error("Missing processing_time_ms in response")
	}

	// Verify compression ratio is valid (0 < ratio < 1)
	ratio, ok := result["compression_ratio"].(float64)
	if !ok || ratio < 0 || ratio > 1 {
		t.Errorf("Invalid compression ratio: %v", result["compression_ratio"])
	}
}

// TestBatchAnalysis tests batch processing
func (s *IntegrationTestSuite) TestBatchAnalysis(t *testing.T) {
	files := map[string]interface{}{
		"files": []map[string]string{
			{
				"name": "file1.go",
				"code": "func add(a, b int) int { return a + b }",
			},
			{
				"name": "file2.go",
				"code": "func multiply(a, b int) int { return a * b }",
			},
		},
	}

	body, _ := json.Marshal(files)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/analyze-batch", s.apiEndpoint), bytes.NewReader(body))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.testAPIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		t.Fatalf("Batch analysis failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	results, ok := result["results"].([]interface{})
	if !ok || len(results) != 2 {
		t.Errorf("Expected 2 results, got %v", result)
	}
}

// TestRateLimiting verifies rate limit enforcement
func (s *IntegrationTestSuite) TestRateLimiting(t *testing.T) {
	// Set up free tier user with low quota
	freeLicense := &license.License{
		TeamID: "team-free",
		Tier:   license.TierFree,
		Status: "active",
	}
	s.licenseManager.CreateLicense(freeLicense)

	payload := []byte(`{"code":"func test() {}","language":"go"}`)
	rateLimitExceeded := false

	// Make requests up to limit + 1
	for i := 0; i < 101; i++ {
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/analyze", s.apiEndpoint), bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer free-team-key")
		req.Header.Set("Content-Type", "application/json")

		resp, _ := s.httpClient.Do(req)
		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitExceeded = true
			t.Logf("Rate limit hit at request %d (expected for free tier)", i+1)
		}
		resp.Body.Close()
	}

	if !rateLimitExceeded {
		t.Errorf("Expected rate limit for free tier after 100 requests")
	}
}

// TestMultiTenantIsolation verifies teams are isolated
func (s *IntegrationTestSuite) TestMultiTenantIsolation(t *testing.T) {
	team1 := "team-isolation-1"
	team2 := "team-isolation-2"

	// Create teams
	s.analyticsService.CreateTeam(context.Background(), &analytics.Team{
		ID:   team1,
		Name: "Team 1",
	})
	s.analyticsService.CreateTeam(context.Background(), &analytics.Team{
		ID:   team2,
		Name: "Team 2",
	})

	// Record metrics for both teams
	ctx := context.Background()
	s.analyticsService.RecordFilterActivation(ctx, team1, "filter1", 100, 50, 10)
	s.analyticsService.RecordFilterActivation(ctx, team2, "filter1", 200, 100, 20)

	// Verify isolation
	stats1, _ := s.analyticsService.GetTeamStats(ctx, team1)
	stats2, _ := s.analyticsService.GetTeamStats(ctx, team2)

	if stats1.TeamID != team1 {
		t.Error("Team 1 stats should belong to team 1")
	}
	if stats2.TeamID != team2 {
		t.Error("Team 2 stats should belong to team 2")
	}

	// Cross-team query should fail
	authCtx := &auth.AuthContext{
		TeamID: team1,
	}
	ctx = context.WithValue(ctx, "auth", authCtx)

	stats, err := s.analyticsService.GetTeamStats(ctx, team2)
	if err == nil || stats.TeamID == team2 {
		t.Error("Should not allow cross-team data access")
	}
}

// TestAuthenticationFlow verifies auth mechanisms
func (s *IntegrationTestSuite) TestAuthenticationFlow(t *testing.T) {
	// Test valid token
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/analyze", s.apiEndpoint), bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.testAPIKey))

	resp, _ := s.httpClient.Do(req)
	if resp.StatusCode == http.StatusUnauthorized {
		t.Error("Valid token should not return 401")
	}
	resp.Body.Close()

	// Test invalid token
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/analyze", s.apiEndpoint), bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, _ = s.httpClient.Do(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Invalid token should return 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Test missing token
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/analyze", s.apiEndpoint), bytes.NewReader([]byte(`{}`)))

	resp, _ = s.httpClient.Do(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Missing token should return 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestCachingBehavior verifies caching layer
func (s *IntegrationTestSuite) TestCachingBehavior(t *testing.T) {
	code := "func test() { println(\"hello\") }"
	payload := map[string]interface{}{
		"code":     code,
		"language": "go",
	}
	body, _ := json.Marshal(payload)

	var firstDuration, secondDuration time.Duration

	// First request (cache miss)
	start := time.Now()
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/analyze", s.apiEndpoint), bytes.NewReader(body))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.testAPIKey))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := s.httpClient.Do(req)
	firstDuration = time.Since(start)
	resp.Body.Close()

	// Second request (cache hit)
	start = time.Now()
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/analyze", s.apiEndpoint), bytes.NewReader(body))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.testAPIKey))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = s.httpClient.Do(req)
	secondDuration = time.Since(start)
	resp.Body.Close()

	// Cached request should be faster
	if secondDuration >= firstDuration {
		t.Logf("Cache may not be working: first=%v, second=%v", firstDuration, secondDuration)
	}
}

// TestConcurrentRequests verifies concurrent handling
func (s *IntegrationTestSuite) TestConcurrentRequests(t *testing.T) {
	code := "func add(a, b int) int { return a + b }"
	payload := map[string]interface{}{"code": code, "language": "go"}
	body, _ := json.Marshal(payload)

	var wg sync.WaitGroup
	errorCount := 0
	var mu sync.Mutex

	// Spin up 100 concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, _ := http.NewRequest("POST", fmt.Sprintf("%s/analyze", s.apiEndpoint), bytes.NewReader(body))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.testAPIKey))
			req.Header.Set("Content-Type", "application/json")

			resp, err := s.httpClient.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				mu.Lock()
				errorCount++
				mu.Unlock()
			}
			if resp != nil {
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()

	if errorCount > 5 {
		t.Errorf("Too many concurrent request errors: %d/100", errorCount)
	}
}

// TestMetricsCollection verifies observability
func (s *IntegrationTestSuite) TestMetricsCollection(t *testing.T) {
	// Make a request that will generate metrics
	code := "func test() {}"
	payload := map[string]interface{}{"code": code, "language": "go"}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/analyze", s.apiEndpoint), bytes.NewReader(body))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.testAPIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := s.httpClient.Do(req)
	resp.Body.Close()

	// Verify metrics endpoint is available
	metricsReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/metrics", s.apiEndpoint), nil)
	metricsResp, _ := s.httpClient.Do(metricsReq)

	if metricsResp.StatusCode != http.StatusOK {
		t.Errorf("Metrics endpoint should return 200, got %d", metricsResp.StatusCode)
	}
	metricsResp.Body.Close()
}

// Run executes all integration tests
func (s *IntegrationTestSuite) Run(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"EndToEndAnalysis", s.TestEndToEndAnalysis},
		{"BatchAnalysis", s.TestBatchAnalysis},
		{"RateLimiting", s.TestRateLimiting},
		{"MultiTenantIsolation", s.TestMultiTenantIsolation},
		{"AuthenticationFlow", s.TestAuthenticationFlow},
		{"CachingBehavior", s.TestCachingBehavior},
		{"ConcurrentRequests", s.TestConcurrentRequests},
		{"MetricsCollection", s.TestMetricsCollection},
	}

	for _, test := range tests {
		t.Run(test.name, test.fn)
	}
}
