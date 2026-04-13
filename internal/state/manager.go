package state

import (
	"sync"

	"github.com/GrayCodeAI/tokman/internal/config"
	"github.com/spf13/cobra"
)

// Manager consolidates all global state with single mutex
type Manager struct {
	mu sync.RWMutex
	
	// CLI state
	rootCmd      *cobra.Command
	cfgFile      string
	verbose      int
	dryRun       bool
	ultraCompact bool
	queryIntent  string
	tokenBudget  int
	
	// Config
	config *config.Config
	
	// Runtime state
	version string
}

var (
	global     *Manager
	globalOnce sync.Once
)

// Global returns the global state manager
func Global() *Manager {
	globalOnce.Do(func() {
		global = &Manager{
			version: "dev",
		}
	})
	return global
}

// SetRootCmd sets the root command
func (m *Manager) SetRootCmd(cmd *cobra.Command) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rootCmd = cmd
}

// GetRootCmd gets the root command
func (m *Manager) GetRootCmd() *cobra.Command {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rootCmd
}

// SetConfig sets configuration
func (m *Manager) SetConfig(cfg *config.Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
}

// GetConfig gets configuration
func (m *Manager) GetConfig() *config.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// SetFlags sets all CLI flags atomically
func (m *Manager) SetFlags(verbose int, dryRun, ultraCompact bool, queryIntent string, budget int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.verbose = verbose
	m.dryRun = dryRun
	m.ultraCompact = ultraCompact
	m.queryIntent = queryIntent
	m.tokenBudget = budget
}

// GetFlags returns current flag values
func (m *Manager) GetFlags() (verbose int, dryRun, ultraCompact bool, queryIntent string, budget int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.verbose, m.dryRun, m.ultraCompact, m.queryIntent, m.tokenBudget
}

// IsVerbose checks verbose flag
func (m *Manager) IsVerbose() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.verbose > 0
}

// GetVersion returns version
func (m *Manager) GetVersion() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.version
}

// SetVersion sets version
func (m *Manager) SetVersion(v string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.version = v
}
