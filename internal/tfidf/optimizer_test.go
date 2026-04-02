package tfidf

import "testing"

func TestTokenize(t *testing.T) {
	tokens := Tokenize("the quick brown fox jumps over the lazy dog")
	if len(tokens) == 0 {
		t.Error("Expected tokens")
	}
	if tokens["the"] != 2 {
		t.Errorf("Expected 2 occurrences of 'the', got %d", tokens["the"])
	}
}

func TestTFIDFOptimizer(t *testing.T) {
	opt := NewTFIDFOptimizer()

	opt.AddDocument(Tokenize("the cat sat on the mat"))
	opt.AddDocument(Tokenize("the dog played in the park"))
	opt.AddDocument(Tokenize("the bird flew over the tree"))
	opt.BuildIndex()

	scores := opt.ScoreTerms(Tokenize("the cat"))
	if len(scores) == 0 {
		t.Error("Expected scores")
	}

	top := opt.SelectTopTerms(scores, 3)
	if len(top) == 0 {
		t.Error("Expected top terms")
	}
}

func TestToolSchemaOptimizer(t *testing.T) {
	opt := NewToolSchemaOptimizer()
	opt.AddToolDefinition("bash", "Execute bash commands")
	opt.AddToolDefinition("read_file", "Read file contents")
	opt.AddToolDefinition("write_file", "Write file contents")

	results := opt.OptimizeForContext("read file contents", 5)
	if len(results) == 0 {
		t.Error("Expected optimized results")
	}
}
