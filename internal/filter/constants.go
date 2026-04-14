package filter

// Pipeline configuration constants.
const (
	// Budget thresholds
	TightBudgetThreshold   = 1000 // Budgets below this need aggressive filtering
	MinimalBudgetThreshold = 100  // Budgets below this are emergency mode
	DefaultBudget          = 0    // No budget limit

	// Content size thresholds
	MinContentLength       = 50     // Below this, entropy calc is unreliable
	SmallContentThreshold  = 200    // Use fast path for content below this
	MediumContentThreshold = 1000   // Switch to normal processing
	LargeContentThreshold  = 10000  // Use aggressive compression
	StreamingThreshold     = 500000 // Process in chunks above this

	// Compression thresholds
	HighCompressionRatio   = 0.95 // 95% compression indicates over-compression
	TargetCompressionRatio = 0.75 // Ideal compression target
	MinCompressionRatio    = 0.10 // Below this, compression isn't worth it

	// Layer thresholds
	EntropyMinLength       = 50  // Min length for entropy filtering
	PerplexityMinLines     = 5   // Min lines for perplexity pruning
	H2OMinTokens           = 50  // Min tokens for H2O filter
	AttentionSinkMinLines  = 3   // Min lines for attention sink
	MetaTokenMinLength     = 500 // Min length for meta-token compression
	SemanticChunkMinLength = 300 // Min length for semantic chunking

	// Early exit configuration
	EarlyExitCheckInterval      = 3 // Check budget every N layers
	EarlyExitAggressiveInterval = 1 // Check every layer for tight budgets

	// Cache configuration
	DefaultCacheSize   = 1000 // Default cache size
	DefaultCacheTTL    = 300  // Default TTL in seconds (5 min)
	CacheEvictionBatch = 100  // Evict this many at a time

	// Token estimation
	TokensPerCharHeuristic = 0.25 // 1 token per 4 chars
	MinTokenEstimate       = 1    // Minimum token count
)
