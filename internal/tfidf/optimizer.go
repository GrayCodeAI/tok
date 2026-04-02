package tfidf

import (
	"math"
	"strings"
)

type TFIDFOptimizer struct {
	documents []map[string]int
	idf       map[string]float64
}

func NewTFIDFOptimizer() *TFIDFOptimizer {
	return &TFIDFOptimizer{
		idf: make(map[string]float64),
	}
}

func (o *TFIDFOptimizer) AddDocument(doc map[string]int) {
	o.documents = append(o.documents, doc)
}

func (o *TFIDFOptimizer) BuildIndex() {
	N := float64(len(o.documents))
	if N == 0 {
		return
	}

	docFreq := make(map[string]int)
	for _, doc := range o.documents {
		seen := make(map[string]bool)
		for term := range doc {
			if !seen[term] {
				docFreq[term]++
				seen[term] = true
			}
		}
	}

	for term, df := range docFreq {
		o.idf[term] = math.Log(N / float64(df))
	}
}

func (o *TFIDFOptimizer) ScoreTerms(terms map[string]int) map[string]float64 {
	scores := make(map[string]float64)
	totalTerms := 0
	for _, count := range terms {
		totalTerms += count
	}

	if totalTerms == 0 {
		return scores
	}

	for term, count := range terms {
		tf := float64(count) / float64(totalTerms)
		idf := o.idf[term]
		scores[term] = tf * idf
	}

	return scores
}

func (o *TFIDFOptimizer) SelectTopTerms(scores map[string]float64, n int) []string {
	type termScore struct {
		term  string
		score float64
	}

	var ts []termScore
	for term, score := range scores {
		ts = append(ts, termScore{term, score})
	}

	for i := 0; i < len(ts); i++ {
		for j := i + 1; j < len(ts); j++ {
			if ts[j].score > ts[i].score {
				ts[i], ts[j] = ts[j], ts[i]
			}
		}
	}

	if n > len(ts) {
		n = len(ts)
	}

	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = ts[i].term
	}
	return result
}

func Tokenize(text string) map[string]int {
	tokens := make(map[string]int)
	words := strings.Fields(strings.ToLower(text))
	for _, word := range words {
		word = strings.Trim(word, ".,;:!?\"'()[]{}")
		if len(word) > 2 {
			tokens[word]++
		}
	}
	return tokens
}

type ToolSchemaOptimizer struct {
	tfidf *TFIDFOptimizer
}

func NewToolSchemaOptimizer() *ToolSchemaOptimizer {
	return &ToolSchemaOptimizer{
		tfidf: NewTFIDFOptimizer(),
	}
}

func (o *ToolSchemaOptimizer) AddToolDefinition(name string, description string) {
	tokens := Tokenize(description)
	o.tfidf.AddDocument(tokens)
}

func (o *ToolSchemaOptimizer) OptimizeForContext(context string, topN int) []string {
	o.tfidf.BuildIndex()
	contextTokens := Tokenize(context)
	scores := o.tfidf.ScoreTerms(contextTokens)
	return o.tfidf.SelectTopTerms(scores, topN)
}
