package siem

import "testing"

func TestSIEMIntegration(t *testing.T) {
	s := NewSIEMIntegration()

	event := s.ToOCSF("threat_detected", "Suspicious activity", 7)
	s.RecordEvent(event)

	export, _ := s.ExportJSON()
	if len(export) == 0 {
		t.Error("Expected JSON export")
	}

	syslog := s.ExportSyslog()
	if syslog == "" {
		t.Error("Expected syslog export")
	}

	webhook := s.ExportWebhook()
	if len(webhook) == 0 {
		t.Error("Expected webhook export")
	}
}

func TestAuditLogger(t *testing.T) {
	l := NewAuditLogger()

	l.Log("compress", "user1", "pipeline", "success", "")
	l.Log("login", "user2", "auth", "failure", "invalid password")

	events := l.GetEvents()
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	searched := l.Search("compress")
	if len(searched) != 1 {
		t.Errorf("Expected 1 compress event, got %d", len(searched))
	}
}
