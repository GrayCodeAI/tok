package clustering

import (
	"context"
	"testing"
)

func TestSemanticClusteringEngine(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewSemanticClusteringEngine(config)

	if len(engine.algorithms) == 0 {
		t.Error("Expected algorithms to be registered")
	}

	algo, ok := engine.algorithms["kmeans"]
	if !ok {
		t.Error("Expected kmeans algorithm to exist")
	}

	if algo.Name() != "kmeans" {
		t.Errorf("Expected kmeans, got %s", algo.Name())
	}
}

func TestKMeansClusterer(t *testing.T) {
	clusterer := &KMeansClusterer{}

	points := []Point{
		{ID: "1", Vector: []float64{1, 1}},
		{ID: "2", Vector: []float64{1.1, 1.1}},
		{ID: "3", Vector: []float64{10, 10}},
		{ID: "4", Vector: []float64{10.1, 10.1}},
	}

	clusters, err := clusterer.Cluster(context.Background(), points, ClusterParams{NumClusters: 2})
	if err != nil {
		t.Fatalf("Cluster failed: %v", err)
	}

	if len(clusters) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(clusters))
	}
}

func TestDBSCANClusterer(t *testing.T) {
	clusterer := &DBSCANClusterer{}

	points := []Point{
		{ID: "1", Vector: []float64{1, 1}},
		{ID: "2", Vector: []float64{1.1, 1.1}},
		{ID: "3", Vector: []float64{10, 10}},
	}

	clusters, err := clusterer.Cluster(context.Background(), points, ClusterParams{Epsilon: 2, MinClusterSize: 2})
	if err != nil {
		t.Fatalf("Cluster failed: %v", err)
	}

	if len(clusters) == 0 {
		t.Error("Expected at least 1 cluster")
	}
}

func TestHierarchicalClusterer(t *testing.T) {
	clusterer := &HierarchicalClusterer{}

	points := []Point{
		{ID: "1", Vector: []float64{1, 1}},
		{ID: "2", Vector: []float64{1.1, 1.1}},
		{ID: "3", Vector: []float64{10, 10}},
		{ID: "4", Vector: []float64{10.1, 10.1}},
		{ID: "5", Vector: []float64{20, 20}},
	}

	clusters, err := clusterer.Cluster(context.Background(), points, ClusterParams{NumClusters: 2})
	if err != nil {
		t.Fatalf("Cluster failed: %v", err)
	}

	if len(clusters) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(clusters))
	}
}

func TestEngineStats(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewSemanticClusteringEngine(config)

	documents := []string{
		"document one content",
		"document two content",
		"document three content",
		"document four content",
	}

	clusters, err := engine.Cluster(context.Background(), documents, ClusterParams{NumClusters: 2})
	if err != nil {
		t.Fatalf("Cluster failed: %v", err)
	}

	if len(clusters) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(clusters))
	}

	stats := engine.GetStats()

	if stats.TotalClusterings != 1 {
		t.Errorf("Expected 1 clustering, got %d", stats.TotalClusterings)
	}

	if stats.TotalDocuments != 4 {
		t.Errorf("Expected 4 documents, got %d", stats.TotalDocuments)
	}
}

func TestClusterQuality(t *testing.T) {
	engine := NewSemanticClusteringEngine(DefaultEngineConfig())

	documents := []string{
		"document one about machine learning",
		"document two about deep learning",
		"document three about python programming",
		"document four about javascript coding",
	}

	clusters, _ := engine.Cluster(context.Background(), documents, ClusterParams{NumClusters: 2})

	if len(clusters) != 2 {
		t.Fatalf("Expected 2 clusters")
	}

	if clusters[0].Quality.SilhouetteScore == 0 {
		t.Log("Silhouette score is 0 (may need more points)")
	}
}

func TestMinibatchKMeans(t *testing.T) {
	clusterer := &MiniBatchKMeans{}

	points := make([]Point, 100)
	for i := range points {
		points[i] = Point{
			ID:     string(rune(i)),
			Vector: []float64{float64(i % 10), float64(i % 5)},
		}
	}

	clusters, err := clusterer.Cluster(context.Background(), points, ClusterParams{NumClusters: 3})
	if err != nil {
		t.Fatalf("Cluster failed: %v", err)
	}

	if len(clusters) != 3 {
		t.Errorf("Expected 3 clusters, got %d", len(clusters))
	}
}
