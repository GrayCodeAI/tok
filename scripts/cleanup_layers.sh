#!/bin/bash
# Remove consolidated experimental layer files
# These are now merged into L14-L16

cd /Users/lakshmanpatel/Desktop/ProjectAlpha/tok/internal/filter

# L21-L25 → L14 (Edge Cases)
rm -f marginal_info_gain.go      # L21
rm -f near_dedup_filter.go       # L22
rm -f cot_compress.go            # L23
rm -f coding_agent_ctx.go        # L24
rm -f perception_compress.go     # L25

# L26-L30 → L15 (Reasoning)
rm -f lightthinker.go            # L26
rm -f think_switcher.go          # L27
rm -f gmsa.go                    # L28
rm -f carl.go                    # L29
rm -f slim_infer.go              # L30
rm -f ssdp.go                    # L33 (also reasoning)

# L31-L45 → L16 (Advanced/Research)
rm -f difft_adapt.go             # L31
rm -f epic.go                    # L32
rm -f agent_ocr.go               # L34
rm -f s2_mad.go                  # L35
rm -f acon.go                    # L36
rm -f latent_collab.go           # L37
rm -f graph_cot.go               # L38
rm -f role_budget.go             # L39
rm -f swe_adaptive_loop.go       # L40
rm -f agent_ocr_history.go       # L41
rm -f plan_budget.go             # L42
rm -f lightmem.go                # L43
rm -f path_shorten.go            # L44
rm -f json_sampler.go            # L45

# Other experimental/consolidated
rm -f structural_collapse.go
rm -f context_crunch.go
rm -f search_crunch.go
rm -f cross_message_dedup.go
rm -f quantum_lock.go
rm -f claw_*.go
rm -f engram_learner.go
rm -f adaptive_learning.go
rm -f tiered_summary.go
rm -f refactored.go
rm -f smallkv_compensate.go
rm -f crunch_bench.go
rm -f smart_truncate.go
rm -f toon.go
rm -f tdd.go
rm -f reversible.go

# Remove test files for removed layers
rm -f pipeline_bench_test.go
rm -f filters_test.go
rm -f pipeline_runtime_test.go

echo "Cleanup complete"