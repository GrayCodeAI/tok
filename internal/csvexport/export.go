package csvexport

import (
	"strings"
)

type CSVExporter struct {
	headers []string
	rows    [][]string
}

func NewCSVExporter(headers []string) *CSVExporter {
	return &CSVExporter{headers: headers}
}

func (e *CSVExporter) AddRow(row []string) {
	e.rows = append(e.rows, row)
}

func (e *CSVExporter) Export() string {
	var sb strings.Builder
	sb.WriteString(strings.Join(e.headers, ",") + "\n")
	for _, row := range e.rows {
		sb.WriteString(strings.Join(row, ",") + "\n")
	}
	return sb.String()
}

func (e *CSVExporter) RowCount() int {
	return len(e.rows)
}

type WebhookSender struct {
	urls []string
}

func NewWebhookSender() *WebhookSender {
	return &WebhookSender{}
}

func (s *WebhookSender) AddURL(url string) {
	s.urls = append(s.urls, url)
}

func (s *WebhookSender) GetURLs() []string {
	return s.urls
}

func (s *WebhookSender) Notify(payload string) []string {
	var sent []string
	for _, url := range s.urls {
		sent = append(sent, url)
	}
	return sent
}
