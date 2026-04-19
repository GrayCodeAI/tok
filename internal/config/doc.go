// Package config provides configuration management for tok.
//
// The config package handles loading, validating, and saving configuration
// from TOML files, environment variables, and command-line flags.
//
// # Configuration Loading
//
// Use Load() to read configuration from files and environment:
//
//	cfg, err := config.Load(cfgFile)
//	if err != nil { ... }
//
// # Defaults
//
// The Defaults() function returns a configuration with sensible defaults:
//
//	cfg := config.Defaults()
//
// # Validation
//
// All configurations are validated via the Validate() method to ensure
// threshold values, limits, and cross-field constraints are satisfied.
package config
