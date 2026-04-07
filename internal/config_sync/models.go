package config_sync

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// ConfigVersion represents a versioned configuration.
type ConfigVersion struct {
	ID        string
	TeamID    string
	UserID    string
	ConfigKey string // e.g., "filters", "settings", "cache_config"
	Content   []byte
	Hash      string // SHA256 hash of content for conflict detection
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeviceID  string // Device that created this version
	IsActive  bool
}

// ComputeHash computes SHA256 hash of content.
func (cv *ConfigVersion) ComputeHash() {
	h := sha256.Sum256(cv.Content)
	cv.Hash = hex.EncodeToString(h[:])
}

// SyncRequest represents a config sync request.
type SyncRequest struct {
	TeamID      string
	UserID      string
	DeviceID    string
	ConfigKey   string
	LocalVersion int
	LocalHash   string
	LocalContent []byte
}

// SyncResponse represents the result of a sync operation.
type SyncResponse struct {
	Status            SyncStatus
	ServerVersion     int
	ServerHash        string
	ServerContent     []byte
	ResolvedContent   []byte
	ResolvedHash      string
	ConflictDetected  bool
	MergeStrategy     string // "local", "remote", "merged"
	Timestamp         time.Time
}

// SyncStatus represents the sync status.
type SyncStatus string

const (
	SyncStatusSuccess       SyncStatus = "success"
	SyncStatusConflict      SyncStatus = "conflict"
	SyncStatusLocalNewer    SyncStatus = "local_newer"
	SyncStatusRemoteNewer   SyncStatus = "remote_newer"
	SyncStatusSynced        SyncStatus = "synced"
	SyncStatusUploadNeeded  SyncStatus = "upload_needed"
	SyncStatusDownloadNeeded SyncStatus = "download_needed"
)

// SyncLog represents a sync event in the audit trail.
type SyncLog struct {
	ID               string
	TeamID           string
	UserID           string
	DeviceID         string
	ConfigKey        string
	Action           string // "upload", "download", "merge", "conflict_resolved"
	LocalVersion     int
	RemoteVersion    int
	ConflictDetected bool
	MergeStrategy    string
	Timestamp        time.Time
	Details          string // JSON details
}

// DeviceInfo represents a device's sync state.
type DeviceInfo struct {
	ID        string
	TeamID    string
	UserID    string
	DeviceName string
	Platform  string // "linux", "macos", "windows", "web"
	LastSync  time.Time
	Enabled   bool
	CreatedAt time.Time
}

// ConfigDiff represents changes between two config versions.
type ConfigDiff struct {
	Added    map[string]interface{}
	Modified map[string]interface{}
	Deleted  []string
}

// MergeStrategy defines how to resolve conflicts.
type MergeStrategy string

const (
	MergeStrategyLocal    MergeStrategy = "local"
	MergeStrategyRemote   MergeStrategy = "remote"
	MergeStrategyManual   MergeStrategy = "manual"
	MergeStrategyThreeWay MergeStrategy = "three_way"
)
