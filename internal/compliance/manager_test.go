package compliance

import "testing"

func TestAuditTrail(t *testing.T) {
	at := NewAuditTrail()

	at.Record(AuditEntry{
		UserID:   "user1",
		Action:   "login",
		Resource: "auth",
		Outcome:  "success",
	})

	entries := at.GetEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	searched := at.Search("user1", "")
	if len(searched) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(searched))
	}
}

func TestComplianceManager(t *testing.T) {
	m := NewComplianceManager()

	gdpr := m.GenerateGDPRReport("user1")
	if gdpr.Status != "compliant" {
		t.Errorf("Expected compliant, got %s", gdpr.Status)
	}

	hipaa := m.GenerateHIPAAReport()
	if hipaa.Status != "ready" {
		t.Errorf("Expected ready, got %s", hipaa.Status)
	}
}
