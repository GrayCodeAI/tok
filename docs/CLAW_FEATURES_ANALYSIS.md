# Claw Compactor Features: Deep Analysis for TokMan Integration

**Date:** April 10, 2026  
**Analysis Scope:** All 14 Claw Compactor stages vs TokMan's 20 layers

---

## Executive Summary

After deep analysis of Claw Compactor's codebase, here's what TokMan can learn:

### Already Implemented in TokMan ✅
- **DiffCrunch** - TokMan has `diff_crunch.go` (simpler version)
- **SearchCrunch** - TokMan has `search_crunch.go` (simpler version)
- **LogCrunch** - TokMan has `log_crunch.go` (simpler version)
- **SemanticDedup** - TokMan has `dedup.go` and `near_dedup_filter.go`
- **AST Compression** - TokMan has `ast_preserve.go` (Layer 4)

### High-Value Missing Features 🎯

| Feature | Impact | Complexity | Recommendation |
|---------|--------|------------|----------------|
| **1. Cross-Message Dedup** | 🔥 CRITICAL | Medium | **IMPLEMENT FIRST** |
| **2. QuantumLock (KV-Cache)** | 🔥 HIGH | Low | **IMPLEMENT SECOND** |
| **3. Photon (Image Compression)** | 🔥 HIGH | Medium | **IMPLEMENT THIRD** |
| **4. Cortex (Centralized Detection)** | 🟡 MEDIUM | Medium | Consider |
| **5. Immutable Architecture** | 🟡 MEDIUM | High | Long-term refactor |

### Low-Value Features ⚠️
- **RLE Stage** - TokMan already has better n-gram compression (Layer 6)
- **TokenOpt** - TokMan handles this in TOML filters
- **Abbrev** - TokMan has better semantic compression (Layer 11)
- **Nexus** - TokMan has multiple ML-based layers (7, 8, 9)
- **StructuralCollapse** - TokMan has `structural_collapse.go`

---

## Feature-by-Feature Analysis

### 1. Cross-Message Deduplication 🔥 CRITICAL

**What it does:**
- Deduplicates content across entire conversation history
- Uses 3-word shingle fingerprinting (Jaccard similarity > 0.8)
- Replaces duplicate messages with compact references

**Why TokMan needs it:**
- Long agent sessions repeat the same context across turns
- Tool results get re-pasted in assistant summaries
- System prompts contain repeated fragments
- **Potential savings: 20-40% in multi-turn conversations**

**Current TokMan status:**
- ❌ No cross-message deduplication
- ✅ Has within-message dedup (`dedup.go`, `near_dedup_filter.go`)
- ✅ Has simhash implementation

**Implementation complexity:** MEDIUM
- Need to add message-level API (currently TokMan processes single strings)
- Can reuse existing simhash/fingerprinting code
- ~300 lines of Go code

**Code location in Claw:**
```python
# scripts/lib/fusion/semantic_dedup.py
def dedup_across_messages(messages: list[dict]) -> tuple[list[dict], dict]:
    # Fingerprints each message
    # Compares with all previous messages
    # Replaces duplicates with references
```

---

### 2. QuantumLock (KV-Cache Alignment) 🔥 HIGH

**What it does:**
- Detects dynamic content in system prompts (dates, UUIDs, JWTs, API keys, timestamps)
- Replaces with stable placeholders
- Appends "dynamic context" block at END of message
- **Result: Stable prefix = cache hit on every request**

**Why TokMan needs it:**
- Anthropic/OpenAI prompt caching keys on first N tokens
- Any dynamic content near the top busts the cache
- **Potential savings: 50-90% cache hit rate improvement**

**Current TokMan status:**
- ❌ No KV-cache alignment
- ✅ Has session tracking (`internal/session/`)
- ✅ Has cache infrastructure (`internal/cache/`)

**Implementation complexity:** LOW
- Simple regex patterns for dynamic content
- Straightforward string replacement
- ~200 lines of Go code

**Patterns to detect:**
```go
// ISO dates: 2026-04-10T23:30:00Z
// JWTs: eyJ...
// API keys: sk-..., pk_live_...
// UUIDs: 550e8400-e29b-41d4-a716-446655440000
// Unix timestamps: 1712778600
// Hex IDs: 32-64 char hex strings
```

**Code location in Claw:**
```python
# scripts/lib/fusion/quantum_lock.py
class QuantumLock(FusionStage):
    order = 3  # Runs BEFORE content detection
    
    def stabilize(content: str) -> str:
        # Extract dynamic fragments
        # Replace with placeholders
        # Append dynamic context at end
```

---

### 3. Photon (Image Compression) 🔥 HIGH

**What it does:**
- Detects base64-encoded images in messages
- Resizes large images (>1MB → 512px, >2MB → 384px)
- Converts PNG screenshots to JPEG
- Sets OpenAI `detail: "low"` to cap vision tokens
- **Supports OpenAI, Anthropic, Google GenAI formats**

**Why TokMan needs it:**
- Vision models are increasingly common
- Images bloat context aggressively (1 image = 1000+ tokens)
- **Potential savings: 40-70% on vision-heavy sessions**

**Current TokMan status:**
- ❌ No image compression
- ❌ No base64 detection
- ✅ Has content detection infrastructure

**Implementation complexity:** MEDIUM
- Need image processing library (Go: `image`, `image/jpeg`, `image/png`)
- Base64 detection is simple (regex)
- Format detection (OpenAI vs Anthropic) requires JSON parsing
- ~400 lines of Go code

**Thresholds:**
```go
// 1MB → resize to 512px wide, JPEG quality 85
// 2MB → resize to 384px wide, JPEG quality 75
// PNG → always convert to JPEG quality 85
```

**Code location in Claw:**
```python
# scripts/lib/fusion/photon.py
class PhotonStage(FusionStage):
    order = 8  # Runs early (images bloat most)
    
    def _optimise_image_data_uri(fmt, b64_payload):
        # Decode base64
        # Resize with Pillow
        # Re-encode as JPEG
```

---


### 4. Cortex (Centralized Content Detection) 🟡 MEDIUM

**What it does:**
- Single stage that detects content type and language
- Runs at order 5 (early in pipeline)
- Detects 6 content types: `code`, `json`, `log`, `diff`, `search`, `text`
- Detects 16 languages: Python, JS, TS, Java, C, C++, C#, Go, Rust, Ruby, PHP, Swift, Kotlin, Scala, Shell, SQL
- All downstream stages read from `ctx.content_type` and `ctx.language`

**Why TokMan might want it:**
- Cleaner architecture (single source of truth)
- Easier to debug (one place to check detection logic)
- Better maintainability

**Current TokMan status:**
- ✅ Has distributed detection (each layer detects independently)
- ✅ Has `content_detect.go` for some detection
- ⚠️ Detection logic scattered across multiple files

**Implementation complexity:** MEDIUM
- Need to refactor existing detection logic
- Need to add language detection (currently implicit)
- Need to thread `content_type` through pipeline
- ~500 lines of Go code + refactoring

**Trade-offs:**
- **Pro:** Cleaner, more maintainable
- **Pro:** Single detection pass (faster)
- **Con:** Requires refactoring existing layers
- **Con:** Less flexible (layers can't override detection)

**Recommendation:** Consider for v2.0 refactor, not urgent

**Code location in Claw:**
```python
# scripts/lib/fusion/cortex.py
class Cortex(FusionStage):
    order = 5
    
    def _detect_type(content: str) -> str:
        # Check for diff headers
        # Check for JSON root token
        # Check for log timestamps
        # Check for code patterns
        # Default to "text"
    
    def _detect_language(content: str) -> str:
        # Keyword density analysis
        # File extension hints
        # Lexical heuristics
```

---

### 5. Immutable Architecture 🟡 MEDIUM

**What it does:**
- All data structures are frozen (Python `@dataclass(frozen=True)`)
- Each stage receives immutable `FusionContext`
- Each stage returns immutable `FusionResult`
- No side effects, no mutation

**Why TokMan might want it:**
- Easier to test (pure functions)
- Easier to debug (no hidden state changes)
- Thread-safe by default
- Better for concurrent processing

**Current TokMan status:**
- ❌ Mutable pipeline (strings mutated in-place)
- ❌ Mutable stats struct
- ✅ Go's value semantics help (but not enforced)

**Implementation complexity:** HIGH
- Requires complete pipeline rewrite
- Need to define immutable context/result structs
- Need to refactor all 20 layers
- ~2000+ lines of code changes

**Trade-offs:**
- **Pro:** Better correctness, testability, concurrency
- **Pro:** Easier to reason about
- **Con:** Higher memory overhead (copying)
- **Con:** Slightly slower (no in-place mutation)
- **Con:** Massive refactoring effort

**Recommendation:** Long-term goal for v3.0, not worth it now

**Code location in Claw:**
```python
# scripts/lib/fusion/base.py
@dataclass(frozen=True)
class FusionContext:
    content: str
    content_type: str
    language: str | None
    metadata: dict  # Frozen copy-on-write

@dataclass(frozen=True)
class FusionResult:
    content: str
    original_tokens: int
    compressed_tokens: int
    markers: list[str]
```

---

## Comparison: TokMan vs Claw Implementations

### LogCrunch Comparison

**Claw Compactor (Python):**
```python
# 300+ lines, sophisticated
- Preserves ERROR/WARN/FATAL always
- Detects stack traces (indentation + frame patterns)
- Normalizes timestamps to relative deltas
- Collapses repeated INFO/DEBUG with counts
- Keeps first + last occurrence of repeated patterns
```

**TokMan (Go):**
```go
// 60 lines, simple
- Preserves errors/warnings
- Normalizes log lines (first 8 words)
- Limits repetitions (2 in minimal, 1 in aggressive)
- No stack trace detection
- No timestamp normalization
```

**Verdict:** Claw's LogCrunch is significantly more sophisticated. TokMan should upgrade.

---

### DiffCrunch Comparison

**Claw Compactor (Python):**
```python
# 150+ lines
- Parses unified diff hunks
- Folds unchanged context lines (configurable window)
- Always preserves changed lines (+/-)
- Preserves 3-line context window around changes
- Emits fold markers with line counts
```

**TokMan (Go):**
```go
// 70 lines
- Simple context line counting
- Keeps first N context lines (3 minimal, 2 aggressive)
- Drops remaining context
- No hunk parsing
```

**Verdict:** Claw's DiffCrunch is more sophisticated. TokMan's is adequate but could be improved.

---

### SearchCrunch Comparison

**Claw Compactor (Python):**
```python
# 200+ lines
- Parses structured search results (title, URL, snippet)
- SimHash deduplication of snippets
- Rank-based cutoff (top N results)
- Stores full results in RewindStore
```

**TokMan (Go):**
```go
// 60 lines
- Simple line-based deduplication
- Normalizes search prefixes (1. 2. 3.)
- Keeps top 60 unique (35 aggressive)
- No structured parsing
```

**Verdict:** Claw's SearchCrunch is more sophisticated. TokMan's is adequate for basic use.

---

### SemanticDedup Comparison

**Claw Compactor (Python):**
```python
# 400+ lines
- 3-word shingle fingerprinting
- Jaccard similarity > 0.8 threshold
- Splits text into blocks (paragraphs + code fences)
- Protects code blocks as atomic units
- Cross-message deduplication
```

**TokMan (Go):**
```go
// dedup.go + near_dedup_filter.go
- SimHash fingerprinting
- Hamming distance threshold
- Line-based deduplication
- No cross-message support
```

**Verdict:** Similar sophistication, but Claw has cross-message support (critical feature).

---

## Implementation Priority Ranking

### Priority 1: MUST IMPLEMENT 🔥

1. **Cross-Message Deduplication**
   - **Impact:** 20-40% savings in multi-turn conversations
   - **Complexity:** Medium (~300 lines)
   - **Dependencies:** None (can reuse existing simhash)
   - **Timeline:** 1-2 days

2. **QuantumLock (KV-Cache Alignment)**
   - **Impact:** 50-90% cache hit rate improvement
   - **Complexity:** Low (~200 lines)
   - **Dependencies:** None
   - **Timeline:** 1 day

3. **Photon (Image Compression)**
   - **Impact:** 40-70% savings on vision sessions
   - **Complexity:** Medium (~400 lines)
   - **Dependencies:** Go image libraries (stdlib)
   - **Timeline:** 2-3 days

**Total effort:** 4-6 days for all three critical features

---

### Priority 2: SHOULD UPGRADE 🟡

4. **Upgrade LogCrunch**
   - Add stack trace detection
   - Add timestamp normalization
   - Add occurrence counts
   - **Effort:** 1 day

5. **Upgrade DiffCrunch**
   - Add hunk parsing
   - Add context window preservation
   - Add fold markers with counts
   - **Effort:** 1 day

6. **Upgrade SearchCrunch**
   - Add structured result parsing
   - Add SimHash snippet deduplication
   - **Effort:** 1 day

**Total effort:** 3 days for upgrades

---

### Priority 3: CONSIDER FOR V2 🔵

7. **Cortex (Centralized Detection)**
   - Cleaner architecture
   - Single source of truth
   - **Effort:** 3-5 days (includes refactoring)

8. **Immutable Architecture**
   - Better testability
   - Thread-safe by default
   - **Effort:** 2-3 weeks (massive refactor)

---


## Implementation Roadmap

### Phase 1: Critical Features (Week 1)

#### Day 1-2: Cross-Message Deduplication
```go
// internal/filter/cross_message_dedup.go

package filter

type CrossMessageDedup struct {
    threshold float64  // Jaccard similarity threshold (0.8)
}

type MessageFingerprint struct {
    Index    int
    Shingles map[string]bool  // 3-word shingles
}

func (d *CrossMessageDedup) DeduplicateMessages(messages []Message) ([]Message, *DedupStats) {
    // 1. Fingerprint each message (3-word shingles)
    // 2. Compare each message with all previous
    // 3. If Jaccard > 0.8, replace with reference
    // 4. Return deduplicated messages + stats
}

func computeShingles(text string, n int) map[string]bool {
    // Tokenize text
    // Create n-grams
    // Return set of shingles
}

func jaccardSimilarity(a, b map[string]bool) float64 {
    // |intersection| / |union|
}
```

**Integration point:** Add to `PipelineCoordinator.ProcessMessages()` (new method)

---

#### Day 3: QuantumLock (KV-Cache Alignment)
```go
// internal/filter/quantum_lock.go

package filter

import "regexp"

type QuantumLockFilter struct {
    patterns []DynamicPattern
}

type DynamicPattern struct {
    Name        string
    Regex       *regexp.Regexp
    Placeholder string
}

var defaultPatterns = []DynamicPattern{
    {"iso_date", regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`), "<DATE>"},
    {"jwt", regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`), "<JWT>"},
    {"api_key", regexp.MustCompile(`\b(?:sk|rk)-[A-Za-z0-9_-]{16,}`), "<API_KEY>"},
    {"uuid", regexp.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`), "<UUID>"},
    {"unix_ts", regexp.MustCompile(`\b1[5-9]\d{8}\b`), "<TIMESTAMP>"},
    {"hex_id", regexp.MustCompile(`\b[0-9a-fA-F]{32,64}\b`), "<HEX_ID>"},
}

func (f *QuantumLockFilter) Apply(input string, mode Mode) (string, int) {
    // 1. Extract all dynamic fragments
    // 2. Replace with placeholders
    // 3. Append dynamic context block at end
    // 4. Return stabilized content
}

func (f *QuantumLockFilter) Stabilize(content string) string {
    fragments := f.extractDynamic(content)
    if len(fragments) == 0 {
        return content
    }
    
    stabilized := content
    for _, frag := range fragments {
        stabilized = strings.ReplaceAll(stabilized, frag.Original, frag.Placeholder)
    }
    
    // Append dynamic context
    stabilized += "\n\n---\n<DYNAMIC_CONTEXT>\n"
    for _, frag := range fragments {
        stabilized += fmt.Sprintf("%s: %s\n", frag.Name, frag.Original)
    }
    stabilized += "</DYNAMIC_CONTEXT>"
    
    return stabilized
}
```

**Integration point:** Add as Layer 0 (runs before all other layers)

---

#### Day 4-6: Photon (Image Compression)
```go
// internal/filter/photon.go

package filter

import (
    "encoding/base64"
    "image"
    "image/jpeg"
    "image/png"
    "regexp"
)

type PhotonFilter struct {
    threshold1MB int  // 1MB threshold
    threshold2MB int  // 2MB threshold
}

var dataURIPattern = regexp.MustCompile(`data:image/([a-zA-Z0-9+.-]+);base64,([A-Za-z0-9+/=\n]+)`)

func (f *PhotonFilter) Apply(input string, mode Mode) (string, int) {
    // 1. Detect base64 images (data URI or JSON)
    // 2. Decode and measure size
    // 3. Resize if > threshold
    // 4. Convert PNG to JPEG
    // 5. Re-encode and replace
}

func (f *PhotonFilter) optimizeImage(format string, b64Data string) (string, string, int, int) {
    // Decode base64
    raw, _ := base64.StdEncoding.DecodeString(b64Data)
    originalSize := len(raw)
    
    // Decode image
    img, _, _ := image.Decode(bytes.NewReader(raw))
    
    // Resize if needed
    if originalSize > f.threshold2MB {
        img = resize(img, 384)  // 384px wide
    } else if originalSize > f.threshold1MB {
        img = resize(img, 512)  // 512px wide
    }
    
    // Convert to JPEG
    var buf bytes.Buffer
    jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
    
    // Re-encode base64
    newB64 := base64.StdEncoding.EncodeToString(buf.Bytes())
    
    return "jpeg", newB64, originalSize, buf.Len()
}

func resize(img image.Image, maxWidth int) image.Image {
    // Use imaging library or stdlib
    // Maintain aspect ratio
}
```

**Integration point:** Add as Layer 0.5 (after QuantumLock, before content detection)

---

### Phase 2: Upgrades (Week 2)

#### Day 7: Upgrade LogCrunch
```go
// internal/filter/log_crunch.go (enhanced)

func (f *LogCrunchFilter) Apply(input string, mode Mode) (string, int) {
    lines := strings.Split(input, "\n")
    classified := classifyLines(lines)
    
    // New: Detect stack traces
    inTrace := false
    traceBuffer := []string{}
    
    // New: Normalize timestamps
    if f.normalizeTimestamps {
        lines = normalizeTimestamps(lines)
    }
    
    // New: Collapse with occurrence counts
    output := compressWithCounts(classified)
    
    return strings.Join(output, "\n"), savedTokens
}

func classifyLines(lines []string) []LineInfo {
    // Classify each line as:
    // - error/warn/info/debug
    // - important content (exception, failure, etc.)
    // - stack trace frame
}

func normalizeTimestamps(lines []string) []string {
    // Replace absolute timestamps with relative deltas
    // [2026-04-10 23:30:00] → [+0.000s]
    // [2026-04-10 23:30:05] → [+5.000s]
}

func compressWithCounts(classified []LineInfo) []string {
    // Collapse repeated INFO/DEBUG lines
    // Keep first + last occurrence
    // Add "[... repeated N times ...]" marker
}
```

---

#### Day 8: Upgrade DiffCrunch
```go
// internal/filter/diff_crunch.go (enhanced)

type DiffHunk struct {
    Header      string
    ContextPre  []string  // Context before changes
    Changes     []string  // +/- lines
    ContextPost []string  // Context after changes
}

func (f *DiffCrunchFilter) Apply(input string, mode Mode) (string, int) {
    hunks := parseUnifiedDiff(input)
    
    compressed := []string{}
    for _, hunk := range hunks {
        compressed = append(compressed, hunk.Header)
        
        // Keep 3-line context window
        contextWindow := 3
        if mode == ModeAggressive {
            contextWindow = 2
        }
        
        // Fold excessive context
        if len(hunk.ContextPre) > contextWindow {
            compressed = append(compressed, hunk.ContextPre[:contextWindow]...)
            compressed = append(compressed, fmt.Sprintf("[... %d context lines folded ...]", len(hunk.ContextPre)-contextWindow))
        } else {
            compressed = append(compressed, hunk.ContextPre...)
        }
        
        // Always keep changes
        compressed = append(compressed, hunk.Changes...)
        
        // Fold post-context
        if len(hunk.ContextPost) > contextWindow {
            compressed = append(compressed, hunk.ContextPost[:contextWindow]...)
            compressed = append(compressed, fmt.Sprintf("[... %d context lines folded ...]", len(hunk.ContextPost)-contextWindow))
        } else {
            compressed = append(compressed, hunk.ContextPost...)
        }
    }
    
    return strings.Join(compressed, "\n"), savedTokens
}

func parseUnifiedDiff(input string) []DiffHunk {
    // Parse @@ hunk headers
    // Separate context from changes
    // Return structured hunks
}
```

---

#### Day 9: Upgrade SearchCrunch
```go
// internal/filter/search_crunch.go (enhanced)

type SearchResult struct {
    Rank    int
    Title   string
    URL     string
    Snippet string
}

func (f *SearchCrunchFilter) Apply(input string, mode Mode) (string, int) {
    results := parseSearchResults(input)
    
    // Deduplicate by snippet similarity (SimHash)
    deduplicated := deduplicateBySnippet(results)
    
    // Rank-based cutoff
    maxResults := 60
    if mode == ModeAggressive {
        maxResults = 35
    }
    
    kept := deduplicated[:min(maxResults, len(deduplicated))]
    
    // Format output
    output := formatSearchResults(kept)
    
    return output, savedTokens
}

func parseSearchResults(input string) []SearchResult {
    // Parse structured search results
    // Extract title, URL, snippet
}

func deduplicateBySnippet(results []SearchResult) []SearchResult {
    // Compute SimHash for each snippet
    // Merge results with Hamming distance < 3
}
```

---

## Testing Strategy

### Unit Tests
```go
// internal/filter/cross_message_dedup_test.go

func TestCrossMessageDedup_DuplicateDetection(t *testing.T) {
    dedup := NewCrossMessageDedup()
    
    messages := []Message{
        {Role: "user", Content: "Fix the login bug in auth.py"},
        {Role: "assistant", Content: "I'll help you fix the login bug..."},
        {Role: "user", Content: "Fix the login bug in auth.py"},  // Duplicate
    }
    
    result, stats := dedup.DeduplicateMessages(messages)
    
    assert.Equal(t, 3, len(result))
    assert.Contains(t, result[2].Content, "[content similar to message 0")
    assert.Equal(t, 1, stats.MessagesDeduped)
}

func TestQuantumLock_DynamicContentStabilization(t *testing.T) {
    filter := NewQuantumLockFilter()
    
    input := "Current time: 2026-04-10T23:30:00Z\nAPI Key: sk-abc123def456"
    output, _ := filter.Apply(input, ModeMinimal)
    
    assert.Contains(t, output, "<DATE>")
    assert.Contains(t, output, "<API_KEY>")
    assert.Contains(t, output, "<DYNAMIC_CONTEXT>")
    assert.Contains(t, output, "iso_date: 2026-04-10T23:30:00Z")
}

func TestPhoton_ImageResize(t *testing.T) {
    filter := NewPhotonFilter()
    
    // Create 2MB test image
    largeImage := createTestImage(2048, 2048)
    b64 := base64.StdEncoding.EncodeToString(largeImage)
    input := fmt.Sprintf("data:image/png;base64,%s", b64)
    
    output, saved := filter.Apply(input, ModeMinimal)
    
    assert.Greater(t, saved, 0)
    assert.Contains(t, output, "image/jpeg")  // Converted to JPEG
}
```

---

## Integration Checklist

### For Each New Feature:

1. **Create filter file** in `internal/filter/`
2. **Add to PipelineCoordinator** in `pipeline_init.go`
3. **Add config flags** in `pipeline_types.go`
4. **Add to presets** in `presets.go`
5. **Add unit tests** in `*_test.go`
6. **Add integration test** in `pipeline_integration_test.go`
7. **Update documentation** in `docs/LAYERS.md`
8. **Add to comparison doc** in `docs/CLAW_COMPACTOR_COMPARISON.md`

---

## Performance Considerations

### Memory Impact
- **Cross-Message Dedup:** O(n) memory for fingerprints (n = message count)
- **QuantumLock:** O(1) memory (regex patterns)
- **Photon:** O(image_size) memory during decode/resize

### CPU Impact
- **Cross-Message Dedup:** O(n²) comparisons (acceptable for <100 messages)
- **QuantumLock:** O(m) regex matches (m = pattern count, ~6)
- **Photon:** O(pixels) for resize (expensive, but rare)

### Optimization Strategies
1. **Lazy evaluation:** Only run Photon if "base64" detected
2. **Early exit:** Skip cross-message dedup if <3 messages
3. **Caching:** Cache fingerprints across pipeline runs
4. **Parallel processing:** Run image compression in goroutines

---

## Success Metrics

### Before Implementation (Baseline)
- Average compression: 60-90% on CLI commands
- Multi-turn session compression: ~70%
- Vision session compression: N/A (not supported)
- Cache hit rate: Unknown

### After Implementation (Target)
- Average compression: 60-90% (maintained)
- Multi-turn session compression: **80-85%** (+10-15%)
- Vision session compression: **50-70%** (new capability)
- Cache hit rate: **70-90%** (new metric)

### Measurement
```bash
# Before
tokman benchmark ./workspace --sessions 10

# After
tokman benchmark ./workspace --sessions 10 --cross-message-dedup
tokman benchmark ./workspace --vision --photon
tokman stats --cache-hit-rate
```

---

## Conclusion

**Recommended Implementation Order:**

1. **Week 1:** Cross-Message Dedup + QuantumLock + Photon (critical features)
2. **Week 2:** Upgrade LogCrunch + DiffCrunch + SearchCrunch (quality improvements)
3. **Future:** Consider Cortex refactor for v2.0

**Expected Impact:**
- **20-40% better compression** on multi-turn conversations
- **50-90% cache hit rate** improvement
- **40-70% compression** on vision sessions (new capability)
- **Total effort:** 2 weeks for all high-value features

**Risk Assessment:** LOW
- All features are additive (no breaking changes)
- Can be feature-flagged for gradual rollout
- Existing tests ensure no regressions

---

**Next Steps:**
1. Review this analysis with team
2. Prioritize features based on user feedback
3. Create GitHub issues for each feature
4. Start with Cross-Message Dedup (highest impact, lowest risk)

