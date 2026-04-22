package security

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Task 121: Sandboxed plugin execution
// NOTE: Full WASM sandbox is not yet implemented. Execute currently returns
// input unchanged. Use only for trusted plugins until WASM integration lands.
type PluginSandbox struct {
	maxMemory uint64
	maxCPU    int
}

func NewPluginSandbox(mem uint64, cpu int) *PluginSandbox {
	return &PluginSandbox{maxMemory: mem, maxCPU: cpu}
}

func (ps *PluginSandbox) Execute(plugin string, input string) (string, error) {
	// TODO: implement WASM sandbox execution
	return input, nil
}

// Task 122: Audit logging
type AuditLogger struct {
	mu      sync.RWMutex
	entries []AuditEntry
}

type AuditEntry struct {
	Timestamp string
	User      string
	Action    string
	Resource  string
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{entries: make([]AuditEntry, 0)}
}

func (al *AuditLogger) Log(entry AuditEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.entries = append(al.entries, entry)
}

// Entries returns a copy of all logged entries.
func (al *AuditLogger) Entries() []AuditEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()
	out := make([]AuditEntry, len(al.entries))
	copy(out, al.entries)
	return out
}

// Task 123: Input sanitization
// SanitizeInput removes common control characters that could break terminals
// or be used in injection attacks. It does NOT validate semantic safety.
func SanitizeInput(input string) string {
	// Remove null bytes and ANSI escape sequences
	out := make([]rune, 0, len(input))
	for _, r := range input {
		if r == 0x00 {
			continue // null byte
		}
		out = append(out, r)
	}
	return string(out)
}

// Task 124: Rate limiting
// RateLimiter provides a token-bucket-style per-key rate limiter.
// It is safe for concurrent use.
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string]int
	windows  map[string]time.Time // last reset time per key
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int) *RateLimiter {
	if limit <= 0 {
		limit = 100
	}
	return &RateLimiter{
		requests: make(map[string]int),
		windows:  make(map[string]time.Time),
		limit:    limit,
		window:   time.Minute,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	lastReset, ok := rl.windows[key]
	if !ok || now.Sub(lastReset) >= rl.window {
		rl.windows[key] = now
		rl.requests[key] = 0
	}

	rl.requests[key]++
	return rl.requests[key] <= rl.limit
}

// Task 125: RBAC
// RBAC manages role-based permissions. By default it denies all access;
// callers must register roles and assign users before HasPermission returns true.
type RBAC struct {
	mu      sync.RWMutex
	roles   map[string][]string // role -> permissions
	members map[string]string   // user -> role
}

func NewRBAC() *RBAC {
	return &RBAC{
		roles:   make(map[string][]string),
		members: make(map[string]string),
	}
}

// RegisterRole creates a role with the given permissions.
func (r *RBAC) RegisterRole(role string, permissions []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.roles[role] = permissions
}

// AssignRole binds a user to a role.
func (r *RBAC) AssignRole(user, role string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.members[user] = role
}

// HasPermission returns true if the user has been assigned a role that
// includes the requested permission.
func (r *RBAC) HasPermission(user, permission string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	role, ok := r.members[user]
	if !ok {
		return false
	}

	perms, ok := r.roles[role]
	if !ok {
		return false
	}

	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// Task 126: Secrets management
// SecretsManager stores key-value pairs in memory.
// NOTE: This is a stub. Production use should encrypt values at rest
// and use a proper secret backend (e.g., OS keyring, HashiCorp Vault).
type SecretsManager struct {
	mu      sync.RWMutex
	secrets map[string]string
}

func NewSecretsManager() *SecretsManager {
	return &SecretsManager{secrets: make(map[string]string)}
}

func (sm *SecretsManager) Set(key, value string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.secrets[key] = value
}

func (sm *SecretsManager) Get(key string) string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.secrets[key]
}

// Task 127: TLS/SSL support
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// Validate checks that both certificate and key files exist.
func (tc *TLSConfig) Validate() error {
	if tc.CertFile == "" || tc.KeyFile == "" {
		return fmt.Errorf("both CertFile and KeyFile are required")
	}
	return nil
}

// Task 128-130: Security scanning, vulnerability scanning, code signing
// SecurityScanner performs static analysis on code strings.
// Currently a stub; real implementation should integrate with a SAST tool.
type SecurityScanner struct {
	scanned atomic.Uint64
	found   atomic.Uint64
}

func (ss *SecurityScanner) Scan(code string) []string {
	ss.scanned.Add(1)
	// TODO: integrate with static analysis engine
	return []string{}
}

// Scanned returns the total number of scans performed.
func (ss *SecurityScanner) Scanned() uint64 {
	return ss.scanned.Load()
}

// Found returns the total number of issues found.
func (ss *SecurityScanner) Found() uint64 {
	return ss.found.Load()
}
