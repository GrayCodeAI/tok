package rightsizing

type ComplexityLevel int

const (
	ComplexityTrivial      ComplexityLevel = 0
	ComplexitySimple       ComplexityLevel = 1
	ComplexityModerate     ComplexityLevel = 2
	ComplexityIntermediate ComplexityLevel = 3
	ComplexityComplex      ComplexityLevel = 4
	ComplexityAdvanced     ComplexityLevel = 5
	ComplexityExpert       ComplexityLevel = 6
	ComplexitySpecialist   ComplexityLevel = 7
	ComplexityResearch     ComplexityLevel = 8
	ComplexityFrontier     ComplexityLevel = 9
)

type ComplexityFactors struct {
	CodeDepth         float64
	DomainSpecificity float64
	ReasoningDepth    float64
	ContextSize       float64
	ToolUsage         float64
	Ambiguity         float64
}

type ModelCapability struct {
	Name            string
	MaxComplexity   ComplexityLevel
	CostPer1KInput  float64
	CostPer1KOutput float64
	Strengths       []string
}

type Recommendation struct {
	CurrentModel     string
	RecommendedModel string
	ComplexityScore  ComplexityLevel
	EstimatedSavings float64
	Confidence       float64
	Reasoning        string
}

type ComplexityScorer struct{}

func NewComplexityScorer() *ComplexityScorer {
	return &ComplexityScorer{}
}

func (s *ComplexityScorer) Score(factors ComplexityFactors) ComplexityLevel {
	weights := map[string]float64{
		"code_depth":         0.25,
		"domain_specificity": 0.20,
		"reasoning_depth":    0.25,
		"context_size":       0.10,
		"tool_usage":         0.10,
		"ambiguity":          0.10,
	}

	score := factors.CodeDepth*weights["code_depth"] +
		factors.DomainSpecificity*weights["domain_specificity"] +
		factors.ReasoningDepth*weights["reasoning_depth"] +
		factors.ContextSize*weights["context_size"] +
		factors.ToolUsage*weights["tool_usage"] +
		factors.Ambiguity*weights["ambiguity"]

	clamped := score * 10
	if clamped < 0 {
		clamped = 0
	}
	if clamped > 9 {
		clamped = 9
	}

	return ComplexityLevel(int(clamped))
}

func (s *ComplexityScorer) ScoreFromText(text string) ComplexityLevel {
	lines := len(text)
	chars := len(text)
	codeBlocks := 0
	questions := 0

	for i := 0; i < len(text); i++ {
		if text[i] == '`' && i+2 < len(text) && text[i+1] == '`' && text[i+2] == '`' {
			codeBlocks++
		}
		if text[i] == '?' {
			questions++
		}
	}

	var factors ComplexityFactors
	if lines > 100 {
		factors.ContextSize = 0.8
	} else if lines > 50 {
		factors.ContextSize = 0.5
	} else if lines > 10 {
		factors.ContextSize = 0.3
	} else {
		factors.ContextSize = 0.1
	}

	factors.CodeDepth = float64(codeBlocks) / 10
	if factors.CodeDepth > 1 {
		factors.CodeDepth = 1
	}

	factors.ReasoningDepth = float64(questions) / 5
	if factors.ReasoningDepth > 1 {
		factors.ReasoningDepth = 1
	}

	if chars > 10000 {
		factors.DomainSpecificity = 0.7
	} else if chars > 5000 {
		factors.DomainSpecificity = 0.4
	} else {
		factors.DomainSpecificity = 0.2
	}

	return s.Score(factors)
}

var DefaultModels = []ModelCapability{
	{Name: "gpt-4o-mini", MaxComplexity: ComplexityIntermediate, CostPer1KInput: 0.00015, CostPer1KOutput: 0.0006},
	{Name: "claude-3-haiku", MaxComplexity: ComplexityIntermediate, CostPer1KInput: 0.00025, CostPer1KOutput: 0.00125},
	{Name: "gpt-4o", MaxComplexity: ComplexityExpert, CostPer1KInput: 0.0025, CostPer1KOutput: 0.01},
	{Name: "claude-3-5-sonnet", MaxComplexity: ComplexitySpecialist, CostPer1KInput: 0.003, CostPer1KOutput: 0.015},
	{Name: "claude-3-opus", MaxComplexity: ComplexityFrontier, CostPer1KInput: 0.015, CostPer1KOutput: 0.075},
	{Name: "gemini-1.5-flash", MaxComplexity: ComplexityModerate, CostPer1KInput: 0.000075, CostPer1KOutput: 0.0003},
	{Name: "gemini-1.5-pro", MaxComplexity: ComplexityExpert, CostPer1KInput: 0.00125, CostPer1KOutput: 0.005},
}

type ModelRecommendationEngine struct {
	models []ModelCapability
}

func NewModelRecommendationEngine() *ModelRecommendationEngine {
	return &ModelRecommendationEngine{
		models: DefaultModels,
	}
}

func (e *ModelRecommendationEngine) Recommend(complexity ComplexityLevel, currentModel string) *Recommendation {
	var best ModelCapability
	var current ModelCapability
	var bestSavings float64

	for _, m := range e.models {
		if m.Name == currentModel {
			current = m
		}
		if m.MaxComplexity >= complexity {
			cost := m.CostPer1KInput + m.CostPer1KOutput
			if best.Name == "" || cost < bestSavings {
				best = m
				bestSavings = cost
			}
		}
	}

	if current.Name == "" || best.Name == "" || best.Name == currentModel {
		return &Recommendation{
			CurrentModel:     currentModel,
			RecommendedModel: currentModel,
			ComplexityScore:  complexity,
			Confidence:       0.5,
			Reasoning:        "Current model is appropriate for this complexity level",
		}
	}

	currentCost := current.CostPer1KInput + current.CostPer1KOutput
	recommendedCost := best.CostPer1KInput + best.CostPer1KOutput
	savings := ((currentCost - recommendedCost) / currentCost) * 100

	return &Recommendation{
		CurrentModel:     currentModel,
		RecommendedModel: best.Name,
		ComplexityScore:  complexity,
		EstimatedSavings: savings,
		Confidence:       0.8,
		Reasoning:        "Lower-cost model can handle complexity level " + string(rune(complexity)),
	}
}

func (e *ModelRecommendationEngine) GetAllModels() []ModelCapability {
	return e.models
}
