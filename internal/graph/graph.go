package graph

type Graph struct{}

func New() *Graph {
	return &Graph{}
}

type GraphStats map[string]any

func NewProjectGraph(root string) *Graph {
	return &Graph{}
}

func FormatGraphStats(stats GraphStats) string {
	return "Graph: 0 nodes, 0 edges"
}

func (g *Graph) Analyze(root string) error {
	return nil
}

func (g *Graph) AnalyzeWithDepth(root string, depth int) error {
	return nil
}

func (g *Graph) Stats() GraphStats {
	return GraphStats{"nodes": 0, "edges": 0}
}

func (g *Graph) FindRelatedFiles(file string, max int) []string {
	return nil
}

func (g *Graph) ImpactAnalysis(file string) []string {
	return nil
}
