package filtercmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var layersCmd = &cobra.Command{
	Use:   "layers",
	Short: "Show TokMan compression layer architecture",
	Long: `Display information about TokMan's compression layer architecture.

Each layer is based on cutting-edge research from 2023-2026:

Layer 1: Entropy Filtering (Selective Context, Mila 2023)
Layer 2: Perplexity Pruning (LLMLingua, Microsoft 2023)
Layer 3: Goal-Driven Selection (SWE-Pruner, Shanghai Jiao Tong 2025)
Layer 4: AST Preservation (LongCodeZip, NUS 2025)
Layer 5: Contrastive Ranking (LongLLMLingua, Microsoft 2024)
Layer 6: N-gram Abbreviation (CompactPrompt 2025)
Layer 7: Evaluator Heads (EHPC, Tsinghua/Huawei 2025)
Layer 8: Gist Compression (Stanford/Berkeley 2023)
Layer 9: Hierarchical Summary (AutoCompressor, Princeton/MIT 2023)
Layer 10: Budget Enforcement (Industry standard)
Layer 11: Compaction (MemGPT, UC Berkeley 2023)
Layer 12: Attribution Filter (ProCut, LinkedIn 2025)
Layer 13: H2O Filter (Heavy-Hitter Oracle, NeurIPS 2023)
Layer 14: Attention Sink (StreamingLLM 2023)
Layer 15: Meta-Token (arXiv:2506.00307 2025)
Layer 16: Semantic Chunk (ChunkKV-style)
Layer 17: Sketch Store (KVReviver, Dec 2025)
Layer 18: Lazy Pruner (LazyLLM, July 2024)
Layer 19: Semantic Anchor (Attention Gradient Detection)
Layer 20: Agent Memory (Knowledge Graph Extraction)

Use --verbose for detailed algorithm explanations.
`,
	Run: runLayers,
}

func init() {
	registry.Add(func() { registry.Register(layersCmd) })
	layersCmd.Flags().BoolP("verbose", "v", false, "show detailed layer descriptions")
}

func runLayers(cmd *cobra.Command, args []string) {
	verbose, _ := cmd.Flags().GetBool("verbose")

	fmt.Println("╔═══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║            TOKMAN LAYERED COMPRESSION PIPELINE                       ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════╣")

	layers := []struct {
		num         int
		name        string
		research    string
		compression string
		desc        string
	}{
		{1, "Entropy Filtering", "Selective Context (Mila 2023)", "2-3x",
			"Removes low-information tokens based on entropy scores. Tokens that appear frequently with little variation are pruned."},
		{2, "Perplexity Pruning", "LLMLingua (Microsoft 2023)", "20x",
			"Uses iterative perplexity scoring to identify and remove less important tokens while preserving semantic meaning."},
		{3, "Goal-Driven Selection", "SWE-Pruner (Shanghai Jiao Tong 2025)", "14.8x",
			"CRF-style line scoring based on query intent. Prioritizes content relevant to the task (debug, review, deploy)."},
		{4, "AST Preservation", "LongCodeZip (NUS 2025)", "4-8x",
			"Syntax-aware compression that preserves abstract syntax tree structure while removing redundant code."},
		{5, "Contrastive Ranking", "LongLLMLingua (Microsoft 2024)", "4-10x",
			"Question-relevance scoring using n-gram contrastive analysis between query and context."},
		{6, "N-gram Abbreviation", "CompactPrompt (2025)", "2.5x",
			"Lossless compression of repeated n-grams using dictionary-based abbreviation."},
		{7, "Evaluator Heads", "EHPC (Tsinghua/Huawei 2025)", "5-7x",
			"Simulates early-layer attention heads to identify important tokens without full model inference."},
		{8, "Gist Compression", "Stanford/Berkeley 2023", "20x+",
			"Compresses prompts into 'gist tokens' - virtual tokens representing semantic meaning."},
		{9, "Hierarchical Summary", "AutoCompressor (Princeton/MIT 2023)", "Extreme",
			"Recursive summarization that compresses context into hierarchical summary vectors."},
		{10, "Budget Enforcement", "Industry Standard", "Guaranteed",
			"Strict token limit enforcement with intelligent truncation preserving critical content."},
		{11, "Compaction", "MemGPT (UC Berkeley 2023)", "10-50x",
			"LLM-based semantic compression that summarizes conversation history into compact state."},
		{12, "Attribution Filter", "ProCut (LinkedIn 2025)", "78%",
			"Prunes low-attribution tokens using importance scoring with positional and frequency bias."},
		{13, "H2O Filter", "Heavy-Hitter Oracle (NeurIPS 2023)", "30x+",
			"Preserves attention sinks, recent tokens, and heavy hitters for efficient KV cache compression."},
		{14, "Attention Sink", "StreamingLLM (2023)", "Infinite",
			"Enables infinite context by preserving initial tokens as attention sinks with rolling cache."},
		{15, "Meta-Token", "arXiv:2506.00307 (2025)", "27%",
			"LZ77-style lossless compression using sliding window token matching for repeated patterns."},
		{16, "Semantic Chunk", "ChunkKV-style", "Variable",
			"Dynamic boundary detection for context-aware chunking that preserves semantic coherence."},
		{17, "Sketch Store", "KVReviver (Dec 2025)", "90%",
			"Sketch-based reversible compression with on-demand reconstruction for memory efficiency."},
		{18, "Lazy Pruner", "LazyLLM (July 2024)", "2.34x",
			"Budget-aware dynamic pruning that defers token decisions until necessary for speedup."},
		{19, "Semantic Anchor", "Attention Gradient Detection", "Preserve",
			"Identifies and preserves semantic anchors using attention gradient analysis."},
		{20, "Agent Memory", "Knowledge Graph Extraction", "Graph",
			"Extracts knowledge graphs from context for efficient agent memory management."},
	}

	for _, l := range layers {
		fmt.Printf("║ Layer %2d: %-20s %-15s         ║\n", l.num, l.name, "("+l.compression+")")
		fmt.Printf("║   Research: %-56s ║\n", l.research)
		if verbose {
			fmt.Printf("║   %-68s ║\n", wrapText(l.desc, 68))
		}
		fmt.Println("╟───────────────────────────────────────────────────────────────────────╢")
	}

	fmt.Println("║ PIPELINE ORDER:                                                       ║")
	fmt.Println("║   Statistical -> Semantic -> Structural -> LLM -> Budget             ║")
	fmt.Println("║                                                                       ║")
	fmt.Println("║ FEATURES:                                                             ║")
	fmt.Println("║   * Streaming for large inputs (up to 2M tokens)                     ║")
	fmt.Println("║   * Automatic validation and fail-safe mode                          ║")
	fmt.Println("║   * Query-aware compression for specific intents                     ║")
	fmt.Println("║   * Configurable layer thresholds                                     ║")
	fmt.Println("║   * Cache for repeated compressions                                  ║")
	fmt.Println("║   * TOML declarative filter support                                   ║")
	fmt.Println("║   * SIMD-optimized operations                                        ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════╝")

	if !verbose {
		fmt.Println("\nUse --verbose for detailed algorithm explanations")
	}
}

func wrapText(s string, width int) string {
	if len(s) <= width {
		return s
	}

	var b strings.Builder
	words := strings.Fields(s)
	lineLen := 0

	for _, word := range words {
		if lineLen+len(word) > width {
			b.WriteString("║   ")
			b.WriteString(word)
			lineLen = len(word)
		} else {
			if lineLen > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(word)
			lineLen += len(word) + 1
		}
	}

	return b.String()
}
