# TokMan Complete Code Analysis - Part 4: Individual Layer Deep Dive

## 4. Compression Layers (1-20)

### Layer 1: Entropy Filtering

**Research**: Selective Context (Mila 2023)

**Algorithm**: Remove low-information tokens based on Shannon entropy

```go
type EntropyFilter struct {
    threshold float64  // Default: 0.5
}

func (f *EntropyFilter) Apply(input string, mode Mode) (string, int) {
    tokens := tokenize(input)
    
    // Calculate entropy for each token
    entropies := make([]float64, len(tokens))
    for i, token := range tokens {
        entropies[i] = calculateEntropy(token)
    }
    
    // Filter low-entropy tokens
    var filtered []string
    for i, token := range tokens {
        if entropies[i] >= f.threshold || isStructural(token) {
            filtered = append(filtered, token)
        }
    }
    
    output := strings.Join(filtered, " ")
    saved := len(input) - len(output)
    return output, saved / 4  // Estimate tokens
}

func calculateEntropy(token string) float64 {
    freq := make(map[rune]int)
    for _, r := range token {
        freq[r]++
    }
    
    var entropy float64
    n := float64(len(token))
    for _, count := range freq {
        p := float64(count) / n
        entropy -= p * math.Log2(p)
    }
    
    return entropy
}
```

**Performance**: O(n) where n = token count
**Typical reduction**: 10-20%
**Best for**: Verbose logs, repetitive output

**Issues**:
- ❌ Tokenization is expensive (regex-based)
- ❌ No caching of entropy calculations
- ❌ Structural token detection is naive

**Improvements**:
```go
type EntropyFilter struct {
    threshold    float64
    entropyCache sync.Map  // Cache entropy calculations
    tokenizer    *Tokenizer // Reusable tokenizer
}

func (f *EntropyFilter) Apply(input string, mode Mode) (string, int) {
    // Fast path: check cache
    if cached, ok := f.entropyCache.Load(input); ok {
        return cached.(string), estimateTokens(input) - estimateTokens(cached.(string))
    }
    
    // Use pre-compiled tokenizer
    tokens := f.tokenizer.Tokenize(input)
    
    // Parallel entropy calculation
    entropies := make([]float64, len(tokens))
    var wg sync.WaitGroup
    for i := range tokens {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            entropies[idx] = calculateEntropy(tokens[idx])
        }(i)
    }
    wg.Wait()
    
    // Filter with better structural detection
    filtered := filterTokens(tokens, entropies, f.threshold)
    output := strings.Join(filtered, " ")
    
    // Cache result
    f.entropyCache.Store(input, output)
    
    return output, estimateTokens(input) - estimateTokens(output)
}
```

---

### Layer 2: Perplexity Pruning

**Research**: LLMLingua (Microsoft/Tsinghua 2023)

**Algorithm**: Iteratively remove tokens with lowest perplexity impact

```go
type PerplexityFilter struct {
    model      *LanguageModel  // Small LM for perplexity
    iterations int             // Default: 3
}

func (f *PerplexityFilter) Apply(input string, mode Mode) (string, int) {
    lines := strings.Split(input, "\n")
    
    // Calculate perplexity for each line
    perplexities := make([]float64, len(lines))
    for i, line := range lines {
        perplexities[i] = f.model.Perplexity(line)
    }
    
    // Iteratively remove low-perplexity lines
    for iter := 0; iter < f.iterations; iter++ {
        // Find line with lowest perplexity
        minIdx := argmin(perplexities)
        
        // Remove line
        lines = append(lines[:minIdx], lines[minIdx+1:]...)
        perplexities = append(perplexities[:minIdx], perplexities[minIdx+1:]...)
        
        // Recalculate perplexities
        for i := range lines {
            perplexities[i] = f.model.Perplexity(lines[i])
        }
    }
    
    output := strings.Join(lines, "\n")
    return output, estimateTokens(input) - estimateTokens(output)
}
```

**Performance**: O(n * k) where n = lines, k = iterations
**Typical reduction**: 15-30%
**Best for**: Code, structured text

**Issues**:
- ❌ Requires external LM (Ollama/LM Studio)
- ❌ Slow (100-500ms per call)
- ❌ Not always available

**Improvements**:
```go
type PerplexityFilter struct {
    model      *LanguageModel
    fallback   *HeuristicPerplexity  // Fast approximation
    timeout    time.Duration
}

func (f *PerplexityFilter) Apply(input string, mode Mode) (string, int) {
    // Try LM with timeout
    ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
    defer cancel()
    
    resultCh := make(chan string, 1)
    go func() {
        output := f.applyWithModel(input, mode)
        resultCh <- output
    }()
    
    select {
    case output := <-resultCh:
        return output, estimateTokens(input) - estimateTokens(output)
    case <-ctx.Done():
        // Fallback to heuristic
        return f.fallback.Apply(input, mode)
    }
}

// Fast heuristic approximation
type HeuristicPerplexity struct{}

func (h *HeuristicPerplexity) Apply(input string, mode Mode) (string, int) {
    lines := strings.Split(input, "\n")
    
    // Score lines by:
    // - Length (shorter = lower perplexity)
    // - Repetition (more repetitive = lower perplexity)
    // - Uniqueness (more unique words = higher perplexity)
    
    scores := make([]float64, len(lines))
    for i, line := range lines {
        scores[i] = h.scoreLine(line)
    }
    
    // Keep top 70% by score
    threshold := percentile(scores, 0.3)
    var filtered []string
    for i, line := range lines {
        if scores[i] >= threshold {
            filtered = append(filtered, line)
        }
    }
    
    output := strings.Join(filtered, "\n")
    return output, estimateTokens(input) - estimateTokens(output)
}
```

---

### Layer 13: H2O Filter (Heavy-Hitter Oracle)

**Research**: NeurIPS 2023

**Algorithm**: Keep only "heavy hitter" tokens with high attention scores

```go
type H2OFilter struct {
    sinkSize        int  // Initial tokens to always keep (default: 4)
    recentSize      int  // Recent tokens to keep (default: 512)
    heavyHitterSize int  // Heavy hitters to keep (default: 256)
}

func (f *H2OFilter) Apply(input string, mode Mode) (string, int) {
    tokens := tokenize(input)
    
    if len(tokens) <= f.sinkSize + f.recentSize {
        return input, 0  // Too short
    }
    
    // Always keep sink tokens (first N)
    sink := tokens[:f.sinkSize]
    
    // Always keep recent tokens (last N)
    recent := tokens[len(tokens)-f.recentSize:]
    
    // Middle tokens: keep heavy hitters
    middle := tokens[f.sinkSize : len(tokens)-f.recentSize]
    
    // Calculate attention scores (approximation)
    scores := make([]float64, len(middle))
    for i, token := range middle {
        scores[i] = f.attentionScore(token, tokens)
    }
    
    // Keep top K heavy hitters
    heavyHitters := topK(middle, scores, f.heavyHitterSize)
    
    // Reconstruct
    result := append(sink, heavyHitters...)
    result = append(result, recent...)
    
    output := detokenize(result)
    return output, estimateTokens(input) - estimateTokens(output)
}

func (f *H2OFilter) attentionScore(token string, context []string) float64 {
    // Approximate attention as:
    // - Frequency in context
    // - Position (earlier = higher)
    // - Uniqueness (rarer = higher)
    
    freq := 0
    for _, t := range context {
        if t == token {
            freq++
        }
    }
    
    uniqueness := 1.0 / float64(freq+1)
    return uniqueness
}
```

**Performance**: O(n log k) where n = tokens, k = heavy hitters
**Typical reduction**: 30-70%
**Best for**: Long context, chat history

**Issues**:
- ❌ Attention score approximation is crude
- ❌ No real attention mechanism
- ❌ May lose important context

**Improvements**:
```go
type H2OFilter struct {
    sinkSize        int
    recentSize      int
    heavyHitterSize int
    attentionModel  *AttentionModel  // Optional real attention
}

func (f *H2OFilter) attentionScore(token string, context []string) float64 {
    // Use real attention if available
    if f.attentionModel != nil {
        return f.attentionModel.Score(token, context)
    }
    
    // Better heuristic:
    // 1. TF-IDF score
    // 2. Position weight (exponential decay)
    // 3. Semantic similarity to query
    
    tf := f.termFrequency(token, context)
    idf := f.inverseDocFrequency(token)
    position := f.positionWeight(token, context)
    
    return tf * idf * position
}

func (f *H2OFilter) termFrequency(token string, context []string) float64 {
    count := 0
    for _, t := range context {
        if t == token {
            count++
        }
    }
    return float64(count) / float64(len(context))
}

func (f *H2OFilter) inverseDocFrequency(token string) float64 {
    // Use pre-computed IDF from corpus
    if idf, ok := f.idfCache[token]; ok {
        return idf
    }
    return 1.0
}

func (f *H2OFilter) positionWeight(token string, context []string) float64 {
    // Exponential decay: earlier tokens = higher weight
    for i, t := range context {
        if t == token {
            return math.Exp(-float64(i) / 100.0)
        }
    }
    return 0.0
}
```

---

### Layer 17: Semantic Cache (Sketch Store)

**Research**: KVReviver-style semantic reuse

**Algorithm**: Cache and reuse compression results for similar inputs

```go
type SketchStoreFilter struct {
    cache       *SemanticCache
    budgetRatio float64  // Fraction of budget for cache lookup
    maxSize     int      // Max cache entries
}

type SemanticCache struct {
    entries map[string]*CacheEntry
    mu      sync.RWMutex
}

type CacheEntry struct {
    input       string
    output      string
    fingerprint uint64
    embedding   []float64  // Semantic embedding
    hits        int
    lastUsed    time.Time
}

func (f *SketchStoreFilter) Apply(input string, mode Mode) (string, int) {
    // Compute fingerprint
    fingerprint := computeFingerprint(input)
    
    // Check exact match
    if entry := f.cache.Get(fingerprint); entry != nil {
        entry.hits++
        entry.lastUsed = time.Now()
        return entry.output, estimateTokens(input) - estimateTokens(entry.output)
    }
    
    // Check semantic similarity
    embedding := computeEmbedding(input)
    if similar := f.cache.FindSimilar(embedding, 0.9); similar != nil {
        // Reuse similar compression
        similar.hits++
        similar.lastUsed = time.Now()
        return similar.output, estimateTokens(input) - estimateTokens(similar.output)
    }
    
    // No cache hit - return unchanged
    return input, 0
}

func (c *SemanticCache) FindSimilar(embedding []float64, threshold float64) *CacheEntry {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    var best *CacheEntry
    bestSim := 0.0
    
    for _, entry := range c.entries {
        sim := cosineSimilarity(embedding, entry.embedding)
        if sim > bestSim && sim >= threshold {
            bestSim = sim
            best = entry
        }
    }
    
    return best
}
```

**Performance**: O(n) where n = cache size
**Typical reduction**: 0-50% (depends on cache hits)
**Best for**: Repetitive commands, similar inputs

**Issues**:
- ❌ Linear search through cache (slow for large caches)
- ❌ Embedding computation is expensive
- ❌ No cache eviction strategy

**Improvements**:
```go
type SketchStoreFilter struct {
    cache       *SemanticCache
    index       *AnnoyIndex  // Approximate nearest neighbor index
    budgetRatio float64
    maxSize     int
}

type SemanticCache struct {
    entries map[uint64]*CacheEntry
    index   *AnnoyIndex  // Fast similarity search
    lru     *LRUCache    // Eviction policy
    mu      sync.RWMutex
}

func (f *SketchStoreFilter) Apply(input string, mode Mode) (string, int) {
    fingerprint := computeFingerprint(input)
    
    // Fast exact match
    if entry := f.cache.Get(fingerprint); entry != nil {
        return entry.output, estimateTokens(input) - estimateTokens(entry.output)
    }
    
    // Fast approximate nearest neighbor search
    embedding := computeEmbedding(input)
    neighbors := f.index.GetNearestNeighbors(embedding, 5, 0.9)
    
    if len(neighbors) > 0 {
        // Use best match
        best := neighbors[0]
        return best.output, estimateTokens(input) - estimateTokens(best.output)
    }
    
    return input, 0
}

func (c *SemanticCache) Set(fingerprint uint64, entry *CacheEntry) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Evict if full
    if len(c.entries) >= c.maxSize {
        oldest := c.lru.RemoveOldest()
        delete(c.entries, oldest)
        c.index.Remove(oldest)
    }
    
    c.entries[fingerprint] = entry
    c.index.Add(fingerprint, entry.embedding)
    c.lru.Add(fingerprint)
}
```

---

## Layer Performance Summary

| Layer | Reduction | Speed | Memory | Best For |
|-------|-----------|-------|--------|----------|
| L1: Entropy | 10-20% | Fast (1ms) | Low (1MB) | Logs, verbose output |
| L2: Perplexity | 15-30% | Slow (100ms) | High (50MB) | Code, structured text |
| L3: Goal-Driven | 20-40% | Medium (10ms) | Medium (10MB) | Query-specific |
| L4: AST | 10-30% | Fast (5ms) | Low (5MB) | Code |
| L13: H2O | 30-70% | Fast (2ms) | Low (2MB) | Long context |
| L17: Semantic Cache | 0-50% | Fast (1ms) | High (100MB) | Repetitive inputs |

## Recommended Layer Combinations

### Fast Preset (< 5ms)
- L1: Entropy
- L4: AST
- L13: H2O
- L17: Semantic Cache

**Total reduction**: 40-60%

### Balanced Preset (< 20ms)
- L1: Entropy
- L2: Perplexity (heuristic)
- L3: Goal-Driven
- L4: AST
- L13: H2O
- L17: Semantic Cache

**Total reduction**: 60-75%

### Full Preset (< 100ms)
- All 20 layers enabled

**Total reduction**: 70-90%
