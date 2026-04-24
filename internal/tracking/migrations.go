package tracking

import (
	"database/sql"
	"fmt"
)

// SchemaVersion is the number of entries in Migrations. Update this whenever
// a new migration is appended.
const SchemaVersion = 19

// CreateCommandsTable creates the main commands table.
const CreateCommandsTable = `
CREATE TABLE IF NOT EXISTS commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command TEXT NOT NULL,
    original_output TEXT,
    filtered_output TEXT,
    original_tokens INTEGER NOT NULL,
    filtered_tokens INTEGER NOT NULL,
    saved_tokens INTEGER NOT NULL,
    project_path TEXT NOT NULL,
    session_id TEXT,
    exec_time_ms INTEGER,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    parse_success BOOLEAN DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_commands_timestamp ON commands(timestamp);
CREATE INDEX IF NOT EXISTS idx_commands_project ON commands(project_path);
CREATE INDEX IF NOT EXISTS idx_commands_session ON commands(session_id);
CREATE INDEX IF NOT EXISTS idx_commands_command ON commands(command);
`

// AddCompositeIndexes adds composite indexes for common query patterns.
// T181: Composite indexes on (project_path, timestamp) and (command, timestamp).
const AddCompositeIndexes = `
CREATE INDEX IF NOT EXISTS idx_commands_project_ts ON commands(project_path, timestamp);
CREATE INDEX IF NOT EXISTS idx_commands_command_ts ON commands(command, timestamp);
CREATE INDEX IF NOT EXISTS idx_commands_saved ON commands(saved_tokens DESC);
`

// CreateSummaryView creates a view for aggregated statistics.
const CreateSummaryView = `
CREATE VIEW IF NOT EXISTS command_summary AS
SELECT 
    project_path,
    command,
    COUNT(*) as execution_count,
    SUM(saved_tokens) as total_saved,
    SUM(original_tokens) as total_original,
    ROUND(100.0 * SUM(saved_tokens) / NULLIF(SUM(original_tokens), 0), 2) as reduction_percent
FROM commands
GROUP BY project_path, command;
`

// CreateParseFailuresTable creates a table for tracking parse failures.
const CreateParseFailuresTable = `
CREATE TABLE IF NOT EXISTS parse_failures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    raw_command TEXT NOT NULL,
    error_message TEXT NOT NULL,
    fallback_succeeded BOOLEAN DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_parse_failures_timestamp ON parse_failures(timestamp);
`

// CreateLayerStatsTable tracks per-layer savings for detailed analysis.
// T184: Per-layer savings tracking.
const CreateLayerStatsTable = `
CREATE TABLE IF NOT EXISTS layer_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command_id INTEGER NOT NULL,
    layer_name TEXT NOT NULL,
    tokens_saved INTEGER NOT NULL DEFAULT 0,
    duration_us INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (command_id) REFERENCES commands(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_layer_stats_command ON layer_stats(command_id);
CREATE INDEX IF NOT EXISTS idx_layer_stats_name ON layer_stats(layer_name);
`

// AddAgentAttributionColumns adds columns for tracking AI agent context.
// Enables per-model, per-provider, per-agent token savings analysis.
const AddAgentAttributionColumns = `
-- SQLite doesn't support IF NOT EXISTS for ALTER TABLE, so we use a safe approach
-- These are run via safeAddColumn which checks if column exists first
`

// CommandColumnDefs defines optional columns added after the base table exists.
var CommandColumnDefs = []struct {
	Name string
	Type string
}{
	{"agent_name", "TEXT"},
	{"model_name", "TEXT"},
	{"provider", "TEXT"},
	{"model_family", "TEXT"},
	{"context_kind", "TEXT"},
	{"context_mode", "TEXT"},
	{"context_resolved_mode", "TEXT"},
	{"context_target", "TEXT"},
	{"context_related_files", "INTEGER NOT NULL DEFAULT 0"},
	{"context_bundle", "BOOLEAN NOT NULL DEFAULT 0"},
}

// AgentAttributionIndexes defines indexes for agent attribution.
const AgentAttributionIndexes = `
CREATE INDEX IF NOT EXISTS idx_commands_agent ON commands(agent_name);
CREATE INDEX IF NOT EXISTS idx_commands_model ON commands(model_name);
CREATE INDEX IF NOT EXISTS idx_commands_provider ON commands(provider);
CREATE INDEX IF NOT EXISTS idx_commands_context_kind ON commands(context_kind);
CREATE INDEX IF NOT EXISTS idx_commands_context_mode ON commands(context_mode);
CREATE INDEX IF NOT EXISTS idx_commands_context_target ON commands(context_target);
CREATE INDEX IF NOT EXISTS idx_commands_context_bundle ON commands(context_bundle);
`

// CreateAgentSummaryView creates a view for per-agent statistics.
const CreateAgentSummaryView = `
CREATE VIEW IF NOT EXISTS agent_summary AS
SELECT
    agent_name,
    model_name,
    provider,
    project_path,
    COUNT(*) as execution_count,
    SUM(saved_tokens) as total_saved,
    SUM(original_tokens) as total_original,
    ROUND(100.0 * SUM(saved_tokens) / NULLIF(SUM(original_tokens), 0), 2) as reduction_percent
FROM commands
WHERE agent_name IS NOT NULL
GROUP BY agent_name, model_name, provider, project_path;
`

// CreateTeamsTable creates the teams/organizations table for multi-tenant support.
const CreateTeamsTable = `
CREATE TABLE IF NOT EXISTS teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    owner_id INTEGER,
    monthly_token_budget INTEGER DEFAULT 10000000,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_teams_slug ON teams(slug);
CREATE INDEX IF NOT EXISTS idx_teams_created ON teams(created_at);
`

// CreateUsersTable creates the users table for multi-tenant RBAC.
const CreateUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    team_id INTEGER NOT NULL,
    full_name TEXT,
    avatar_url TEXT,
    role TEXT NOT NULL DEFAULT 'viewer',
    is_active BOOLEAN NOT NULL DEFAULT 1,
    last_login DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_team ON users(team_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
`

// CreateFilterMetricsTable tracks per-filter effectiveness metrics.
const CreateFilterMetricsTable = `
CREATE TABLE IF NOT EXISTS filter_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filter_name TEXT NOT NULL,
    team_id INTEGER,
    command_id INTEGER,
    tokens_before INTEGER NOT NULL,
    tokens_after INTEGER NOT NULL,
    tokens_saved INTEGER NOT NULL,
    processing_time_us INTEGER NOT NULL,
    effectiveness_score REAL NOT NULL DEFAULT 0.0,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (command_id) REFERENCES commands(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_filter_metrics_name ON filter_metrics(filter_name);
CREATE INDEX IF NOT EXISTS idx_filter_metrics_team ON filter_metrics(team_id);
CREATE INDEX IF NOT EXISTS idx_filter_metrics_created ON filter_metrics(created_at);
CREATE INDEX IF NOT EXISTS idx_filter_metrics_effective ON filter_metrics(effectiveness_score DESC);
`

// CreateCostAggregationsTable stores pre-aggregated cost data for fast dashboard queries.
const CreateCostAggregationsTable = `
CREATE TABLE IF NOT EXISTS cost_aggregations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
    date_bucket TEXT NOT NULL,
    period TEXT NOT NULL,
    total_commands INTEGER NOT NULL DEFAULT 0,
    total_original_tokens INTEGER NOT NULL DEFAULT 0,
    total_filtered_tokens INTEGER NOT NULL DEFAULT 0,
    total_saved_tokens INTEGER NOT NULL DEFAULT 0,
    estimated_cost_usd REAL NOT NULL DEFAULT 0.0,
    estimated_savings_usd REAL NOT NULL DEFAULT 0.0,
    avg_reduction_percent REAL NOT NULL DEFAULT 0.0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cost_agg_team_period ON cost_aggregations(team_id, date_bucket, period);
CREATE INDEX IF NOT EXISTS idx_cost_agg_date ON cost_aggregations(date_bucket);
`

// CreateAuditLogsTable stores all audit trail events.
const CreateAuditLogsTable = `
CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
    user_id INTEGER,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    old_value TEXT,
    new_value TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_team ON audit_logs(team_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);
`

// CreateTrendAnalysisView provides historical trend data for dashboard.
const CreateTrendAnalysisView = `
CREATE VIEW IF NOT EXISTS trend_analysis AS
SELECT
    ca.team_id,
    ca.date_bucket,
    ca.total_commands,
    ca.total_original_tokens,
    ca.total_filtered_tokens,
    ca.total_saved_tokens,
    ca.estimated_cost_usd,
    ca.estimated_savings_usd,
    ca.avg_reduction_percent,
    ROUND(CAST(ca.total_saved_tokens AS REAL) / NULLIF(ca.total_original_tokens, 0) * 100, 2) as actual_reduction_percent
FROM cost_aggregations ca
ORDER BY ca.team_id, ca.date_bucket DESC;
`

// CreateFilterEffectivenessView ranks filters by average effectiveness.
const CreateFilterEffectivenessView = `
CREATE VIEW IF NOT EXISTS filter_effectiveness AS
SELECT
    filter_name,
    team_id,
    COUNT(*) as usage_count,
    AVG(effectiveness_score) as avg_effectiveness,
    SUM(tokens_saved) as total_saved,
    ROUND(AVG(processing_time_us), 2) as avg_processing_time_us,
    MIN(created_at) as first_used,
    MAX(created_at) as last_used
FROM filter_metrics
GROUP BY filter_name, team_id
ORDER BY avg_effectiveness DESC;
`

// CreateUserSessionsTable creates the user sessions table for authentication.
const CreateUserSessionsTable = `
CREATE TABLE IF NOT EXISTS user_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    team_id TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_team ON user_sessions(team_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at);
`

// CreateConfigVersionsTable creates the config versions table for sync.
const CreateConfigVersionsTable = `
CREATE TABLE IF NOT EXISTS config_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    config_key TEXT NOT NULL,
    content BLOB NOT NULL,
    hash TEXT NOT NULL,
    version INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    device_id TEXT,
    is_active BOOLEAN DEFAULT 1,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_config_versions_team_user ON config_versions(team_id, user_id);
CREATE INDEX IF NOT EXISTS idx_config_versions_key ON config_versions(config_key);
CREATE INDEX IF NOT EXISTS idx_config_versions_active ON config_versions(is_active, version DESC);
`

// CreateDevicesTable creates the devices table for tracking device sync state.
const CreateDevicesTable = `
CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    device_name TEXT NOT NULL,
    platform TEXT,
    last_sync DATETIME,
    enabled BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_devices_team_user ON devices(team_id, user_id);
CREATE INDEX IF NOT EXISTS idx_devices_last_sync ON devices(last_sync DESC);
`

// CreateSyncLogsTable creates the sync logs table for audit trail.
const CreateSyncLogsTable = `
CREATE TABLE IF NOT EXISTS sync_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    device_id TEXT,
    config_key TEXT NOT NULL,
    action TEXT NOT NULL,
    local_version INTEGER,
    remote_version INTEGER,
    conflict_detected BOOLEAN DEFAULT 0,
    merge_strategy TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    details TEXT,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sync_logs_team ON sync_logs(team_id);
CREATE INDEX IF NOT EXISTS idx_sync_logs_user ON sync_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_sync_logs_device ON sync_logs(device_id);
CREATE INDEX IF NOT EXISTS idx_sync_logs_timestamp ON sync_logs(timestamp DESC);
`

// CreateCheckpointEventsTable stores runtime checkpoint trigger events.
const CreateCheckpointEventsTable = `
CREATE TABLE IF NOT EXISTS checkpoint_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command_id INTEGER NOT NULL,
    session_id TEXT,
    trigger TEXT NOT NULL,
    reason TEXT,
    fill_pct REAL NOT NULL DEFAULT 0.0,
    quality_score REAL NOT NULL DEFAULT 0.0,
    cooldown_sec INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (command_id) REFERENCES commands(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_checkpoint_events_created ON checkpoint_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_checkpoint_events_trigger ON checkpoint_events(trigger);
CREATE INDEX IF NOT EXISTS idx_checkpoint_events_session ON checkpoint_events(session_id);
`

// Migrations contains all migration statements in order.
var Migrations = []string{
	CreateCommandsTable,
	CreateSummaryView,
	CreateParseFailuresTable,
	AddCompositeIndexes,
	CreateLayerStatsTable,
	AddAgentAttributionColumns,
	CreateAgentSummaryView,
	CreateTeamsTable,
	CreateUsersTable,
	CreateFilterMetricsTable,
	CreateCostAggregationsTable,
	CreateAuditLogsTable,
	CreateTrendAnalysisView,
	CreateFilterEffectivenessView,
	CreateUserSessionsTable,
	CreateConfigVersionsTable,
	CreateDevicesTable,
	CreateSyncLogsTable,
	CreateCheckpointEventsTable,
}

// RunMigrations applies only the migrations that have not yet been run,
// using PRAGMA user_version as the persistent version counter.
//
// When user_version == 0 (new DB or legacy un-versioned DB), all migrations
// are executed. Because every migration uses IF NOT EXISTS, re-running them
// on an existing schema is safe — any missing tables or indexes are created
// and existing ones are left untouched.
func RunMigrations(db *sql.DB) error {
	var current int
	if err := db.QueryRow("PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("read user_version: %w", err)
	}

	target := len(Migrations)

	// Run only the delta: migrations[current:target].
	// When current==0 this covers all migrations (handles both new and legacy DBs).
	for i := current; i < target; i++ {
		if _, err := db.Exec(Migrations[i]); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		// Advance the version after each successful migration so a mid-run
		// crash leaves the DB in a recoverable state.
		if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1)); err != nil {
			return fmt.Errorf("update user_version to %d: %w", i+1, err)
		}
	}

	return nil
}
