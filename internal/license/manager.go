package license

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// LicenseManager handles licensing and feature gating.
type LicenseManager struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewLicenseManager creates a new license manager.
func NewLicenseManager(db *sql.DB, logger *slog.Logger) *LicenseManager {
	if logger == nil {
		logger = slog.Default()
	}
	return &LicenseManager{
		db:     db,
		logger: logger,
	}
}

// GetLicense retrieves a license by team ID.
func (lm *LicenseManager) GetLicense(teamID string) (*License, error) {
	var license License

	row := lm.db.QueryRow(`
		SELECT id, team_id, tier, status, monthly_quota, api_limit, can_customize, analytics, team_users, sso_enabled, priority_support, created_at, expires_at, renews_at
		FROM licenses
		WHERE team_id = ? AND status = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, teamID, StatusActive)

	err := row.Scan(&license.ID, &license.TeamID, &license.Tier, &license.Status,
		&license.MonthlyQuota, &license.ApiLimit, &license.CanCustomize, &license.Analytics,
		&license.TeamUsers, &license.SsoEnabled, &license.PrioritySupport,
		&license.CreatedAt, &license.ExpiresAt, &license.RenewsAt)

	if err == sql.ErrNoRows {
		// Return free tier license
		features := TierFeatures[TierFree]
		return &License{
			TeamID:         teamID,
			Tier:           TierFree,
			Status:         StatusActive,
			MonthlyQuota:   features.MonthlyTokenQuota,
			ApiLimit:       features.RequestsPerDay,
			CanCustomize:   features.CustomFilters,
			Analytics:      features.Analytics,
			TeamUsers:      features.MaxTeamSize,
			SsoEnabled:     features.SsoEnabled,
			PrioritySupport: features.PrioritySupport,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	// Check if expired
	if license.ExpiresAt.Before(time.Now()) {
		license.Status = StatusExpired
	}

	return &license, nil
}

// CreateLicense creates a new license.
func (lm *LicenseManager) CreateLicense(teamID string, tier Tier, duration time.Duration) (*License, error) {
	features, ok := TierFeatures[tier]
	if !ok {
		return nil, fmt.Errorf("invalid tier: %s", tier)
	}

	license := &License{
		ID:             generateID(),
		TeamID:         teamID,
		Tier:           tier,
		Status:         StatusActive,
		MonthlyQuota:   features.MonthlyTokenQuota,
		ApiLimit:       features.RequestsPerDay,
		CanCustomize:   features.CustomFilters,
		Analytics:      features.Analytics,
		TeamUsers:      features.MaxTeamSize,
		SsoEnabled:     features.SsoEnabled,
		PrioritySupport: features.PrioritySupport,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(duration),
		RenewsAt:       time.Now().Add(duration),
	}

	_, err := lm.db.Exec(`
		INSERT INTO licenses (id, team_id, tier, status, monthly_quota, api_limit, can_customize, analytics, team_users, sso_enabled, priority_support, created_at, expires_at, renews_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, license.ID, license.TeamID, license.Tier, license.Status, license.MonthlyQuota, license.ApiLimit,
		license.CanCustomize, license.Analytics, license.TeamUsers, license.SsoEnabled,
		license.PrioritySupport, license.CreatedAt, license.ExpiresAt, license.RenewsAt)

	if err != nil {
		return nil, err
	}

	lm.logger.Info("license created",
		"team_id", teamID,
		"tier", tier,
		"expires_at", license.ExpiresAt,
	)

	return license, nil
}

// UpgradeLicense upgrades a license to a new tier.
func (lm *LicenseManager) UpgradeLicense(teamID string, newTier Tier) (*License, error) {
	// Revoke old license
	_, _ = lm.db.Exec(`UPDATE licenses SET status = ? WHERE team_id = ? AND status = ?`,
		StatusCanceled, teamID, StatusActive)

	// Create new license (annual)
	return lm.CreateLicense(teamID, newTier, 365*24*time.Hour)
}

// GetQuota retrieves current month's usage quota.
func (lm *LicenseManager) GetQuota(teamID string) (*UsageQuota, error) {
	currentMonth := time.Now().Format("2006-01")

	var quota UsageQuota
	row := lm.db.QueryRow(`
		SELECT team_id, month, tokens_used, tokens_quota, api_calls_used, api_calls_quota, last_updated
		FROM usage_quotas
		WHERE team_id = ? AND month = ?
	`, teamID, currentMonth)

	err := row.Scan(&quota.TeamID, &quota.Month, &quota.TokensUsed, &quota.TokensQuota,
		&quota.ApiCallsUsed, &quota.ApiCallsQuota, &quota.LastUpdated)

	if err == sql.ErrNoRows {
		// Create new quota
		license, err := lm.GetLicense(teamID)
		if err != nil {
			return nil, err
		}

		quota = UsageQuota{
			TeamID:        teamID,
			Month:         currentMonth,
			TokensUsed:    0,
			TokensQuota:   license.MonthlyQuota,
			ApiCallsUsed:  0,
			ApiCallsQuota: license.ApiLimit,
			LastUpdated:   time.Now(),
		}

		_, _ = lm.db.Exec(`
			INSERT INTO usage_quotas (team_id, month, tokens_used, tokens_quota, api_calls_used, api_calls_quota, last_updated)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, quota.TeamID, quota.Month, quota.TokensUsed, quota.TokensQuota,
			quota.ApiCallsUsed, quota.ApiCallsQuota, quota.LastUpdated)

		return &quota, nil
	}

	if err != nil {
		return nil, err
	}

	return &quota, nil
}

// RecordUsage records token usage towards quota.
func (lm *LicenseManager) RecordUsage(teamID string, tokensUsed int) error {
	quota, err := lm.GetQuota(teamID)
	if err != nil {
		return err
	}

	newUsage := quota.TokensUsed + tokensUsed

	_, err = lm.db.Exec(`
		UPDATE usage_quotas
		SET tokens_used = ?, last_updated = NOW()
		WHERE team_id = ? AND month = ?
	`, newUsage, teamID, quota.Month)

	// Alert if exceeding quota
	if newUsage > quota.TokensQuota && quota.TokensQuota > 0 {
		lm.logger.Warn("quota exceeded",
			"team_id", teamID,
			"used", newUsage,
			"quota", quota.TokensQuota,
			"overage_percent", (newUsage-quota.TokensQuota)*100/quota.TokensQuota,
		)
	}

	return err
}

// IsFeatureEnabled checks if a feature is enabled for a team.
func (lm *LicenseManager) IsFeatureEnabled(teamID string, featureName string) (bool, error) {
	license, err := lm.GetLicense(teamID)
	if err != nil {
		return false, err
	}

	// Check explicit feature flags
	var enabled bool
	row := lm.db.QueryRow(`
		SELECT enabled FROM feature_flags
		WHERE team_id = ? AND feature_name = ?
	`, teamID, featureName)

	err = row.Scan(&enabled)
	if err == nil {
		return enabled, nil
	}

	// Fall back to tier features
	switch featureName {
	case "analytics":
		return license.Analytics, nil
	case "custom_filters":
		return license.CanCustomize, nil
	case "api_access":
		return license.Tier != TierFree, nil
	case "ide_plugins":
		return license.Tier == TierPro || license.Tier == TierEnterprise, nil
	case "cloud_sync":
		return license.Tier == TierPro || license.Tier == TierEnterprise, nil
	case "sso":
		return license.SsoEnabled, nil
	case "priority_support":
		return license.PrioritySupport, nil
	default:
		return false, nil
	}
}

// EnforceQuota checks if usage would exceed quota.
func (lm *LicenseManager) EnforceQuota(teamID string, tokensRequested int) error {
	quota, err := lm.GetQuota(teamID)
	if err != nil {
		return err
	}

	if quota.IsUnlimited() {
		return nil
	}

	if quota.TokensUsed+tokensRequested > quota.TokensQuota {
		overage := quota.TokensUsed + tokensRequested - quota.TokensQuota
		return fmt.Errorf("quota exceeded by %d tokens", overage)
	}

	return nil
}

func generateID() string {
	return fmt.Sprintf("lic_%d", time.Now().UnixNano())
}
