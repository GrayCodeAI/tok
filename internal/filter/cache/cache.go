// Package cache provides caching support for the filter pipeline.
//
// Includes:
//   - LRU cache for frequently compressed content
//   - Semantic cache for similar inputs
//   - Key-Value cache for KVReviver-style storage
//   - Fingerprint-based result caching
//
// The cache reduces redundant compression work for repeated or similar inputs,
// improving performance by 3-10x on recurring command patterns.
package cache
