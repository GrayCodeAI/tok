package cortex

type Finding struct {
	Rule     string
	Severity string
	Message  string
}

type GateRegistry struct{}

func NewGateRegistry() *GateRegistry {
	return &GateRegistry{}
}

func (g *GateRegistry) ApplyGates(content string) string {
	return content
}

func (g *GateRegistry) Analyze(content string) *Finding {
	return nil
}

func (g *GateRegistry) GetApplicableGates(content string) []string {
	return nil
}
