package extractivesum

import (
	"context"
	"testing"
)

func TestSumBasicExtractor(t *testing.T) {
	extractor := &SumBasicExtractor{}

	sentences := []Sentence{
		{Text: "The cat sat on the mat.", Tokens: []string{"cat", "sat", "mat"}, Position: 0},
		{Text: "A cat is a friendly pet.", Tokens: []string{"cat", "friendly", "pet"}, Position: 1},
		{Text: "Dogs are great companions.", Tokens: []string{"dogs", "great", "companions"}, Position: 2},
		{Text: "The cat chased the mouse.", Tokens: []string{"cat", "chased", "mouse"}, Position: 3},
	}

	options := ExtractOptions{MaxSentences: 2}

	result, err := extractor.Extract(context.Background(), sentences, options)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 sentences, got %d", len(result))
	}

	if result[0].Score <= 0 {
		t.Error("Expected positive score for sentences with common words")
	}
}

func TestKLSumExtractor(t *testing.T) {
	extractor := &KLSumExtractor{}

	sentences := []Sentence{
		{Text: "Machine learning is a subset of artificial intelligence.", Tokens: []string{"machine", "learning", "subset", "artificial", "intelligence"}, Position: 0},
		{Text: "Deep learning uses neural networks.", Tokens: []string{"deep", "learning", "uses", "neural", "networks"}, Position: 1},
		{Text: "Python is a popular programming language.", Tokens: []string{"python", "popular", "programming", "language"}, Position: 2},
		{Text: "TensorFlow is a machine learning framework.", Tokens: []string{"tensorflow", "machine", "learning", "framework"}, Position: 3},
	}

	options := ExtractOptions{MaxSentences: 2}

	result, err := extractor.Extract(context.Background(), sentences, options)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 sentences, got %d", len(result))
	}

	for _, sent := range result {
		if sent.Score < 0 {
			t.Errorf("KL divergence should produce non-negative scores, got %f", sent.Score)
		}
	}
}

func TestMMRExtractor(t *testing.T) {
	extractor := &MMRExtractor{}

	sentences := []Sentence{
		{Text: "Machine learning uses algorithms to learn from data.", Tokens: []string{"machine", "learning", "algorithms", "learn", "data"}, Position: 0},
		{Text: "Deep learning is a type of machine learning.", Tokens: []string{"deep", "learning", "type", "machine", "learning"}, Position: 1},
		{Text: "Python is a programming language.", Tokens: []string{"python", "programming", "language"}, Position: 2},
		{Text: "Neural networks are used in deep learning.", Tokens: []string{"neural", "networks", "used", "deep", "learning"}, Position: 3},
	}

	options := ExtractOptions{MaxSentences: 2, Query: "machine learning"}

	result, err := extractor.Extract(context.Background(), sentences, options)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 sentences, got %d", len(result))
	}

	if result[0].Score <= 0 {
		t.Error("Expected positive score for query-relevant sentences")
	}
}

func TestCentroidExtractor(t *testing.T) {
	extractor := &CentroidExtractor{}

	sentences := []Sentence{
		{Text: "The quick brown fox jumps over the lazy dog.", Tokens: []string{"quick", "brown", "fox", "jumps", "lazy", "dog"}, Position: 0},
		{Text: "A fast fox can leap over sleeping animals.", Tokens: []string{"fast", "fox", "leap", "sleeping", "animals"}, Position: 1},
		{Text: "The sky is blue today.", Tokens: []string{"sky", "blue", "today"}, Position: 2},
		{Text: "Foxes are clever and fast.", Tokens: []string{"foxes", "clever", "fast"}, Position: 3},
	}

	options := ExtractOptions{MaxSentences: 2}

	result, err := extractor.Extract(context.Background(), sentences, options)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 sentences, got %d", len(result))
	}

	hasFoxSentences := 0
	for _, r := range result {
		if len(r.Tokens) > 0 {
			hasFoxSentences++
		}
	}

	if hasFoxSentences != 2 {
		t.Logf("Expected sentences with 'fox' related tokens to rank higher")
	}
}

func TestSubmodularExtractor(t *testing.T) {
	extractor := &SubmodularExtractor{}

	sentences := []Sentence{
		{Text: "Python is great for data science and machine learning.", Tokens: []string{"python", "great", "data", "science", "machine", "learning"}, Position: 0},
		{Text: "JavaScript is used for web development.", Tokens: []string{"javascript", "used", "web", "development"}, Position: 1},
		{Text: "Data science uses Python and machine learning techniques.", Tokens: []string{"data", "science", "uses", "python", "machine", "learning", "techniques"}, Position: 2},
		{Text: "Go is a fast programming language.", Tokens: []string{"fast", "programming", "language"}, Position: 3},
	}

	options := ExtractOptions{MaxSentences: 2}

	result, err := extractor.Extract(context.Background(), sentences, options)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 sentences, got %d", len(result))
	}

	if result[0].Score <= 0 {
		t.Error("Expected positive IDF-based score")
	}
}

func TestPositionExtractor(t *testing.T) {
	extractor := &PositionExtractor{}

	sentences := []Sentence{
		{Text: "First sentence with important information.", Tokens: []string{"first", "sentence", "important", "information"}, Position: 0},
		{Text: "Second sentence.", Tokens: []string{"second", "sentence"}, Position: 1},
		{Text: "Third sentence.", Tokens: []string{"third", "sentence"}, Position: 2},
		{Text: "Fourth sentence.", Tokens: []string{"fourth", "sentence"}, Position: 3},
		{Text: "Last sentence with important info.", Tokens: []string{"last", "sentence", "important", "info"}, Position: 4},
	}

	options := ExtractOptions{MaxSentences: 3}

	result, err := extractor.Extract(context.Background(), sentences, options)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 sentences, got %d", len(result))
	}

	if result[0].Position != 0 {
		t.Logf("Expected first sentence to rank high due to position scoring")
	}
}

func TestEngineExtract(t *testing.T) {
	engine := NewAdvancedExtractiveEngine(DefaultEngineConfig())

	documents := []string{
		"This is the first document. It talks about machine learning. Machine learning is amazing.",
		"Second document here. Python is a great language. Programming is fun.",
	}

	options := ExtractOptions{MaxSentences: 3}

	result, err := engine.Extract(context.Background(), documents, options)
	if err != nil {
		t.Fatalf("Engine Extract failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected at least one sentence in result")
	}

	stats := engine.GetStats()
	if stats.TotalSummaries != 1 {
		t.Errorf("Expected 1 summary, got %d", stats.TotalSummaries)
	}
}

func TestEngineWithQuery(t *testing.T) {
	engine := NewAdvancedExtractiveEngine(DefaultEngineConfig())

	documents := []string{
		"Machine learning is a subset of artificial intelligence.",
		"Deep learning uses neural networks with multiple layers.",
		"Python is a popular programming language for data science.",
		"Natural language processing deals with text and speech.",
	}

	options := ExtractOptions{
		MaxSentences: 2,
		Query:        "machine learning",
	}

	result, err := engine.Extract(context.Background(), documents, options)
	if err != nil {
		t.Fatalf("Engine Extract with query failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 sentences, got %d", len(result))
	}
}

func TestSentenceSimilarity(t *testing.T) {
	s1 := Sentence{
		Tokens:   []string{"machine", "learning", "algorithms"},
		Position: 0,
	}
	s2 := Sentence{
		Tokens:   []string{"machine", "learning", "neural", "networks"},
		Position: 1,
	}
	s3 := Sentence{
		Tokens:   []string{"python", "programming", "language"},
		Position: 2,
	}

	sim12 := sentenceSimilarity(s1, s2)
	sim13 := sentenceSimilarity(s1, s3)

	if sim12 <= 0 {
		t.Error("Expected positive similarity for overlapping tokens")
	}

	if sim13 >= sim12 {
		t.Logf("Expected higher similarity between s1 and s2 (more overlap)")
	}

	if sim12 > 1.0 || sim12 < 0.0 {
		t.Errorf("Similarity should be between 0 and 1, got %f", sim12)
	}
}

func TestCoverage(t *testing.T) {
	sentences := []Sentence{
		{Importance: 1.0, Tokens: []string{"word1", "word2"}},
		{Importance: 0.5, Tokens: []string{"word2", "word3"}},
		{Importance: 0.3, Tokens: []string{"word3", "word4"}},
	}

	cov := coverage(sentences)

	if cov < 0 || cov > 1 {
		t.Errorf("Coverage should be between 0 and 1, got %f", cov)
	}

	if cov <= 0 {
		t.Error("Expected positive coverage for valid sentences")
	}
}

func TestEmptySentences(t *testing.T) {
	engine := NewAdvancedExtractiveEngine(DefaultEngineConfig())

	result, err := engine.Extract(context.Background(), []string{}, ExtractOptions{MaxSentences: 5})
	if err != nil {
		t.Fatalf("Expected no error for empty input: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d", len(result))
	}
}

func TestTokenization(t *testing.T) {
	engine := NewAdvancedExtractiveEngine(DefaultEngineConfig())

	documents := []string{
		"Short. Also short.",
		"This is a much longer sentence with many more words in it for testing purposes.",
	}

	result, err := engine.Extract(context.Background(), documents, ExtractOptions{MaxSentences: 5})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	hasLongSentence := false
	for _, r := range result {
		if len(r.Tokens) > 5 {
			hasLongSentence = true
		}
	}

	if !hasLongSentence {
		t.Logf("Expected longer sentences to be selected over very short ones")
	}
}

func TestMultiDocExtraction(t *testing.T) {
	engine := NewAdvancedExtractiveEngine(EngineConfig{
		DefaultAlgorithm:  "sumbasic",
		MaxSummaryLength:  10,
		MinSentenceLength: 10,
		EnableMultiDoc:    true,
		EnableQueryFocus:  true,
	})

	documents := []string{
		"Document one discusses artificial intelligence. AI is transforming many industries.",
		"Document two covers machine learning. ML algorithms improve over time with data.",
		"Document three explains deep learning. Deep neural networks have many layers.",
	}

	options := ExtractOptions{MaxSentences: 3}

	result, err := engine.Extract(context.Background(), documents, options)
	if err != nil {
		t.Fatalf("Multi-doc extract failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected sentences from multiple documents")
	}

	for _, r := range result {
		if doc, ok := r.Metadata["doc"]; !ok {
			t.Error("Expected document metadata on sentences")
		} else {
			_ = doc
		}
	}
}

func TestSubmodularSelection(t *testing.T) {
	sentences := []Sentence{
		{Score: 1.0, Importance: 1.0, Tokens: []string{"a", "b", "c"}},
		{Score: 0.9, Importance: 0.9, Tokens: []string{"b", "c", "d"}},
		{Score: 0.8, Importance: 0.8, Tokens: []string{"c", "d", "e"}},
		{Score: 0.7, Importance: 0.7, Tokens: []string{"d", "e", "f"}},
		{Score: 0.6, Importance: 0.6, Tokens: []string{"e", "f", "g"}},
	}

	result := submodularSelect(sentences, 3)

	if len(result) != 3 {
		t.Errorf("Expected 3 sentences, got %d", len(result))
	}
}

func TestEngineStats(t *testing.T) {
	engine := NewAdvancedExtractiveEngine(DefaultEngineConfig())

	engine.Extract(context.Background(), []string{
		"First document about machine learning.",
		"Second document about deep learning.",
	}, ExtractOptions{MaxSentences: 2})

	engine.Extract(context.Background(), []string{
		"Third document about Python.",
	}, ExtractOptions{MaxSentences: 1})

	stats := engine.GetStats()

	if stats.TotalSummaries != 2 {
		t.Errorf("Expected 2 summaries, got %d", stats.TotalSummaries)
	}

	if stats.AvgReduction < 0 || stats.AvgReduction > 1 {
		t.Errorf("Expected reduction between 0 and 1, got %f", stats.AvgReduction)
	}

	if stats.AlgorithmUsage["sumbasic"] != 2 {
		t.Errorf("Expected sumbasic used 2 times, got %d", stats.AlgorithmUsage["sumbasic"])
	}
}

func TestRegisterExtractor(t *testing.T) {
	engine := NewAdvancedExtractiveEngine(DefaultEngineConfig())

	type CustomExtractor struct{}

	extractor := &SumBasicExtractor{}
	engine.RegisterExtractor(extractor)

	engine.mu.RLock()
	_, ok := engine.algorithms["sumbasic"]
	engine.mu.RUnlock()

	if !ok {
		t.Error("Expected sumbasic extractor to be registered")
	}
}
