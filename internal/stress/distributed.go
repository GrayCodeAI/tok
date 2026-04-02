// Package stress provides distributed stress testing capabilities
package stress

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DistributedRunner manages distributed stress testing
type DistributedRunner struct {
	config      DistributedConfig
	nodes       map[string]*Node
	coordinator *Coordinator
	mu          sync.RWMutex
}

// DistributedConfig holds distributed test configuration
type DistributedConfig struct {
	CoordinatorAddress string
	NodeID             string
	IsCoordinator      bool
	HeartbeatInterval  time.Duration
	SyncTimeout        time.Duration
}

// Node represents a stress testing node
type Node struct {
	ID       string
	Address  string
	Status   NodeStatus
	Capacity int
	Load     float64
	LastPing time.Time
	client   NodeClient
}

// NodeStatus represents node status
type NodeStatus string

const (
	NodeStatusOnline  NodeStatus = "online"
	NodeStatusOffline NodeStatus = "offline"
	NodeStatusBusy    NodeStatus = "busy"
)

// NodeClient interface for node communication
type NodeClient interface {
	Connect(address string) error
	Execute(scenario *Scenario) (*Result, error)
	HealthCheck() error
	Close() error
}

// Coordinator manages distributed test execution
type Coordinator struct {
	nodes     map[string]*Node
	scenarios map[string]*DistributedScenario
	results   map[string][]Result
	mu        sync.RWMutex
	stopCh    chan struct{}
}

// DistributedScenario represents a scenario for distributed execution
type DistributedScenario struct {
	*Scenario
	NodeAssignments map[string][]string // scenario -> nodes
	TotalRPS        int
	GeographicDist  map[string]float64 // region -> percentage
}

// NewDistributedRunner creates a new distributed runner
func NewDistributedRunner(config DistributedConfig) *DistributedRunner {
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = 5 * time.Second
	}
	if config.SyncTimeout == 0 {
		config.SyncTimeout = 30 * time.Second
	}

	return &DistributedRunner{
		config: config,
		nodes:  make(map[string]*Node),
	}
}

// RegisterNode registers a stress testing node
func (dr *DistributedRunner) RegisterNode(node *Node) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	if _, exists := dr.nodes[node.ID]; exists {
		return fmt.Errorf("node %s already registered", node.ID)
	}

	// Connect to node
	if node.client != nil {
		if err := node.client.Connect(node.Address); err != nil {
			return fmt.Errorf("failed to connect to node %s: %w", node.ID, err)
		}
	}

	node.Status = NodeStatusOnline
	node.LastPing = time.Now()
	dr.nodes[node.ID] = node

	return nil
}

// UnregisterNode removes a node
func (dr *DistributedRunner) UnregisterNode(nodeID string) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	if node, exists := dr.nodes[nodeID]; exists && node.client != nil {
		node.client.Close()
	}

	delete(dr.nodes, nodeID)
}

// ListNodes returns all registered nodes
func (dr *DistributedRunner) ListNodes() []*Node {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	nodes := make([]*Node, 0, len(dr.nodes))
	for _, node := range dr.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// GetOnlineNodes returns only online nodes
func (dr *DistributedRunner) GetOnlineNodes() []*Node {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	nodes := make([]*Node, 0)
	for _, node := range dr.nodes {
		if node.Status == NodeStatusOnline {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// RunDistributed executes a stress test across all nodes
func (dr *DistributedRunner) RunDistributed(ctx context.Context, scenario *Scenario) (*DistributedResult, error) {
	nodes := dr.GetOnlineNodes()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no online nodes available")
	}

	// Distribute load across nodes
	config := DefaultConfig()
	config.TargetRPS = config.TargetRPS / len(nodes) // Divide RPS among nodes

	results := make([]Result, 0, len(nodes))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()

			// Create runner for this node
			runner := NewRunner(config)
			runner.RegisterScenario(scenario)

			// Execute
			result, err := runner.Run(ctx, scenario.Name)
			if err != nil {
				return
			}

			mu.Lock()
			results = append(results, *result)
			mu.Unlock()
		}(node)
	}

	wg.Wait()

	// Aggregate results
	return dr.aggregateResults(results), nil
}

func (dr *DistributedRunner) aggregateResults(results []Result) *DistributedResult {
	drResult := &DistributedResult{
		NodeResults: results,
		StartTime:   time.Now(),
	}

	if len(results) == 0 {
		return drResult
	}

	// Aggregate metrics
	var totalRequests, successCount, errorCount, timeoutCount int64
	var totalLatency time.Duration

	for _, r := range results {
		totalRequests += r.TotalRequests
		successCount += r.SuccessCount
		errorCount += r.ErrorCount
		timeoutCount += r.TimeoutCount
		totalLatency += r.AvgLatency
	}

	drResult.TotalRequests = totalRequests
	drResult.SuccessCount = successCount
	drResult.ErrorCount = errorCount
	drResult.TimeoutCount = timeoutCount

	if totalRequests > 0 {
		drResult.SuccessRate = float64(successCount) / float64(totalRequests) * 100
		drResult.ThroughputRPS = float64(totalRequests) / results[0].Duration.Seconds()
	}

	if len(results) > 0 {
		drResult.AvgLatency = totalLatency / time.Duration(len(results))
	}

	drResult.EndTime = time.Now()
	drResult.Duration = drResult.EndTime.Sub(drResult.StartTime)

	return drResult
}

// DistributedResult holds aggregated results from distributed test
type DistributedResult struct {
	NodeResults   []Result
	TotalRequests int64
	SuccessCount  int64
	ErrorCount    int64
	TimeoutCount  int64
	SuccessRate   float64
	ThroughputRPS float64
	AvgLatency    time.Duration
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
}

// GenerateReport generates a distributed test report
func (dr *DistributedResult) GenerateReport() string {
	report := "Distributed Stress Test Report\n"
	report += "===============================\n\n"
	report += fmt.Sprintf("Duration: %v\n", dr.Duration)
	report += fmt.Sprintf("Total Requests: %d\n", dr.TotalRequests)
	report += fmt.Sprintf("Success Rate: %.2f%%\n", dr.SuccessRate)
	report += fmt.Sprintf("Throughput: %.2f RPS\n", dr.ThroughputRPS)
	report += fmt.Sprintf("Avg Latency: %v\n\n", dr.AvgLatency)

	report += "Node Results:\n"
	for i, node := range dr.NodeResults {
		successRate := 0.0
		if node.TotalRequests > 0 {
			successRate = float64(node.SuccessCount) / float64(node.TotalRequests) * 100
		}
		report += fmt.Sprintf("  Node %d: %d requests, %.2f%% success\n",
			i+1, node.TotalRequests, successRate)
	}

	return report
}

// GeographicDistribution defines load distribution by region
type GeographicDistribution struct {
	Regions map[string]RegionConfig
}

// RegionConfig holds region-specific configuration
type RegionConfig struct {
	Name       string
	Percentage float64
	Nodes      []string
}

// NewGeographicDistribution creates geographic distribution
func NewGeographicDistribution() *GeographicDistribution {
	return &GeographicDistribution{
		Regions: make(map[string]RegionConfig),
	}
}

// AddRegion adds a region
func (gd *GeographicDistribution) AddRegion(name string, percentage float64, nodes []string) {
	gd.Regions[name] = RegionConfig{
		Name:       name,
		Percentage: percentage,
		Nodes:      nodes,
	}
}

// DistributeScenario distributes scenario across regions
func (gd *GeographicDistribution) DistributeScenario(scenario *Scenario, totalRPS int) map[string]*Scenario {
	distributed := make(map[string]*Scenario)

	for region, config := range gd.Regions {
		regionRPS := int(float64(totalRPS) * config.Percentage / 100)

		// Clone scenario with adjusted RPS
		regionScenario := &Scenario{
			Name:        fmt.Sprintf("%s-%s", scenario.Name, region),
			Type:        scenario.Type,
			Description: scenario.Description,
			Fn:          scenario.Fn,
			Weight:      scenario.Weight,
		}

		// Store region-specific RPS in metadata
		// In real implementation, would adjust runner config
		_ = regionRPS

		distributed[region] = regionScenario
	}

	return distributed
}

// LoadBalancer balances load across nodes
type LoadBalancer struct {
	strategy LoadBalanceStrategy
	nodes    []*Node
	counter  int
	mu       sync.Mutex
}

// LoadBalanceStrategy defines load balancing strategy
type LoadBalanceStrategy string

const (
	StrategyRoundRobin LoadBalanceStrategy = "round_robin"
	StrategyRandom     LoadBalanceStrategy = "random"
	StrategyLeastLoad  LoadBalanceStrategy = "least_load"
)

// NewLoadBalancer creates a load balancer
func NewLoadBalancer(strategy LoadBalanceStrategy) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
		nodes:    make([]*Node, 0),
	}
}

// AddNode adds a node
func (lb *LoadBalancer) AddNode(node *Node) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.nodes = append(lb.nodes, node)
}

// NextNode returns the next node based on strategy
func (lb *LoadBalancer) NextNode() *Node {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.nodes) == 0 {
		return nil
	}

	switch lb.strategy {
	case StrategyRoundRobin:
		node := lb.nodes[lb.counter%len(lb.nodes)]
		lb.counter++
		return node

	case StrategyRandom:
		// In real implementation, use crypto/rand
		return lb.nodes[lb.counter%len(lb.nodes)]

	case StrategyLeastLoad:
		// Find node with lowest load
		var bestNode *Node
		bestLoad := float64(999999)
		for _, node := range lb.nodes {
			if node.Load < bestLoad {
				bestLoad = node.Load
				bestNode = node
			}
		}
		return bestNode

	default:
		return lb.nodes[0]
	}
}

// HeartbeatMonitor monitors node health
type HeartbeatMonitor struct {
	runner   *DistributedRunner
	interval time.Duration
	timeout  time.Duration
	stopCh   chan struct{}
}

// NewHeartbeatMonitor creates a heartbeat monitor
func NewHeartbeatMonitor(runner *DistributedRunner, interval, timeout time.Duration) *HeartbeatMonitor {
	return &HeartbeatMonitor{
		runner:   runner,
		interval: interval,
		timeout:  timeout,
		stopCh:   make(chan struct{}),
	}
}

// Start starts monitoring
func (hm *HeartbeatMonitor) Start() {
	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.stopCh:
			return
		case <-ticker.C:
			hm.checkNodes()
		}
	}
}

// Stop stops monitoring
func (hm *HeartbeatMonitor) Stop() {
	close(hm.stopCh)
}

func (hm *HeartbeatMonitor) checkNodes() {
	nodes := hm.runner.ListNodes()
	cutoff := time.Now().Add(-hm.timeout)

	for _, node := range nodes {
		if node.LastPing.Before(cutoff) {
			// Node is offline
			node.Status = NodeStatusOffline
		}
	}
}

// SyncBarrier synchronizes distributed test phases
type SyncBarrier struct {
	expected int
	count    int
	ch       chan struct{}
	mu       sync.Mutex
}

// NewSyncBarrier creates a sync barrier
func NewSyncBarrier(expected int) *SyncBarrier {
	return &SyncBarrier{
		expected: expected,
		ch:       make(chan struct{}),
	}
}

// Wait waits for all participants
func (sb *SyncBarrier) Wait() {
	sb.mu.Lock()
	sb.count++
	current := sb.count
	sb.mu.Unlock()

	if current >= sb.expected {
		close(sb.ch)
	}

	<-sb.ch
}

// DistributedTestConfig holds test configuration
type DistributedTestConfig struct {
	Name            string
	Scenario        *Scenario
	Duration        time.Duration
	TotalRPS        int
	GeographicDist  map[string]float64
	NodeSelection   LoadBalanceStrategy
	FailOnNodeError bool
}

// RunDistributedTest runs a complete distributed test
func RunDistributedTest(config DistributedTestConfig, nodes []*Node) (*DistributedResult, error) {
	runner := NewDistributedRunner(DistributedConfig{
		IsCoordinator: true,
	})

	// Register nodes
	for _, node := range nodes {
		if err := runner.RegisterNode(node); err != nil {
			return nil, err
		}
	}

	// Set up load balancer
	lb := NewLoadBalancer(config.NodeSelection)
	for _, node := range nodes {
		lb.AddNode(node)
	}

	// Distribute scenario geographically if specified
	if len(config.GeographicDist) > 0 {
		gd := NewGeographicDistribution()
		for region, percentage := range config.GeographicDist {
			// Assign nodes to regions
			regionNodes := make([]string, 0)
			for _, node := range nodes {
				regionNodes = append(regionNodes, node.ID)
			}
			gd.AddRegion(region, percentage, regionNodes)
		}

		// Get distributed scenarios
		scenarios := gd.DistributeScenario(config.Scenario, config.TotalRPS)
		_ = scenarios
	}

	// Run test
	ctx := context.Background()
	return runner.RunDistributed(ctx, config.Scenario)
}
