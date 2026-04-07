// Package session_recovery provides crash recovery for interrupted sessions.
//
// When a user's terminal session is interrupted (power loss, crash, network disconnect),
// SessionRecovery stores all state needed to resume exactly where the user left off.
//
// Inspired by OMNI's transcript & recovery architecture.
package session_recovery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// SessionState represents a checkpointable session state.
type SessionState struct {
	ID         string              `json:"id"`
	StartedAt  time.Time           `json:"started_at"`
	LastUpdate time.Time           `json:"last_update"`
	CWD        string              `json:"cwd"`
	Env        map[string]string   `json:"env,omitempty"`
	Commands   []CommandEntry      `json:"commands"`
	HotFiles   map[string]FileStat `json:"hot_files"`
	Metadata   map[string]string   `json:"metadata"`
	Checkpoint int                 `json:"checkpoint"`
}

// CommandEntry stores a single command that was executed.
type CommandEntry struct {
	Command     string    `json:"command"`
	Output      string    `json:"output"`
	Error       string    `json:"error,omitempty"`
	TokensIn    int       `json:"tokens_in"`
	TokensOut   int       `json:"tokens_out"`
	SavedTokens int       `json:"saved_tokens"`
	Duration    string    `json:"duration"`
	Hash        string    `json:"hash"`
	Timestamp   time.Time `json:"timestamp"`
	ExitCode    int       `json:"exit_code"`
}

// FileStat tracks hot file metadata.
type FileStat struct {
	Path       string    `json:"path"`
	Modified   time.Time `json:"modified"`
	Size       int64     `json:"size"`
	AccessCount int      `json:"access_count"`
}

// RecoveryInfo contains information needed to resume a session.
type RecoveryInfo struct {
	Session              *SessionState `json:"session"`
	InterruptedCommands []CommandEntry `json:"interrupted_commands"`
	Summary              string        `json:"summary"`
	CanResume            bool          `json:"can_resume"`
}

// RecoveryStore manages session state persistence.
type RecoveryStore struct {
	baseDir string
	mu      sync.RWMutex
}

// Config holds recovery store configuration.
type Config struct {
	BaseDir       string
	MaxSessions   int
	CheckpointInterval time.Duration
	Enabled       bool
}

// DefaultConfig returns default recovery configuration.
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	return Config{
		BaseDir:            filepath.Join(homeDir, ".local", "share", "tokman", "sessions"),
		MaxSessions:        10,
		CheckpointInterval: 5 * time.Second,
		Enabled:            true,
	}
}

// New creates a new RecoveryStore.
func New(cfg Config) (*RecoveryStore, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	if err := os.MkdirAll(cfg.BaseDir, 0700); err != nil {
		return nil, fmt.Errorf("create sessions directory: %w", err)
	}

	return &RecoveryStore{baseDir: cfg.BaseDir}, nil
}

// GenerateID creates a unique session ID.
func GenerateID() string {
	h := sha256.New()
	h.Write([]byte(time.Now().String()))
	return hex.EncodeToString(h.Sum(nil))[:8]
}

// BeginSession starts tracking a new session.
func (s *RecoveryStore) BeginSession(cwd string, env []string) (*SessionState, error) {
	state := &SessionState{
		ID:         GenerateID(),
		StartedAt:  time.Now(),
		LastUpdate: time.Now(),
		CWD:        cwd,
		HotFiles:   make(map[string]FileStat),
		Env:        make(map[string]string),
		Metadata:   make(map[string]string),
	}

	// Parse env vars
	for _, e := range env {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				state.Env[e[:i]] = e[i+1:]
				break
			}
		}
	}

	if err := s.save(state); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	return state, nil
}

// RecordCommand adds a command to the active session.
func (s *RecoveryStore) RecordCommand(sessionID string, cmd CommandEntry) error {
	state, err := s.load(sessionID)
	if err != nil {
		return err
	}

	state.Commands = append(state.Commands, cmd)
	state.LastUpdate = time.Now()

	// Track hot files
	for _, cmdStr := range []string{cmd.Command} {
		for _, word := range splitWords(cmdStr) {
			if info, err := os.Stat(word); err == nil && !info.IsDir() {
				state.HotFiles[word] = FileStat{
					Path:       word,
					Modified:   info.ModTime(),
					Size:       info.Size(),
					AccessCount: state.HotFiles[word].AccessCount + 1,
				}
			}
		}
	}

	return s.save(state)
}

// CreateCheckpoint creates a checkpoint for the session.
func (s *RecoveryStore) CreateCheckpoint(sessionID string) error {
	state, err := s.load(sessionID)
	if err != nil {
		return err
	}

	state.Checkpoint = len(state.Commands)
	state.LastUpdate = time.Now()

	return s.save(state)
}

// CheckRecovery checks if a recoverable session exists.
func (s *RecoveryStore) CheckRecovery() (*RecoveryInfo, error) {
	sessions, err := s.listSessions()
	if err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return &RecoveryInfo{CanResume: false}, nil
	}

	// Get most recent
	recent := sessions[0]

	// Check if session was interrupted (no explicit close)
	interrupted := recent.Commands
	lastCmds := []CommandEntry{}
	if len(interrupted) > 5 {
		lastCmds = interrupted[len(interrupted)-5:]
	} else {
		lastCmds = interrupted
	}

	// Generate summary
	var summary string
	if len(interrupted) > 0 {
		summary = fmt.Sprintf(
			"Previous session had %d commands, %d hot files. Last command: %s",
			len(interrupted),
			len(recent.HotFiles),
			interrupted[len(interrupted)-1].Command,
		)
	} else {
		summary = "Previous session was started but had no commands."
	}

	return &RecoveryInfo{
		Session:              recent,
		InterruptedCommands:  lastCmds,
		Summary:              summary,
		CanResume:            true,
	}, nil
}

// CloseSession marks a session as complete.
func (s *RecoveryStore) CloseSession(sessionID string) error {
	path := filepath.Join(s.baseDir, sessionID+".json")
	
	// Archive rather than delete
	archiveDir := filepath.Join(s.baseDir, "archive")
	os.MkdirAll(archiveDir, 0700)
	
	archivePath := filepath.Join(archiveDir, sessionID+".json")
	if _, err := os.Stat(path); err == nil {
		return os.Rename(path, archivePath)
	}
	return nil
}

// ListSessions returns all stored sessions.
func (s *RecoveryStore) ListSessions() ([]SessionState, error) {
	sessions, err := s.listSessions()
	if err != nil {
		return nil, err
	}

	result := make([]SessionState, len(sessions))
	for i, sess := range sessions {
		result[i] = *sess
	}
	return result, nil
}

func (s *RecoveryStore) listSessions() ([]*SessionState, error) {
	files, err := filepath.Glob(filepath.Join(s.baseDir, "*.json"))
	if err != nil {
		return nil, err
	}

	var sessions []*SessionState
	for _, f := range files {
		if filepath.Base(f) == "archive" {
			continue
		}
		state, err := s.loadFile(f)
		if err != nil {
			continue
		}
		sessions = append(sessions, state)
	}

	// Sort by last update (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastUpdate.After(sessions[j].LastUpdate)
	})

	return sessions, nil
}

func (s *RecoveryStore) save(state *SessionState) error {
	state.LastUpdate = time.Now()

	path := filepath.Join(s.baseDir, state.ID+".json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	// Atomic write
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("write temp: %w", err)
	}

	return os.Rename(tmpPath, path)
}

func (s *RecoveryStore) load(sessionID string) (*SessionState, error) {
	path := filepath.Join(s.baseDir, sessionID+".json")
	return s.loadFile(path)
}

func (s *RecoveryStore) loadFile(path string) (*SessionState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read session file: %w", err)
	}

	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &state, nil
}

// PruneOldSessions removes sessions older than maxSessions.
func (s *RecoveryStore) PruneOldSessions(maxSessions int) (int, error) {
	sessions, err := s.listSessions()
	if err != nil {
		return 0, err
	}

	if len(sessions) <= maxSessions {
		return 0, nil
	}

	removed := 0
	for _, sess := range sessions[maxSessions:] {
		path := filepath.Join(s.baseDir, sess.ID+".json")
		if err := os.Remove(path); err != nil {
			continue
		}
		removed++
	}

	return removed, nil
}

func splitWords(s string) []string {
	var words []string
	var current string
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
