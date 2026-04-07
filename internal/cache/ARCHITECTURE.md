# Multi-Tier Caching System Architecture

## Overview

The multi-tier caching system provides L1 (in-memory), L2 (disk), and L3 (remote) caching with different eviction policies and performance characteristics.

## Architecture

### Cache Tiers

```
┌─────────────────────────────────────────────────────────┐
│                    CacheManager                         │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │    L1 Cache  │  │    L2 Cache  │  │    L3 Cache  │ │
│  │  (In-Memory) │  │    (Disk)    │  │   (Remote)   │ │
│  │              │  │              │  │              │ │
│  │  - Fastest   │  │  - Persistent│  │  - Distributed│ │
│  │  - LRU       │  │  - LFU       │  │  - FIFO      │ │
│  │  - 100MB     │  │  - 1GB       │  │  - 10GB      │ │
│  │  - <1ms      │  │  - <10ms     │  │  - <100ms    │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Tier Promotion/Demotion

```
Request Flow:
  1. Check L1 (fastest)
  2. If miss, check L2
  3. If miss, check L3
  4. If miss, fetch from source
  5. Store in all tiers

Promotion:
  L3 → L2: On frequent access
  L2 → L1: On very frequent access

Demotion:
  L1 → L2: When L1 full (LRU eviction)
  L2 → L3: When L2 full (LFU eviction)
  L3 → Delete: When L3 full (FIFO eviction)
```

## Components

### CacheManager
- Coordinates all cache tiers
- Handles promotion/demotion
- Collects statistics
- Manages configuration

### L1Cache (In-Memory)
- **Policy:** LRU (Least Recently Used)
- **Size:** 100MB default
- **Speed:** <1ms access
- **Use case:** Hot data, frequent access
- **Eviction:** When full, evict least recently used

### L2Cache (Disk)
- **Policy:** LFU (Least Frequently Used)
- **Size:** 1GB default
- **Speed:** <10ms access
- **Use case:** Warm data, persistent cache
- **Eviction:** When full, evict least frequently used
- **Compression:** Brotli for space efficiency

### L3Cache (Remote/Distributed)
- **Policy:** FIFO (First In First Out)
- **Size:** 10GB default
- **Speed:** <100ms access
- **Use case:** Cold data, shared cache
- **Eviction:** When full, evict oldest entries

## Key Generation

```go
// Hash-based keys for cache entries
key = sha256(command + working_dir + args + timestamp)
```

## Statistics

- Hit rate per tier
- Miss rate per tier
- Average latency per tier
- Promotion/demotion counts
- Eviction counts
- Size per tier

## Configuration

```toml
[cache]
enabled = true

[cache.l1]
enabled = true
max_size = "100MB"
eviction_policy = "lru"
ttl = "5m"

[cache.l2]
enabled = true
max_size = "1GB"
eviction_policy = "lfu"
compression = true
ttl = "1h"

[cache.l3]
enabled = false
max_size = "10GB"
eviction_policy = "fifo"
remote_url = "redis://localhost:6379"
ttl = "24h"
```

## Performance Targets

| Metric | L1 | L2 | L3 |
|--------|-----|-----|-----|
| Read latency | <1ms | <10ms | <100ms |
| Write latency | <1ms | <10ms | <100ms |
| Hit rate target | 80% | 15% | 4% |
| Miss rate | 1% | 5% | 95% |

## Use Cases

1. **Command Results:** Cache expensive command outputs
2. **File Contents:** Cache frequently accessed files
3. **Filtered Output:** Cache compressed/filtered results
4. **Archive Lookups:** Cache archive metadata
5. **Session Data:** Cache session snapshots
