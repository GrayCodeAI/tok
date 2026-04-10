# TokMan Code Migration Scripts

This directory contains automated migration scripts to fix critical issues.

## Usage

```bash
# Run all migrations
./migrate-all.sh

# Run specific migration
./migrate-race-conditions.sh
./migrate-magic-numbers.sh
./migrate-nil-safety.sh
```

## Migrations

### 1. Race Condition Fix (Critical)
**File:** `migrate-race-conditions.sh`
- Replaces unsafe PipelineStats with SafePipelineStats
- Adds mutex protection to hot paths

### 2. Magic Numbers Fix (High)
**File:** `migrate-magic-numbers.sh`
- Replaces hardcoded numbers with named constants
- Improves code readability

### 3. Nil Safety Fix (Critical)
**File:** `migrate-nil-safety.sh`
- Adds nil checks to all filter methods
- Prevents panics

### 4. Performance Optimization (Medium)
**File:** `migrate-performance.sh`
- Integrates memory pools
- Optimizes hot paths

## Rollback

Each migration creates a `.backup` file. To rollback:

```bash
# Rollback specific migration
./rollback.sh race-conditions

# Rollback all
./rollback.sh all
```
