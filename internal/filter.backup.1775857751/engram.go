package filter

import (
	"fmt"
	"strings"
)

// EngramMemory implements LLM-driven observational memory system.
// Inspired by claw-compactor's Engram Observer/Reflector.
type EngramMemory struct {
	observations []Observation
	reflections  []Reflection
	threshold    float64
}

// Observation represents a single observation from LLM output.
type Observation struct {
	Content    string
	Importance float64
	Timestamp  string
}

// Reflection represents a consolidated insight from multiple observations.
type Reflection struct {
	Insight    string
	Sources    []int
	Confidence float64
}

// NewEngramMemory creates a new engram memory system.
func NewEngramMemory(threshold float64) *EngramMemory {
	if threshold == 0 {
		threshold = 0.7
	}
	return &EngramMemory{threshold: threshold}
}

// Observe records an observation from LLM output.
func (em *EngramMemory) Observe(content string, importance float64) {
	if importance >= em.threshold {
		em.observations = append(em.observations, Observation{
			Content:    content,
			Importance: importance,
		})
	}
}

// Reflect consolidates observations into insights.
func (em *EngramMemory) Reflect() []Reflection {
	if len(em.observations) < 2 {
		return nil
	}
	groups := em.groupObservations()
	var reflections []Reflection
	for _, group := range groups {
		if len(group) >= 2 {
			reflections = append(reflections, Reflection{
				Insight:    em.synthesizeInsight(group),
				Sources:    group,
				Confidence: em.calculateConfidence(group),
			})
		}
	}
	return reflections
}

// TieredSummary generates L0/L1/L2 tiered summaries.
func (em *EngramMemory) TieredSummary() map[string]string {
	summaries := make(map[string]string)
	summaries["L0"] = em.l0Summary()
	summaries["L1"] = em.l1Summary()
	summaries["L2"] = em.l2Summary()
	return summaries
}

func (em *EngramMemory) groupObservations() [][]int {
	groups := [][]int{{0}}
	for i := 1; i < len(em.observations); i++ {
		lastGroup := groups[len(groups)-1]
		lastIdx := lastGroup[len(lastGroup)-1]
		if em.observations[i].Importance == em.observations[lastIdx].Importance {
			groups[len(groups)-1] = append(groups[len(groups)-1], i)
		} else {
			groups = append(groups, []int{i})
		}
	}
	return groups
}

func (em *EngramMemory) synthesizeInsight(group []int) string {
	var contents []string
	for _, idx := range group {
		contents = append(contents, em.observations[idx].Content)
	}
	return strings.Join(contents, " | ")
}

func (em *EngramMemory) calculateConfidence(group []int) float64 {
	if len(group) == 0 {
		return 0
	}
	var sum float64
	for _, idx := range group {
		sum += em.observations[idx].Importance
	}
	return sum / float64(len(group))
}

func (em *EngramMemory) l0Summary() string {
	if len(em.observations) == 0 {
		return "No observations"
	}
	return fmt.Sprintf("%d observations, %d reflections", len(em.observations), len(em.reflections))
}

func (em *EngramMemory) l1Summary() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Observations: %d\n", len(em.observations)))
	sb.WriteString(fmt.Sprintf("Reflections: %d\n", len(em.reflections)))
	for i, obs := range em.observations {
		if i >= 5 {
			sb.WriteString(fmt.Sprintf("... and %d more\n", len(em.observations)-5))
			break
		}
		limit := len(obs.Content)
		if limit > 50 {
			limit = 50
		}
		sb.WriteString(fmt.Sprintf("  - %.1f: %s\n", obs.Importance, obs.Content[:limit]))
	}
	return sb.String()
}

func (em *EngramMemory) l2Summary() string {
	var sb strings.Builder
	sb.WriteString(em.l1Summary())
	sb.WriteString("\nReflections:\n")
	for _, r := range em.reflections {
		sb.WriteString(fmt.Sprintf("  - %.0f%% confidence: %s\n", r.Confidence*100, r.Insight))
	}
	return sb.String()
}
