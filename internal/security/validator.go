// Package security provides security validation utilities for TokMan.
package security

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/errors"
)

// Common security patterns
var (
	// PathTraversalPattern matches path traversal attempts
	PathTraversalPattern = regexp.MustCompile(`\.\./|\.\.\\`)

	// NullBytePattern matches null byte injection attempts
	NullBytePattern = regexp.MustCompile(`\x00`)

	// ControlCharPattern matches control characters (except common whitespace)
	ControlCharPattern = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)

	// ShellMetaChars matches shell metacharacters
	ShellMetaChars = regexp.MustCompile(`[;&|\\$\(\)\{\}\[\]\*\?\` + "`" + `<>]`)
)

// Validator provides security validation utilities
type Validator struct {
	// AllowedPresets defines valid pipeline presets
	AllowedPresets map[string]bool

	// AllowedModes defines valid compression modes
	AllowedModes map[string]bool

	// MaxPathLength is the maximum allowed path length
	MaxPathLength int
}

// NewValidator creates a new security validator with defaults
func NewValidator() *Validator {
	return &Validator{
		AllowedPresets: map[string]bool{
			"fast":     true,
			"balanced": true,
			"full":     true,
			"":         true, // Empty is valid (uses default)
		},
		AllowedModes: map[string]bool{
			"minimal":    true,
			"aggressive": true,
			"":           true, // Empty is valid (uses default)
		},
		MaxPathLength: 4096,
	}
}

// ValidatePreset checks if a pipeline preset is valid
func (v *Validator) ValidatePreset(preset string) error {
	if !v.AllowedPresets[preset] {
		return errors.Wrapf(errors.ErrInvalidPreset, "preset must be one of: fast, balanced, full, got: %s", preset)
	}
	return nil
}

// ValidateMode checks if a compression mode is valid
func (v *Validator) ValidateMode(mode string) error {
	if !v.AllowedModes[mode] {
		return errors.Wrapf(errors.ErrInvalidMode, "mode must be one of: minimal, aggressive, got: %s", mode)
	}
	return nil
}

// ValidatePath checks if a file path is safe (no traversal, null bytes, etc.)
func (v *Validator) ValidatePath(path string) error {
	if len(path) > v.MaxPathLength {
		return errors.Wrapf(errors.ErrInvalidInput, "path exceeds maximum length of %d characters", v.MaxPathLength)
	}

	if NullBytePattern.MatchString(path) {
		return errors.Wrap(errors.ErrInvalidInput, "path contains null bytes")
	}

	if PathTraversalPattern.MatchString(path) {
		return errors.Wrap(errors.ErrInvalidInput, "path contains directory traversal sequences")
	}

	// Check for absolute paths that might be dangerous
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		// Absolute paths are allowed but should be carefully validated
		// In a production environment, you might want to whitelist allowed roots
	}

	return nil
}

// ValidateCommandName checks if a command name is safe
func (v *Validator) ValidateCommandName(name string) error {
	if name == "" {
		return errors.Wrap(errors.ErrInvalidInput, "command name cannot be empty")
	}

	if ControlCharPattern.MatchString(name) {
		return errors.Wrapf(errors.ErrInvalidInput, "command name contains control characters")
	}

	if ShellMetaChars.MatchString(name) {
		return errors.Wrap(errors.ErrShellMetaChars, fmt.Sprintf("command name %q contains shell meta-characters", name))
	}

	return nil
}

// ValidateBudget checks if a token budget is valid
func (v *Validator) ValidateBudget(budget int) error {
	if budget < 0 {
		return errors.Wrapf(errors.ErrInvalidInput, "budget must be non-negative, got: %d", budget)
	}

	// Reasonable upper limit to prevent abuse
	if budget > 10_000_000 {
		return errors.Wrapf(errors.ErrInvalidInput, "budget exceeds maximum of 10,000,000 tokens")
	}

	return nil
}

// SanitizeInput removes potentially dangerous characters from input
func (v *Validator) SanitizeInput(input string) string {
	// Remove control characters except common whitespace
	sanitized := ControlCharPattern.ReplaceAllString(input, "")

	// Remove null bytes
	sanitized = NullBytePattern.ReplaceAllString(sanitized, "")

	return sanitized
}

// SanitizePath cleans and validates a path, returning a safe version
func (v *Validator) SanitizePath(path string) (string, error) {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// Validate the cleaned path
	if err := v.ValidatePath(cleanPath); err != nil {
		return "", err
	}

	return cleanPath, nil
}

// IsSafeFilename checks if a filename is safe (no path separators, no traversal)
func (v *Validator) IsSafeFilename(filename string) bool {
	if filename == "" {
		return false
	}

	// Check for path separators
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return false
	}

	// Check for traversal patterns
	if PathTraversalPattern.MatchString(filename) {
		return false
	}

	// Check for null bytes
	if NullBytePattern.MatchString(filename) {
		return false
	}

	return true
}

// ValidateLayerName checks if a layer name is valid
func (v *Validator) ValidateLayerName(layer string) error {
	validLayers := map[string]bool{
		"entropy":         true,
		"perplexity":      true,
		"goal_driven":     true,
		"ast":             true,
		"contrastive":     true,
		"ngram":           true,
		"evaluator":       true,
		"gist":            true,
		"hierarchical":    true,
		"budget":          true,
		"compaction":      true,
		"attribution":     true,
		"h2o":             true,
		"attention_sink":  true,
		"meta_token":      true,
		"semantic_chunk":  true,
		"sketch_store":    true,
		"lazy_pruner":     true,
		"semantic_anchor": true,
		"agent_memory":    true,
	}

	if !validLayers[layer] {
		return errors.Wrapf(errors.ErrInvalidInput, "invalid layer name: %s", layer)
	}

	return nil
}

// ValidateProfile checks if a compression profile is valid
func (v *Validator) ValidateProfile(profile string) error {
	validProfiles := map[string]bool{
		"surface": true,
		"trim":    true,
		"extract": true,
		"core":    true,
		"code":    true,
		"log":     true,
		"thread":  true,
		"":        true, // Empty is valid (auto-detect)
	}

	if !validProfiles[profile] {
		return errors.Wrapf(errors.ErrInvalidInput, "invalid profile: %s, must be one of: surface, trim, extract, core, code, log, thread", profile)
	}

	return nil
}
