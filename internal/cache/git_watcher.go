package cache

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// GitWatcher monitors git state for cache invalidation
type GitWatcher struct {
	mu         sync.RWMutex
	repoStates map[string]*RepoState // workingDir -> state
	cache      *QueryCache
}

// RepoState tracks the state of a git repository
type RepoState struct {
	WorkingDir  string
	HeadCommit  string
	DirtyFiles  map[string]string // file -> hash
	LastChecked time.Time
}

// NewGitWatcher creates a new git watcher
func NewGitWatcher(cache *QueryCache) *GitWatcher {
	return &GitWatcher{
		repoStates: make(map[string]*RepoState),
		cache:      cache,
	}
}

// GetFileHashes returns hashes for files in a directory
// Used to generate cache keys and detect changes
func (w *GitWatcher) GetFileHashes(workingDir string, files []string) (map[string]string, error) {
	hashes := make(map[string]string)

	// Get current git commit
	commit, err := w.getHeadCommit(workingDir)
	if err != nil {
		// Not a git repo or no commits, use file content hashes
		return w.hashFiles(workingDir, files)
	}

	hashes[".git/HEAD"] = commit

	// Check which files are dirty
	dirtyFiles, err := w.getDirtyFiles(workingDir)
	if err != nil {
		return hashes, nil // Return just commit hash
	}

	// Add dirty file hashes
	for _, file := range files {
		if _, isDirty := dirtyFiles[file]; isDirty {
			// Hash the dirty file content
			contentHash, err := w.hashFile(filepath.Join(workingDir, file))
			if err == nil {
				hashes[file] = contentHash
			}
		}
		// Clean files are covered by commit hash
	}

	return hashes, nil
}

// InvalidateChanged invalidates cache entries for changed files
func (w *GitWatcher) InvalidateChanged(workingDir string) error {
	state, exists := w.getRepoState(workingDir)
	if !exists {
		// First time seeing this repo
		w.updateRepoState(workingDir)
		return nil
	}

	// Check for changes
	currentCommit, err := w.getHeadCommit(workingDir)
	if err != nil {
		return err
	}

	// Commit changed - invalidate all entries for this repo
	if currentCommit != state.HeadCommit {
		err := w.cache.InvalidateByPrefix(workingDir)
		if err != nil {
			return fmt.Errorf("invalidate by commit change: %w", err)
		}
		w.updateRepoState(workingDir)
		return nil
	}

	// Check for new dirty files
	dirtyFiles, err := w.getDirtyFiles(workingDir)
	if err != nil {
		return err
	}

	// Invalidate entries affected by dirty files
	for file := range dirtyFiles {
		if _, wasDirty := state.DirtyFiles[file]; !wasDirty {
			// File became dirty since last check
			// Invalidate entries that might depend on it
			// This is conservative but safe
			err := w.invalidateAffected(workingDir, file)
			if err != nil {
				return err
			}
		}
	}

	w.updateRepoState(workingDir)
	return nil
}

// getRepoState gets cached repo state
func (w *GitWatcher) getRepoState(workingDir string) (*RepoState, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	state, ok := w.repoStates[workingDir]
	return state, ok
}

// updateRepoState updates the cached repo state
func (w *GitWatcher) updateRepoState(workingDir string) {
	commit, _ := w.getHeadCommit(workingDir)
	dirty, _ := w.getDirtyFiles(workingDir)

	w.mu.Lock()
	defer w.mu.Unlock()

	w.repoStates[workingDir] = &RepoState{
		WorkingDir:  workingDir,
		HeadCommit:  commit,
		DirtyFiles:  dirty,
		LastChecked: time.Now(),
	}
}

// invalidateAffected invalidates cache entries that might be affected by a file change
func (w *GitWatcher) invalidateAffected(workingDir string, changedFile string) error {
	// Conservative: invalidate all entries in the working dir
	// A smarter implementation would track file dependencies
	return w.cache.InvalidateByPrefix(workingDir)
}

// getHeadCommit gets the current HEAD commit hash
func (w *GitWatcher) getHeadCommit(workingDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workingDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getDirtyFiles returns a map of dirty files and their status
func (w *GitWatcher) getDirtyFiles(workingDir string) (map[string]string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = workingDir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	dirty := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 4 {
			continue
		}

		// Parse porcelain output: XY filename or XY orig -> rename
		status := line[:2]
		file := strings.TrimSpace(line[3:])

		// Handle rename case
		if strings.Contains(file, " -> ") {
			parts := strings.Split(file, " -> ")
			if len(parts) == 2 {
				file = parts[1] // Use new name
			}
		}

		dirty[file] = status
	}

	return dirty, scanner.Err()
}

// hashFiles hashes file contents
func (w *GitWatcher) hashFiles(workingDir string, files []string) (map[string]string, error) {
	hashes := make(map[string]string)

	for _, file := range files {
		hash, err := w.hashFile(filepath.Join(workingDir, file))
		if err != nil {
			continue
		}
		hashes[file] = hash
	}

	return hashes, nil
}

// hashFile hashes a single file's content
func (w *GitWatcher) hashFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))[:16], nil // Truncate for storage
}

// IsGitRepo checks if a directory is a git repository
func IsGitRepo(workingDir string) bool {
	gitDir := filepath.Join(workingDir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// GetChangedFiles returns files changed between two commits
func GetChangedFiles(workingDir string, oldCommit, newCommit string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", oldCommit, newCommit)
	cmd.Dir = workingDir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}
