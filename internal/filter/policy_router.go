package filter

import "strings"

// PolicyRoute captures lightweight routing decisions derived from raw input.
type PolicyRoute struct {
	QueryIntent string
	Class       string
}

// PolicyRouter infers command/output class and a query intent hint.
type PolicyRouter struct{}

func NewPolicyRouter() *PolicyRouter {
	return &PolicyRouter{}
}

// Route infers route metadata. If explicitIntent is set, it always wins.
func (r *PolicyRouter) Route(input, explicitIntent string) PolicyRoute {
	if explicitIntent != "" {
		return PolicyRoute{QueryIntent: explicitIntent, Class: "explicit"}
	}

	text := strings.ToLower(input)
	switch {
	case containsAny(text, "panic:", "traceback", "exception"):
		return PolicyRoute{QueryIntent: "debug", Class: "failure"}
	case containsAny(text, "vulnerability", "cve-", "security", "gosec", "govulncheck", "audit"):
		return PolicyRoute{QueryIntent: "review", Class: "security"}
	case containsAny(text, "test", "assert", "pytest", "jest", "go test", "vitest"):
		return PolicyRoute{QueryIntent: "test", Class: "test"}
	case containsAny(text, "build", "compile", "linker", "bundl", "transpil", "webpack", "tsc"):
		return PolicyRoute{QueryIntent: "build", Class: "build"}
	case containsAny(text, "error:", "fatal:", "failed"):
		return PolicyRoute{QueryIntent: "debug", Class: "failure"}
	case containsAny(text, "diff --git", "@@", "changed files", "pull request"):
		return PolicyRoute{QueryIntent: "review", Class: "diff"}
	case containsAny(text, "deploy", "kubectl", "helm", "terraform apply", "rollout"):
		return PolicyRoute{QueryIntent: "deploy", Class: "ops"}
	case containsAny(text, "grep ", "rg ", "find ", "search ", "where is", "locate "):
		return PolicyRoute{QueryIntent: "search", Class: "search"}
	default:
		return PolicyRoute{QueryIntent: "", Class: "generic"}
	}
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}

func (p *PipelineCoordinator) applyPolicyRouting(input string) {
	if p.policyRouter == nil {
		return
	}

	route := p.policyRouter.Route(input, p.config.QueryIntent)
	p.runtimeQueryIntent = route.QueryIntent
	if route.QueryIntent == "" {
		return
	}

	if p.goalDrivenFilter == nil && p.config.EnableGoalDriven {
		p.goalDrivenFilter = NewGoalDrivenFilter(route.QueryIntent)
		p.layers[2] = filterLayer{p.goalDrivenFilter, "3_goal_driven"}
	}

	if p.contrastiveFilter == nil && p.config.EnableContrastive {
		p.contrastiveFilter = NewContrastiveFilter(route.QueryIntent)
		p.layers[4] = filterLayer{p.contrastiveFilter, "5_contrastive"}
	}

	if p.questionAwareFilter == nil && p.config.EnableQuestionAware {
		p.questionAwareFilter = NewQuestionAwareFilter(route.QueryIntent)
	}
}
