package tracking

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	corepricing "github.com/GrayCodeAI/tokman/internal/core"
)

// PricingCoverageSummary describes how well tracked model usage maps to known pricing.
type PricingCoverageSummary struct {
	TotalAttributedCommands int64    `json:"total_attributed_commands"`
	KnownPricingCommands    int64    `json:"known_pricing_commands"`
	FallbackPricingCommands int64    `json:"fallback_pricing_commands"`
	UnknownModels           []string `json:"unknown_models,omitempty"`
}

// CoveragePct returns the share of model-attributed commands with explicit pricing.
func (p PricingCoverageSummary) CoveragePct() float64 {
	if p.TotalAttributedCommands == 0 {
		return 100
	}
	return float64(p.KnownPricingCommands) / float64(p.TotalAttributedCommands) * 100
}

// DashboardDataQuality summarizes attribution and pricing gaps in dashboard data.
type DashboardDataQuality struct {
	WindowDays              int                    `json:"window_days"`
	TotalCommands           int64                  `json:"total_commands"`
	CommandsMissingAgent    int64                  `json:"commands_missing_agent"`
	CommandsMissingProvider int64                  `json:"commands_missing_provider"`
	CommandsMissingModel    int64                  `json:"commands_missing_model"`
	CommandsMissingSession  int64                  `json:"commands_missing_session"`
	ParseFailures           int64                  `json:"parse_failures"`
	PricingCoverage         PricingCoverageSummary `json:"pricing_coverage"`
}

// GetDashboardDataQuality returns attribution/pricing coverage for the dashboard window.
func (t *Tracker) GetDashboardDataQuality(opts DashboardQueryOptions) (DashboardDataQuality, error) {
	days := normalizeDashboardDays(opts.Days)
	where, args := buildDashboardFilters(opts, days)

	query := `
		SELECT
			COUNT(*) AS total_commands,
			COALESCE(SUM(CASE WHEN TRIM(COALESCE(agent_name, '')) = '' THEN 1 ELSE 0 END), 0) AS missing_agent,
			COALESCE(SUM(CASE WHEN TRIM(COALESCE(provider, '')) = '' THEN 1 ELSE 0 END), 0) AS missing_provider,
			COALESCE(SUM(CASE WHEN TRIM(COALESCE(model_name, '')) = '' THEN 1 ELSE 0 END), 0) AS missing_model,
			COALESCE(SUM(CASE WHEN TRIM(COALESCE(session_id, '')) = '' THEN 1 ELSE 0 END), 0) AS missing_session,
			COALESCE(SUM(CASE WHEN parse_success = 0 THEN 1 ELSE 0 END), 0) AS parse_failures
		FROM commands
		WHERE ` + where

	var quality DashboardDataQuality
	quality.WindowDays = days
	if err := t.db.QueryRow(query, args...).Scan(
		&quality.TotalCommands,
		&quality.CommandsMissingAgent,
		&quality.CommandsMissingProvider,
		&quality.CommandsMissingModel,
		&quality.CommandsMissingSession,
		&quality.ParseFailures,
	); err != nil {
		return DashboardDataQuality{}, fmt.Errorf("dashboard data quality query: %w", err)
	}

	pricingCoverage, err := t.GetPricingCoverage(opts)
	if err != nil {
		return DashboardDataQuality{}, err
	}
	quality.PricingCoverage = pricingCoverage

	return quality, nil
}

// GetPricingCoverage returns how many tracked model-attributed commands map to explicit pricing.
func (t *Tracker) GetPricingCoverage(opts DashboardQueryOptions) (PricingCoverageSummary, error) {
	days := normalizeDashboardDays(opts.Days)
	where, args := buildDashboardFilters(opts, days)

	query := `
		SELECT
			TRIM(model_name) AS model_name,
			COUNT(*) AS commands
		FROM commands
		WHERE ` + where + `
			AND TRIM(COALESCE(model_name, '')) <> ''
		GROUP BY TRIM(model_name)
		ORDER BY commands DESC, model_name ASC`

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return PricingCoverageSummary{}, fmt.Errorf("pricing coverage query: %w", err)
	}
	defer rows.Close()

	var summary PricingCoverageSummary
	unknownSet := map[string]struct{}{}
	for rows.Next() {
		var modelName sql.NullString
		var commandCount int64
		if err := rows.Scan(&modelName, &commandCount); err != nil {
			return PricingCoverageSummary{}, fmt.Errorf("pricing coverage scan: %w", err)
		}

		normalized := strings.TrimSpace(modelName.String)
		if normalized == "" {
			continue
		}
		summary.TotalAttributedCommands += commandCount
		if corepricing.HasModelPricing(normalized) || hasTrackingPricing(normalized) {
			summary.KnownPricingCommands += commandCount
			continue
		}
		summary.FallbackPricingCommands += commandCount
		unknownSet[normalized] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return PricingCoverageSummary{}, fmt.Errorf("pricing coverage iteration: %w", err)
	}

	summary.UnknownModels = make([]string, 0, len(unknownSet))
	for model := range unknownSet {
		summary.UnknownModels = append(summary.UnknownModels, model)
	}
	sort.Strings(summary.UnknownModels)
	return summary, nil
}

func hasTrackingPricing(model string) bool {
	_, ok := ModelPricing[strings.ToLower(strings.TrimSpace(model))]
	return ok
}
