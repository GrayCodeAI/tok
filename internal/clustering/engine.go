package clustering

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
)

type SemanticClusteringEngine struct {
	mu         sync.RWMutex
	config     EngineConfig
	algorithms map[string]Clusterer
	history    *ClusterHistory
	stats      EngineStats
}

type EngineConfig struct {
	DefaultAlgorithm string
	EmbeddingDim     int
	MinClusterSize   int
	MaxClusters      int
	DistanceMetric   string
}

type EngineStats struct {
	TotalClusterings int64
	TotalDocuments   int64
	AvgClustersPer   float64
	AlgorithmUsage   map[string]int64
}

type Clusterer interface {
	Cluster(ctx context.Context, points []Point, params ClusterParams) ([]Cluster, error)
	Name() string
}

type ClusterParams struct {
	NumClusters    int
	MinClusterSize int
	Epsilon        float64
	Iterations     int
}

type Point struct {
	ID       string
	Vector   []float64
	Metadata map[string]interface{}
}

type Cluster struct {
	ID       string
	Points   []Point
	Centroid []float64
	Label    string
	Summary  string
	Quality  ClusterQuality
}

type ClusterQuality struct {
	SilhouetteScore  float64
	DaviesBouldin    float64
	CalinskiHarabasz float64
}

func NewSemanticClusteringEngine(config EngineConfig) *SemanticClusteringEngine {
	e := &SemanticClusteringEngine{
		config:     config,
		algorithms: make(map[string]Clusterer),
		history:    NewClusterHistory(100),
		stats: EngineStats{
			AlgorithmUsage: make(map[string]int64),
		},
	}

	e.registerAlgorithms()

	return e
}

func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		DefaultAlgorithm: "kmeans",
		EmbeddingDim:     768,
		MinClusterSize:   2,
		MaxClusters:      20,
		DistanceMetric:   "cosine",
	}
}

func (e *SemanticClusteringEngine) registerAlgorithms() {
	e.RegisterClusterer(&KMeansClusterer{})
	e.RegisterClusterer(&DBSCANClusterer{})
	e.RegisterClusterer(&HierarchicalClusterer{})
	e.RegisterClusterer(&MiniBatchKMeans{})
}

func (e *SemanticClusteringEngine) RegisterClusterer(c Clusterer) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.algorithms[c.Name()] = c
}

func (e *SemanticClusteringEngine) Cluster(ctx context.Context, documents []string, params ClusterParams) ([]Cluster, error) {
	points := e.generateEmbeddings(documents)

	algoName := e.config.DefaultAlgorithm
	if params.NumClusters > 0 {
		algoName = "kmeans"
	}

	e.mu.RLock()
	algo, ok := e.algorithms[algoName]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("algorithm not found: %s", algoName)
	}

	clusters, err := algo.Cluster(ctx, points, params)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.stats.TotalClusterings++
	e.stats.TotalDocuments += int64(len(documents))
	e.stats.AlgorithmUsage[algoName]++
	e.mu.Unlock()

	for i := range clusters {
		clusters[i].Label = fmt.Sprintf("Cluster %d", i)
		clusters[i].Summary = e.summarizeCluster(clusters[i])
		clusters[i].Quality = e.calculateQuality(points, clusters)
	}

	return clusters, nil
}

func (e *SemanticClusteringEngine) generateEmbeddings(documents []string) []Point {
	points := make([]Point, len(documents))

	for i, doc := range documents {
		vector := generateMockEmbedding(doc, e.config.EmbeddingDim)
		points[i] = Point{
			ID:       fmt.Sprintf("doc_%d", i),
			Vector:   vector,
			Metadata: map[string]interface{}{"text": doc},
		}
	}

	return points
}

func generateMockEmbedding(text string, dim int) []float64 {
	vector := make([]float64, dim)
	seed := hashString(text)
	r := seededRandom(seed)

	for i := 0; i < dim; i++ {
		vector[i] = r.Float64()*2 - 1
	}

	_norm := 0.0
	for _, v := range vector {
		_norm += v * v
	}
	_norm = math.Sqrt(_norm)
	if _norm > 0 {
		for i := range vector {
			vector[i] /= _norm
		}
	}

	return vector
}

func hashString(s string) uint64 {
	var h uint64
	for i, c := range s {
		h = h*31 + uint64(c)*uint64(i+1)
	}
	return h
}

type deterministic struct {
	seed uint64
	pos  int
}

func seededRandom(seed uint64) *deterministic {
	return &deterministic{seed: seed}
}

func (r *deterministic) Float64() float64 {
	r.seed = r.seed*1103515245 + 12345
	r.pos++
	return float64(r.seed%1000) / 1000.0
}

func (e *SemanticClusteringEngine) summarizeCluster(cluster Cluster) string {
	if len(cluster.Points) == 0 {
		return "Empty cluster"
	}

	words := make(map[string]int)
	for _, p := range cluster.Points {
		if text, ok := p.Metadata["text"].(string); ok {
			for _, word := range extractWords(text) {
				words[word]++
			}
		}
	}

	type wordCount struct {
		word  string
		count int
	}

	var counts []wordCount
	for w, c := range words {
		counts = append(counts, wordCount{w, c})
	}

	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	topWords := counts[:min(5, len(counts))]
	result := "Contains: "
	for i, wc := range topWords {
		if i > 0 {
			result += ", "
		}
		result += wc.word
	}

	return result
}

func extractWords(text string) []string {
	words := make([]string, 0)
	current := ""
	for _, c := range text {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
			current += string(c)
		} else if len(current) > 2 {
			words = append(words, current)
			current = ""
		}
	}
	if len(current) > 2 {
		words = append(words, current)
	}
	return words
}

func (e *SemanticClusteringEngine) calculateQuality(points []Point, clusters []Cluster) ClusterQuality {
	quality := ClusterQuality{}

	if len(clusters) < 2 || len(points) < 2 {
		return quality
	}

	quality.SilhouetteScore = calculateSilhouette(points, clusters)
	quality.DaviesBouldin = calculateDaviesBouldin(points, clusters)
	quality.CalinskiHarabasz = calculateCalinskiHarabasz(points, clusters)

	return quality
}

func calculateSilhouette(points []Point, clusters []Cluster) float64 {
	var total float64
	count := 0

	for _, cluster := range clusters {
		for _, p := range cluster.Points {
			a := avgDistanceToCluster(p, cluster)
			b := math.MaxFloat64
			for _, other := range clusters {
				if other.ID != cluster.ID {
					d := avgDistanceToCluster(p, other)
					if d < b {
						b = d
					}
				}
			}
			if b > 0 {
				s := (b - a) / math.Max(a, b)
				total += s
				count++
			}
		}
	}

	if count > 0 {
		return total / float64(count)
	}
	return 0
}

func avgDistanceToCluster(p Point, cluster Cluster) float64 {
	if len(cluster.Points) == 0 {
		return 0
	}

	var total float64
	for _, other := range cluster.Points {
		total += cosineDistance(p.Vector, other.Vector)
	}

	return total / float64(len(cluster.Points))
}

func cosineDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return 1
	}

	dot := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 1
	}

	return 1 - dot/(math.Sqrt(normA)*math.Sqrt(normB))
}

func calculateDaviesBouldin(points []Point, clusters []Cluster) float64 {
	if len(clusters) < 2 {
		return 0
	}

	var total float64
	count := 0

	for i, cluster := range clusters {
		maxR := 0.0
		for j, other := range clusters {
			if i == j {
				continue
			}

			si := clusterDispersion(cluster)
			sj := clusterDispersion(other)
			dij := clusterDistance(cluster, other)

			if dij > 0 {
				r := (si + sj) / dij
				if r > maxR {
					maxR = r
				}
			}
		}
		total += maxR
		count++
	}

	if count > 0 {
		return total / float64(count)
	}
	return 0
}

func clusterDispersion(cluster Cluster) float64 {
	if len(cluster.Points) == 0 || len(cluster.Centroid) == 0 {
		return 0
	}

	var total float64
	for _, p := range cluster.Points {
		total += euclideanDistance(p.Vector, cluster.Centroid)
	}

	return total / float64(len(cluster.Points))
}

func clusterDistance(a, b Cluster) float64 {
	return euclideanDistance(a.Centroid, b.Centroid)
}

func euclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}

	var sum float64
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}

	return math.Sqrt(sum)
}

func calculateCalinskiHarabasz(points []Point, clusters []Cluster) float64 {
	n := len(points)
	k := len(clusters)

	if n <= k || k < 2 {
		return 0
	}

	var between float64
	for _, cluster := range clusters {
		center := make([]float64, len(cluster.Centroid))
		for i := range center {
			center[i] = cluster.Centroid[i]
		}

		for _, p := range cluster.Points {
			d := euclideanDistance(p.Vector, center)
			between += d * d
		}
	}

	var within float64
	for _, cluster := range clusters {
		for _, p := range cluster.Points {
			d := euclideanDistance(p.Vector, cluster.Centroid)
			within += d * d
		}
	}

	B := between / float64(k-1)
	W := within / float64(n-k)

	if W > 0 {
		return B / W
	}
	return 0
}

func (e *SemanticClusteringEngine) GetStats() EngineStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

type ClusterHistory struct {
	mu      sync.RWMutex
	entries []ClusteringEntry
	maxSize int
}

type ClusteringEntry struct {
	Timestamp    string
	NumDocuments int
	NumClusters  int
	Algorithm    string
}

func NewClusterHistory(maxSize int) *ClusterHistory {
	return &ClusterHistory{
		entries: make([]ClusteringEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (h *ClusterHistory) Add(entry ClusteringEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, entry)
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[1:]
	}
}

type KMeansClusterer struct{}

func (c *KMeansClusterer) Name() string { return "kmeans" }

func (c *KMeansClusterer) Cluster(ctx context.Context, points []Point, params ClusterParams) ([]Cluster, error) {
	k := params.NumClusters
	if k <= 0 {
		k = 5
	}

	if len(points) < k {
		k = len(points)
	}

	centroids := make([][]float64, k)
	for i := 0; i < k; i++ {
		centroids[i] = make([]float64, len(points[0].Vector))
		copy(centroids[i], points[i].Vector)
	}

	iterations := params.Iterations
	if iterations <= 0 {
		iterations = 100
	}

	for iter := 0; iter < iterations; iter++ {
		clusters := make([][]Point, k)

		for _, p := range points {
			minDist := math.MaxFloat64
			closest := 0

			for i, centroid := range centroids {
				d := euclideanDistance(p.Vector, centroid)
				if d < minDist {
					minDist = d
					closest = i
				}
			}

			clusters[closest] = append(clusters[closest], p)
		}

		for i := range centroids {
			if len(clusters[i]) > 0 {
				for dim := range centroids[i] {
					var sum float64
					for _, p := range clusters[i] {
						sum += p.Vector[dim]
					}
					centroids[i][dim] = sum / float64(len(clusters[i]))
				}
			}
		}
	}

	result := make([]Cluster, k)
	for i := 0; i < k; i++ {
		result[i].ID = fmt.Sprintf("cluster_%d", i)
	}

	for _, p := range points {
		minDist := math.MaxFloat64
		closest := 0

		for i, centroid := range centroids {
			d := euclideanDistance(p.Vector, centroid)
			if d < minDist {
				minDist = d
				closest = i
			}
		}

		result[closest].Points = append(result[closest].Points, p)
		result[closest].Centroid = centroids[closest]
	}

	return result, nil
}

type DBSCANClusterer struct{}

func (c *DBSCANClusterer) Name() string { return "dbscan" }

func (c *DBSCANClusterer) Cluster(ctx context.Context, points []Point, params ClusterParams) ([]Cluster, error) {
	epsilon := params.Epsilon
	if epsilon <= 0 {
		epsilon = 0.5
	}

	minPts := params.MinClusterSize
	if minPts <= 0 {
		minPts = 2
	}

	visited := make(map[string]bool)
	clusters := []Cluster{}
	clusterID := 0

	for _, p := range points {
		if visited[p.ID] {
			continue
		}

		neighbors := regionQuery(points, p, epsilon)

		if len(neighbors) < minPts {
			continue
		}

		cluster := Cluster{
			ID: fmt.Sprintf("cluster_%d", clusterID),
		}
		clusterID++

		expandCluster(points, p, neighbors, &cluster, visited, epsilon, minPts)
		cluster.Centroid = calculateCentroid(cluster.Points)

		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func regionQuery(points []Point, p Point, epsilon float64) []Point {
	neighbors := []Point{}

	for _, other := range points {
		if euclideanDistance(p.Vector, other.Vector) <= epsilon {
			neighbors = append(neighbors, other)
		}
	}

	return neighbors
}

func expandCluster(points []Point, p Point, neighbors []Point, cluster *Cluster, visited map[string]bool, epsilon float64, minPts int) {
	visited[p.ID] = true
	cluster.Points = append(cluster.Points, p)

	queue := make([]Point, len(neighbors))
	copy(queue, neighbors)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.ID] {
			continue
		}

		visited[current.ID] = true
		cluster.Points = append(cluster.Points, current)

		currentNeighbors := regionQuery(points, current, epsilon)
		if len(currentNeighbors) >= minPts {
			queue = append(queue, currentNeighbors...)
		}
	}
}

func calculateCentroid(points []Point) []float64 {
	if len(points) == 0 {
		return nil
	}

	dim := len(points[0].Vector)
	centroid := make([]float64, dim)

	for _, p := range points {
		for i := range dim {
			centroid[i] += p.Vector[i]
		}
	}

	for i := range dim {
		centroid[i] /= float64(len(points))
	}

	return centroid
}

type HierarchicalClusterer struct{}

func (c *HierarchicalClusterer) Name() string { return "hierarchical" }

func (c *HierarchicalClusterer) Cluster(ctx context.Context, points []Point, params ClusterParams) ([]Cluster, error) {
	k := params.NumClusters
	if k <= 0 {
		k = 5
	}

	distanceMatrix := make([][]float64, len(points))
	for i := range distanceMatrix {
		distanceMatrix[i] = make([]float64, len(points))
		for j := range distanceMatrix[i] {
			distanceMatrix[i][j] = euclideanDistance(points[i].Vector, points[j].Vector)
		}
	}

	clusters := make([][]int, len(points))
	for i := range clusters {
		clusters[i] = []int{i}
	}

	for len(clusters) > k {
		minDist := math.MaxFloat64
		mergeA, mergeB := 0, 1

		for i := 0; i < len(clusters); i++ {
			for j := i + 1; j < len(clusters); j++ {
				d := clusterDistanceMatrix(clusters[i], clusters[j], distanceMatrix)
				if d < minDist {
					minDist = d
					mergeA, mergeB = i, j
				}
			}
		}

		clusters[mergeA] = append(clusters[mergeA], clusters[mergeB]...)
		clusters = append(clusters[:mergeB], clusters[mergeB+1:]...)
	}

	result := make([]Cluster, len(clusters))
	for i, indices := range clusters {
		result[i].ID = fmt.Sprintf("cluster_%d", i)
		for _, idx := range indices {
			result[i].Points = append(result[i].Points, points[idx])
		}
		result[i].Centroid = calculateCentroid(result[i].Points)
	}

	return result, nil
}

func clusterDistanceMatrix(a, b []int, matrix [][]float64) float64 {
	minDist := math.MaxFloat64

	for _, i := range a {
		for _, j := range b {
			if matrix[i][j] < minDist {
				minDist = matrix[i][j]
			}
		}
	}

	return minDist
}

type MiniBatchKMeans struct{}

func (c *MiniBatchKMeans) Name() string { return "minibatch_kmeans" }

func (c *MiniBatchKMeans) Cluster(ctx context.Context, points []Point, params ClusterParams) ([]Cluster, error) {
	k := params.NumClusters
	if k <= 0 {
		k = 5
	}

	batchSize := min(32, len(points))

	centroids := make([][]float64, k)
	for i := 0; i < k; i++ {
		centroids[i] = make([]float64, len(points[0].Vector))
		copy(centroids[i], points[i].Vector)
	}

	iterations := params.Iterations
	if iterations <= 0 {
		iterations = 100
	}

	for iter := 0; iter < iterations; iter++ {
		batch := make([]Point, batchSize)
		indices := make([]int, batchSize)
		for i := 0; i < batchSize; i++ {
			indices[i] = i * len(points) / batchSize
			batch[i] = points[indices[i]]
		}

		for _, p := range batch {
			minDist := math.MaxFloat64
			closest := 0

			for i, centroid := range centroids {
				d := euclideanDistance(p.Vector, centroid)
				if d < minDist {
					minDist = d
					closest = i
				}
			}

			for dim := range centroids[closest] {
				centroids[closest][dim] += 0.1 * (p.Vector[dim] - centroids[closest][dim])
			}
		}
	}

	result := make([]Cluster, k)
	for i := 0; i < k; i++ {
		result[i].ID = fmt.Sprintf("cluster_%d", i)
	}

	for _, p := range points {
		minDist := math.MaxFloat64
		closest := 0

		for i, centroid := range centroids {
			d := euclideanDistance(p.Vector, centroid)
			if d < minDist {
				minDist = d
				closest = i
			}
		}

		result[closest].Points = append(result[closest].Points, p)
		result[closest].Centroid = centroids[closest]
	}

	return result, nil
}
