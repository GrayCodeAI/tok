package stress

import (
	"fmt"
	"testing"
	"time"
)

func TestNewDistributedRunner(t *testing.T) {
	config := DistributedConfig{
		IsCoordinator: true,
	}

	runner := NewDistributedRunner(config)
	if runner == nil {
		t.Fatal("expected runner to be created")
	}

	if runner.config.IsCoordinator != true {
		t.Error("expected IsCoordinator to be true")
	}

	if runner.config.HeartbeatInterval != 5*time.Second {
		t.Errorf("expected default heartbeat 5s, got %v", runner.config.HeartbeatInterval)
	}
}

func TestDistributedRunnerRegisterNode(t *testing.T) {
	runner := NewDistributedRunner(DistributedConfig{})

	node := &Node{
		ID:      "node-1",
		Address: "localhost:8080",
	}

	err := runner.RegisterNode(node)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(runner.nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(runner.nodes))
	}
}

func TestDistributedRunnerRegisterDuplicateNode(t *testing.T) {
	runner := NewDistributedRunner(DistributedConfig{})

	node := &Node{
		ID:      "node-1",
		Address: "localhost:8080",
	}

	runner.RegisterNode(node)
	err := runner.RegisterNode(node)

	if err == nil {
		t.Error("expected error for duplicate node")
	}
}

func TestDistributedRunnerListNodes(t *testing.T) {
	runner := NewDistributedRunner(DistributedConfig{})

	runner.RegisterNode(&Node{ID: "node-1", Address: "localhost:8080"})
	runner.RegisterNode(&Node{ID: "node-2", Address: "localhost:8081"})

	nodes := runner.ListNodes()

	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestDistributedRunnerGetOnlineNodes(t *testing.T) {
	runner := NewDistributedRunner(DistributedConfig{})

	// Register nodes - they start as Online
	runner.RegisterNode(&Node{ID: "node-1"})
	runner.RegisterNode(&Node{ID: "node-2"})

	// Manually set one as offline
	runner.nodes["node-2"].Status = NodeStatusOffline

	nodes := runner.GetOnlineNodes()

	if len(nodes) != 1 {
		t.Errorf("expected 1 online node, got %d", len(nodes))
	}

	if nodes[0].ID != "node-1" {
		t.Errorf("expected 'node-1', got %s", nodes[0].ID)
	}
}

func TestDistributedRunnerUnregisterNode(t *testing.T) {
	runner := NewDistributedRunner(DistributedConfig{})

	runner.RegisterNode(&Node{ID: "node-1"})
	runner.UnregisterNode("node-1")

	if len(runner.nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(runner.nodes))
	}
}

func TestDistributedResultGenerateReport(t *testing.T) {
	result := &DistributedResult{
		TotalRequests: 1000,
		SuccessCount:  950,
		SuccessRate:   95.0,
		ThroughputRPS: 100.0,
		Duration:      10 * time.Second,
		NodeResults: []Result{
			{TotalRequests: 500, SuccessCount: 475},
			{TotalRequests: 500, SuccessCount: 475},
		},
	}

	report := result.GenerateReport()

	if report == "" {
		t.Error("expected non-empty report")
	}

	if !contains(report, "1000") {
		t.Error("expected total requests in report")
	}
}

func TestNewGeographicDistribution(t *testing.T) {
	gd := NewGeographicDistribution()
	if gd == nil {
		t.Fatal("expected geographic distribution to be created")
	}

	if gd.Regions == nil {
		t.Error("expected regions map to be initialized")
	}
}

func TestGeographicDistributionAddRegion(t *testing.T) {
	gd := NewGeographicDistribution()
	gd.AddRegion("us-west", 50.0, []string{"node-1", "node-2"})

	if len(gd.Regions) != 1 {
		t.Errorf("expected 1 region, got %d", len(gd.Regions))
	}

	region, ok := gd.Regions["us-west"]
	if !ok {
		t.Fatal("expected 'us-west' region")
	}

	if region.Percentage != 50.0 {
		t.Errorf("expected 50%%, got %.2f%%", region.Percentage)
	}
}

func TestGeographicDistributionDistributeScenario(t *testing.T) {
	gd := NewGeographicDistribution()
	gd.AddRegion("us-west", 50.0, []string{"node-1"})
	gd.AddRegion("us-east", 50.0, []string{"node-2"})

	scenario := &Scenario{
		Name: "test",
		Type: TypeLoad,
	}

	distributed := gd.DistributeScenario(scenario, 100)

	if len(distributed) != 2 {
		t.Errorf("expected 2 distributed scenarios, got %d", len(distributed))
	}
}

func TestNewLoadBalancer(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	if lb == nil {
		t.Fatal("expected load balancer to be created")
	}

	if lb.strategy != StrategyRoundRobin {
		t.Errorf("expected round robin strategy, got %s", lb.strategy)
	}
}

func TestLoadBalancerAddNode(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	lb.AddNode(&Node{ID: "node-1"})

	if len(lb.nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(lb.nodes))
	}
}

func TestLoadBalancerNextNodeRoundRobin(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	lb.AddNode(&Node{ID: "node-1"})
	lb.AddNode(&Node{ID: "node-2"})

	// First call should return node-1
	node1 := lb.NextNode()
	if node1 == nil || node1.ID != "node-1" {
		t.Errorf("expected node-1 first, got %v", node1)
	}

	// Second call should return node-2
	node2 := lb.NextNode()
	if node2 == nil || node2.ID != "node-2" {
		t.Errorf("expected node-2 second, got %v", node2)
	}

	// Third call should return node-1 again
	node3 := lb.NextNode()
	if node3 == nil || node3.ID != "node-1" {
		t.Errorf("expected node-1 third, got %v", node3)
	}
}

func TestLoadBalancerNextNodeEmpty(t *testing.T) {
	lb := NewLoadBalancer(StrategyRoundRobin)

	node := lb.NextNode()
	if node != nil {
		t.Error("expected nil for empty load balancer")
	}
}

func TestLoadBalancerNextNodeLeastLoad(t *testing.T) {
	lb := NewLoadBalancer(StrategyLeastLoad)
	lb.AddNode(&Node{ID: "busy", Load: 80.0})
	lb.AddNode(&Node{ID: "idle", Load: 20.0})

	node := lb.NextNode()
	if node == nil || node.ID != "idle" {
		t.Errorf("expected 'idle' node, got %v", node)
	}
}

func TestNewHeartbeatMonitor(t *testing.T) {
	runner := NewDistributedRunner(DistributedConfig{})
	monitor := NewHeartbeatMonitor(runner, 5*time.Second, 30*time.Second)

	if monitor == nil {
		t.Fatal("expected monitor to be created")
	}

	if monitor.interval != 5*time.Second {
		t.Errorf("expected 5s interval, got %v", monitor.interval)
	}
}

func TestNewSyncBarrier(t *testing.T) {
	barrier := NewSyncBarrier(3)
	if barrier == nil {
		t.Fatal("expected barrier to be created")
	}

	if barrier.expected != 3 {
		t.Errorf("expected 3 participants, got %d", barrier.expected)
	}
}

func BenchmarkLoadBalancerNextNode(b *testing.B) {
	lb := NewLoadBalancer(StrategyRoundRobin)
	for i := 0; i < 10; i++ {
		lb.AddNode(&Node{ID: fmt.Sprintf("node-%d", i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.NextNode()
	}
}
