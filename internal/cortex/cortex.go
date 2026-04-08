package cortex

type Finding struct {
	ContentType string
	Language    string
	Rule        string
	Severity    string
	Message     string
}

type GateRegistry struct{}

func NewGateRegistry() *GateRegistry {
	return &GateRegistry{}
}

func (g *GateRegistry) ApplyGates(content string) string {
	return content
}

func (g *GateRegistry) Analyze(content string) *Finding {
	return &Finding{
		ContentType: "text",
		Language:    "unknown",
	}
}

func (g *GateRegistry) GetApplicableGates(content string) []string {
	return nil
}
