package security

type Finding struct {
	Rule     string
	Severity string
	Message  string
}

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Scan(content string) []Finding {
	return nil
}

func RedactPII(content string) string {
	return content
}
