package specializedagents

import (
	"strings"
)

type AgentRole string

const (
	AgentAnvil  AgentRole = "anvil"
	AgentShield AgentRole = "shield"
	AgentHarbor AgentRole = "harbor"
	AgentBeacon AgentRole = "beacon"
	AgentLens   AgentRole = "lens"
)

type SpecializedAgent struct {
	Role        AgentRole `json:"role"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
}

type AgentOrchestrator struct {
	agents map[AgentRole]*SpecializedAgent
}

func NewAgentOrchestrator() *AgentOrchestrator {
	o := &AgentOrchestrator{
		agents: make(map[AgentRole]*SpecializedAgent),
	}
	o.registerDefaultAgents()
	return o
}

func (o *AgentOrchestrator) registerDefaultAgents() {
	o.agents[AgentAnvil] = &SpecializedAgent{
		Role:        AgentAnvil,
		Name:        "Anvil",
		Description: "Compression pipeline optimizer - analyzes and tunes compression settings",
		IsActive:    true,
	}
	o.agents[AgentShield] = &SpecializedAgent{
		Role:        AgentShield,
		Name:        "Shield",
		Description: "Security scanner - detects PII, secrets, injection attacks",
		IsActive:    true,
	}
	o.agents[AgentHarbor] = &SpecializedAgent{
		Role:        AgentHarbor,
		Name:        "Harbor",
		Description: "Egress firewall manager - controls outbound connections",
		IsActive:    true,
	}
	o.agents[AgentBeacon] = &SpecializedAgent{
		Role:        AgentBeacon,
		Name:        "Beacon",
		Description: "Monitoring and alerting - tracks metrics and triggers alerts",
		IsActive:    true,
	}
	o.agents[AgentLens] = &SpecializedAgent{
		Role:        AgentLens,
		Name:        "Lens",
		Description: "Analytics and reporting - generates insights and reports",
		IsActive:    true,
	}
}

func (o *AgentOrchestrator) GetAgent(role AgentRole) *SpecializedAgent {
	return o.agents[role]
}

func (o *AgentOrchestrator) ListAgents() []*SpecializedAgent {
	var result []*SpecializedAgent
	for _, a := range o.agents {
		result = append(result, a)
	}
	return result
}

func (o *AgentOrchestrator) ActivateAgent(role AgentRole) {
	if agent, ok := o.agents[role]; ok {
		agent.IsActive = true
	}
}

func (o *AgentOrchestrator) DeactivateAgent(role AgentRole) {
	if agent, ok := o.agents[role]; ok {
		agent.IsActive = false
	}
}

func (o *AgentOrchestrator) GetActiveAgents() []*SpecializedAgent {
	var result []*SpecializedAgent
	for _, a := range o.agents {
		if a.IsActive {
			result = append(result, a)
		}
	}
	return result
}

func (o *AgentOrchestrator) ProcessThroughAgent(role AgentRole, input string) string {
	agent := o.GetAgent(role)
	if agent == nil || !agent.IsActive {
		return input
	}

	switch role {
	case AgentAnvil:
		return o.anvilProcess(input)
	case AgentShield:
		return o.shieldProcess(input)
	case AgentHarbor:
		return o.harborProcess(input)
	case AgentBeacon:
		return o.beaconProcess(input)
	case AgentLens:
		return o.lensProcess(input)
	}
	return input
}

func (o *AgentOrchestrator) anvilProcess(input string) string {
	result := strings.ReplaceAll(input, "\t", "  ")
	for strings.Contains(result, "    ") {
		result = strings.ReplaceAll(result, "    ", "  ")
	}
	return result
}

func (o *AgentOrchestrator) shieldProcess(input string) string {
	lines := strings.Split(input, "\n")
	var filtered []string
	for _, line := range lines {
		if strings.Contains(line, "api_key") || strings.Contains(line, "secret") {
			filtered = append(filtered, "[REDACTED]")
		} else {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (o *AgentOrchestrator) harborProcess(input string) string {
	return input
}

func (o *AgentOrchestrator) beaconProcess(input string) string {
	return input
}

func (o *AgentOrchestrator) lensProcess(input string) string {
	return input
}
