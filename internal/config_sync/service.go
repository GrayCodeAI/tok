package config_sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"
)

var (
	ErrNotFound       = errors.New("config version not found")
	ErrConflict       = errors.New("config conflict detected")
	ErrInvalidInput   = errors.New("invalid input")
)

// ConfigSyncService handles configuration synchronization across devices.
type ConfigSyncService struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewConfigSyncService creates a new config sync service.
func NewConfigSyncService(db *sql.DB, logger *slog.Logger) *ConfigSyncService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ConfigSyncService{
		db:     db,
		logger: logger,
	}
}

// Sync synchronizes a configuration between client and server.
// Returns the server's latest version and detects conflicts.
func (s *ConfigSyncService) Sync(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	if req.TeamID == "" || req.UserID == "" || req.ConfigKey == "" {
		return nil, ErrInvalidInput
	}

	// Get latest server version
	serverVersion, err := s.getLatestVersion(ctx, req.TeamID, req.UserID, req.ConfigKey)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Check for conflicts
	resp := &SyncResponse{
		Timestamp: time.Now(),
	}

	if serverVersion == nil {
		// No server version yet, upload local version
		if err := s.saveVersion(ctx, &ConfigVersion{
			TeamID:    req.TeamID,
			UserID:    req.UserID,
			ConfigKey: req.ConfigKey,
			Content:   req.LocalContent,
			DeviceID:  req.DeviceID,
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
		}); err != nil {
			return nil, err
		}

		resp.Status = SyncStatusSuccess
		resp.ServerVersion = 1
		resp.ServerContent = req.LocalContent
		resp.ServerContent = req.LocalContent
		return resp, nil
	}

	// Compare versions
	if req.LocalVersion == serverVersion.Version && req.LocalHash == serverVersion.Hash {
		// Client and server are in sync
		resp.Status = SyncStatusSynced
		resp.ServerVersion = serverVersion.Version
		resp.ServerHash = serverVersion.Hash
		resp.ServerContent = serverVersion.Content
		return resp, nil
	}

	if req.LocalVersion < serverVersion.Version {
		// Server has newer version
		if req.LocalHash != serverVersion.Hash {
			resp.ConflictDetected = true
			resp.Status = SyncStatusConflict
			resp.ServerVersion = serverVersion.Version
			resp.ServerHash = serverVersion.Hash
			resp.ServerContent = serverVersion.Content
			return resp, nil
		}
		// Server is newer but client changes aren't newer
		resp.Status = SyncStatusDownloadNeeded
		resp.ServerVersion = serverVersion.Version
		resp.ServerContent = serverVersion.Content
		return resp, nil
	}

	// Local version is newer, upload it
	newVersion := &ConfigVersion{
		TeamID:    req.TeamID,
		UserID:    req.UserID,
		ConfigKey: req.ConfigKey,
		Content:   req.LocalContent,
		DeviceID:  req.DeviceID,
		Version:   serverVersion.Version + 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}
	newVersion.ComputeHash()

	if err := s.saveVersion(ctx, newVersion); err != nil {
		return nil, err
	}

	resp.Status = SyncStatusSuccess
	resp.ServerVersion = newVersion.Version
	resp.ServerHash = newVersion.Hash
	resp.ServerContent = newVersion.Content
	resp.ResolvedContent = newVersion.Content
	resp.ResolvedHash = newVersion.Hash
	resp.MergeStrategy = "local"

	// Log the sync
	_ = s.logSync(ctx, &SyncLog{
		TeamID:       req.TeamID,
		UserID:       req.UserID,
		DeviceID:     req.DeviceID,
		ConfigKey:    req.ConfigKey,
		Action:       "upload",
		LocalVersion: req.LocalVersion,
		RemoteVersion: serverVersion.Version,
		Timestamp:    time.Now(),
	})

	return resp, nil
}

// GetLatestVersion retrieves the latest version of a configuration.
func (s *ConfigSyncService) GetLatestVersion(ctx context.Context, teamID, userID, configKey string) (*ConfigVersion, error) {
	return s.getLatestVersion(ctx, teamID, userID, configKey)
}

// GetVersionHistory retrieves the version history of a configuration.
func (s *ConfigSyncService) GetVersionHistory(ctx context.Context, teamID, userID, configKey string, limit int) ([]*ConfigVersion, error) {
	if limit > 100 {
		limit = 100 // Cap limit
	}
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, team_id, user_id, config_key, content, hash, version, created_at, updated_at, device_id, is_active
		FROM config_versions
		WHERE team_id = ? AND user_id = ? AND config_key = ?
		ORDER BY version DESC
		LIMIT ?
	`, teamID, userID, configKey, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*ConfigVersion
	for rows.Next() {
		var v ConfigVersion
		if err := rows.Scan(&v.ID, &v.TeamID, &v.UserID, &v.ConfigKey, &v.Content, &v.Hash,
			&v.Version, &v.CreatedAt, &v.UpdatedAt, &v.DeviceID, &v.IsActive); err != nil {
			return nil, err
		}
		versions = append(versions, &v)
	}

	return versions, rows.Err()
}

// RegisterDevice registers a new device for sync.
func (s *ConfigSyncService) RegisterDevice(ctx context.Context, info *DeviceInfo) error {
	info.CreatedAt = time.Now()
	info.LastSync = time.Now()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO devices (id, team_id, user_id, device_name, platform, last_sync, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, info.ID, info.TeamID, info.UserID, info.DeviceName, info.Platform, info.LastSync, info.Enabled, info.CreatedAt)

	return err
}

// UpdateLastSync updates the last sync time for a device.
func (s *ConfigSyncService) UpdateLastSync(ctx context.Context, deviceID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE devices SET last_sync = NOW() WHERE id = ?
	`, deviceID)
	return err
}

// getLatestVersion gets the latest version of a config (internal).
func (s *ConfigSyncService) getLatestVersion(ctx context.Context, teamID, userID, configKey string) (*ConfigVersion, error) {
	var v ConfigVersion

	row := s.db.QueryRowContext(ctx, `
		SELECT id, team_id, user_id, config_key, content, hash, version, created_at, updated_at, device_id, is_active
		FROM config_versions
		WHERE team_id = ? AND user_id = ? AND config_key = ? AND is_active = 1
		ORDER BY version DESC
		LIMIT 1
	`, teamID, userID, configKey)

	err := row.Scan(&v.ID, &v.TeamID, &v.UserID, &v.ConfigKey, &v.Content, &v.Hash,
		&v.Version, &v.CreatedAt, &v.UpdatedAt, &v.DeviceID, &v.IsActive)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// saveVersion saves a new config version.
func (s *ConfigSyncService) saveVersion(ctx context.Context, v *ConfigVersion) error {
	v.ComputeHash()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO config_versions (team_id, user_id, config_key, content, hash, version, created_at, updated_at, device_id, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, v.TeamID, v.UserID, v.ConfigKey, v.Content, v.Hash, v.Version,
		v.CreatedAt, v.UpdatedAt, v.DeviceID, v.IsActive)

	return err
}

// logSync logs a sync event.
func (s *ConfigSyncService) logSync(ctx context.Context, log *SyncLog) error {
	details, _ := json.Marshal(log)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sync_logs (team_id, user_id, device_id, config_key, action, local_version, remote_version, conflict_detected, merge_strategy, timestamp, details)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, log.TeamID, log.UserID, log.DeviceID, log.ConfigKey, log.Action,
		log.LocalVersion, log.RemoteVersion, log.ConflictDetected, log.MergeStrategy,
		log.Timestamp, string(details))

	return err
}
