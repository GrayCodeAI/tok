package onboarding

import "testing"

func TestOnboardingFlow(t *testing.T) {
	flow := NewOnboardingFlow()

	step := flow.CurrentStep()
	if step == nil {
		t.Fatal("Expected first step")
	}
	if step.Title != "Install TokMan" {
		t.Errorf("Expected 'Install TokMan', got %s", step.Title)
	}

	flow.CompleteStep()
	step = flow.CurrentStep()
	if step.Title != "Initialize Config" {
		t.Errorf("Expected 'Initialize Config', got %s", step.Title)
	}

	if flow.Progress() == 0 {
		t.Error("Expected non-zero progress")
	}

	if flow.IsComplete() {
		t.Error("Should not be complete after 1 step")
	}
}

func TestOnboardingFlowComplete(t *testing.T) {
	flow := NewOnboardingFlow()
	for !flow.IsComplete() {
		flow.CompleteStep()
	}
	if flow.Progress() != 100 {
		t.Errorf("Expected 100%% progress, got %.1f", flow.Progress())
	}
}

func TestInteractiveTutorial(t *testing.T) {
	tut := NewInteractiveTutorial()

	if tut.CurrentTopic() != "What is TokMan?" {
		t.Errorf("Expected first topic, got %s", tut.CurrentTopic())
	}

	tut.Next()
	if tut.CurrentTopic() != "How Compression Works" {
		t.Errorf("Expected second topic, got %s", tut.CurrentTopic())
	}

	tut.Prev()
	if tut.CurrentTopic() != "What is TokMan?" {
		t.Errorf("Expected first topic after prev, got %s", tut.CurrentTopic())
	}
}

func TestHelpSystem(t *testing.T) {
	h := NewHelpSystem()

	help := h.Get("compression")
	if help == "" {
		t.Error("Expected compression help")
	}

	topics := h.ListTopics()
	if len(topics) < 5 {
		t.Errorf("Expected at least 5 topics, got %d", len(topics))
	}

	results := h.Search("filter")
	if len(results) == 0 {
		t.Error("Expected search results for 'filter'")
	}
}
