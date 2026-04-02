package specializedagents

import (
	"strings"
	"testing"
)

func TestAgentOrchestrator(t *testing.T) {
	orch := NewAgentOrchestrator()

	agents := orch.ListAgents()
	if len(agents) != 5 {
		t.Errorf("Expected 5 agents, got %d", len(agents))
	}

	active := orch.GetActiveAgents()
	if len(active) != 5 {
		t.Errorf("Expected 5 active agents, got %d", len(active))
	}
}

func TestAnvilProcess(t *testing.T) {
	orch := NewAgentOrchestrator()
	input := "hello\t\tworld    foo"
	result := orch.ProcessThroughAgent(AgentAnvil, input)
	if len(result) >= len(input) {
		t.Error("Expected shorter output from anvil")
	}
}

func TestShieldProcess(t *testing.T) {
	orch := NewAgentOrchestrator()
	input := "user data\napi_key: secret123\nmore data"
	result := orch.ProcessThroughAgent(AgentShield, input)
	if !strings.Contains(result, "[REDACTED]") {
		t.Errorf("Expected shield to redact secrets, got: %s", result)
	}
}

func TestAgentActivation(t *testing.T) {
	orch := NewAgentOrchestrator()

	orch.DeactivateAgent(AgentAnvil)
	agent := orch.GetAgent(AgentAnvil)
	if agent.IsActive {
		t.Error("Anvil should be deactivated")
	}

	orch.ActivateAgent(AgentAnvil)
	agent = orch.GetAgent(AgentAnvil)
	if !agent.IsActive {
		t.Error("Anvil should be activated")
	}
}
