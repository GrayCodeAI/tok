package siem

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type OCSFEvent struct {
	Time         int64                  `json:"time"`
	EventCode    int                    `json:"event_code"`
	EventName    string                 `json:"event_name"`
	Severity     string                 `json:"severity"`
	SeverityID   int                    `json:"severity_id"`
	CategoryName string                 `json:"category_name"`
	CategoryUID  int                    `json:"category_uid"`
	ClassName    string                 `json:"class_name"`
	ClassUID     int                    `json:"class_uid"`
	Message      string                 `json:"message"`
	Status       string                 `json:"status"`
	StatusID     int                    `json:"status_id"`
	UnmappedData map[string]interface{} `json:"unmapped"`
}

type SIEMIntegration struct {
	syslogEndpoint string
	webhookURL     string
	events         []OCSFEvent
}

func NewSIEMIntegration() *SIEMIntegration {
	return &SIEMIntegration{}
}

func (s *SIEMIntegration) Configure(syslogEndpoint, webhookURL string) {
	s.syslogEndpoint = syslogEndpoint
	s.webhookURL = webhookURL
}

func (s *SIEMIntegration) RecordEvent(event OCSFEvent) {
	event.Time = time.Now().UnixMilli()
	s.events = append(s.events, event)
}

func (s *SIEMIntegration) ToOCSF(eventType string, message string, severity int) OCSFEvent {
	severityStr := "Informational"
	if severity >= 7 {
		severityStr = "High"
	} else if severity >= 4 {
		severityStr = "Medium"
	}

	return OCSFEvent{
		EventCode:    1001,
		EventName:    eventType,
		Severity:     severityStr,
		SeverityID:   severity,
		CategoryName: "Security Finding",
		CategoryUID:  2,
		ClassName:    "Finding",
		ClassUID:     2001,
		Message:      message,
		Status:       "New",
		StatusID:     1,
		UnmappedData: make(map[string]interface{}),
	}
}

func (s *SIEMIntegration) ExportJSON() ([]byte, error) {
	return json.MarshalIndent(s.events, "", "  ")
}

func (s *SIEMIntegration) ExportSyslog() string {
	var sb strings.Builder
	for _, e := range s.events {
		sb.WriteString(fmt.Sprintf("<%d>%s tokman: %s\n",
			e.SeverityID*8+1,
			time.UnixMilli(e.Time).Format(time.RFC3339),
			e.Message))
	}
	return sb.String()
}

func (s *SIEMIntegration) ExportWebhook() []map[string]interface{} {
	var payloads []map[string]interface{}
	for _, e := range s.events {
		payloads = append(payloads, map[string]interface{}{
			"event":     e,
			"source":    "tokman",
			"timestamp": e.Time,
		})
	}
	return payloads
}

type AuditLogger struct {
	events []AuditEvent
}

type AuditEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Resource  string    `json:"resource"`
	Outcome   string    `json:"outcome"`
	Details   string    `json:"details,omitempty"`
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

func (l *AuditLogger) Log(action, actor, resource, outcome, details string) {
	l.events = append(l.events, AuditEvent{
		Timestamp: time.Now(),
		Action:    action,
		Actor:     actor,
		Resource:  resource,
		Outcome:   outcome,
		Details:   details,
	})
}

func (l *AuditLogger) GetEvents() []AuditEvent {
	return l.events
}

func (l *AuditLogger) ExportJSON() ([]byte, error) {
	return json.MarshalIndent(l.events, "", "  ")
}

func (l *AuditLogger) Search(action string) []AuditEvent {
	var results []AuditEvent
	for _, e := range l.events {
		if e.Action == action {
			results = append(results, e)
		}
	}
	return results
}
