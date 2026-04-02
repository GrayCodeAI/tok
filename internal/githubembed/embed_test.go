package githubembed

import "testing"

func TestEmbedEngine(t *testing.T) {
	engine := NewEmbedEngine()

	badge := engine.GenerateBadge("alice", 50000, "gold")
	if badge == nil {
		t.Fatal("Expected non-nil badge")
	}
	if badge.Type != EmbedBadge {
		t.Errorf("Expected badge type, got %s", badge.Type)
	}
	if badge.Content == "" {
		t.Error("Expected non-empty SVG content")
	}
}

func TestEmbedEngineCard(t *testing.T) {
	engine := NewEmbedEngine()

	card := engine.GenerateCard("bob", 100000, "platinum", 500)
	if card == nil {
		t.Fatal("Expected non-nil card")
	}
	if card.Type != EmbedCard {
		t.Errorf("Expected card type, got %s", card.Type)
	}
}

func TestEmbedEngine3D(t *testing.T) {
	engine := NewEmbedEngine()

	days := []int64{100, 200, 150, 300, 50}
	graph := engine.Generate3DGraph("charlie", days)
	if graph == nil {
		t.Fatal("Expected non-nil 3D graph")
	}
	if graph.Type != Embed3D {
		t.Errorf("Expected 3D type, got %s", graph.Type)
	}
}

func TestEmbedEngineMarkdown(t *testing.T) {
	engine := NewEmbedEngine()
	badge := engine.GenerateBadge("user", 1000, "silver")

	md := engine.EmbedMarkdown("user", badge)
	if md == "" {
		t.Error("Expected non-empty markdown")
	}
}
