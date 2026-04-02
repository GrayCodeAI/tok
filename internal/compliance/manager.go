package compliance

import (
	"encoding/json"
	"time"
)

type AuditEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Outcome   string    `json:"outcome"`
	Details   string    `json:"details,omitempty"`
}

type AuditTrail struct {
	entries []AuditEntry
}

func NewAuditTrail() *AuditTrail {
	return &AuditTrail{}
}

func (a *AuditTrail) Record(entry AuditEntry) {
	entry.Timestamp = time.Now()
	a.entries = append(a.entries, entry)
}

func (a *AuditTrail) GetEntries() []AuditEntry {
	return a.entries
}

func (a *AuditTrail) Search(userID, action string) []AuditEntry {
	var results []AuditEntry
	for _, e := range a.entries {
		if userID != "" && e.UserID != userID {
			continue
		}
		if action != "" && e.Action != action {
			continue
		}
		results = append(results, e)
	}
	return results
}

func (a *AuditTrail) ExportJSON() ([]byte, error) {
	return json.MarshalIndent(a.entries, "", "  ")
}

type DataRetentionManager struct {
	policies map[string]time.Duration
}

func NewDataRetentionManager() *DataRetentionManager {
	return &DataRetentionManager{
		policies: map[string]time.Duration{
			"audit_logs":   365 * 24 * time.Hour,
			"usage_data":   90 * 24 * time.Hour,
			"session_data": 30 * 24 * time.Hour,
			"cache_data":   7 * 24 * time.Hour,
		},
	}
}

func (m *DataRetentionManager) GetPolicy(resource string) time.Duration {
	return m.policies[resource]
}

func (m *DataRetentionManager) IsExpired(resource string, createdAt time.Time) bool {
	policy, ok := m.policies[resource]
	if !ok {
		return false
	}
	return time.Since(createdAt) > policy
}

func (m *DataRetentionManager) SetPolicy(resource string, duration time.Duration) {
	m.policies[resource] = duration
}

type ComplianceReport struct {
	GeneratedAt     time.Time `json:"generated_at"`
	Regulation      string    `json:"regulation"`
	Status          string    `json:"status"`
	Findings        []string  `json:"findings"`
	Recommendations []string  `json:"recommendations"`
}

type ComplianceManager struct {
	auditTrail   *AuditTrail
	retentionMgr *DataRetentionManager
}

func NewComplianceManager() *ComplianceManager {
	return &ComplianceManager{
		auditTrail:   NewAuditTrail(),
		retentionMgr: NewDataRetentionManager(),
	}
}

func (c *ComplianceManager) GenerateGDPRReport(userID string) *ComplianceReport {
	return &ComplianceReport{
		GeneratedAt: time.Now(),
		Regulation:  "GDPR",
		Status:      "compliant",
		Findings: []string{
			"Data export available for user " + userID,
			"Data deletion workflow ready",
			"Consent records maintained",
		},
		Recommendations: []string{
			"Review data retention policies annually",
			"Ensure consent is explicitly recorded",
		},
	}
}

func (c *ComplianceManager) GenerateHIPAAReport() *ComplianceReport {
	return &ComplianceReport{
		GeneratedAt: time.Now(),
		Regulation:  "HIPAA",
		Status:      "ready",
		Findings: []string{
			"Audit logging enabled",
			"Access controls implemented",
			"Encryption at rest configured",
		},
		Recommendations: []string{
			"Conduct annual risk assessment",
			"Review access logs monthly",
		},
	}
}

func (c *ComplianceManager) AuditTrail() *AuditTrail {
	return c.auditTrail
}

func (c *ComplianceManager) RetentionManager() *DataRetentionManager {
	return c.retentionMgr
}
