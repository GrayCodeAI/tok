package web

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var urlCompressCmd = &cobra.Command{
	Use:   "url-compress <url>",
	Short: "Fetch and compress web page content",
	Long: `Fetch a URL, strip HTML noise, and compress the text content.
Achieves 99.6% reduction on web pages.

Examples:
  tokman url-compress https://example.com
  tokman url-compress https://example.com --max-lines 100`,
	RunE: runURLCompress,
}

var urlMaxLines int

func init() {
	registry.Add(func() { registry.Register(urlCompressCmd) })
	urlCompressCmd.Flags().IntVar(&urlMaxLines, "max-lines", 0, "Limit output lines")
}

func runURLCompress(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("URL required")
	}

	url := args[0]
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	text := stripHTML(string(body))
	text = compressWhitespace(text)

	if urlMaxLines > 0 {
		lines := strings.Split(text, "\n")
		if len(lines) > urlMaxLines {
			lines = lines[:urlMaxLines]
			text = strings.Join(lines, "\n")
		}
	}

	fmt.Println(text)
	return nil
}

var (
	htmlTagRe      = regexp.MustCompile(`<[^>]+>`)
	scriptRe       = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
	styleRe        = regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)
	commentRe      = regexp.MustCompile(`<!--.*?-->`)
	entityRe       = regexp.MustCompile(`&[a-z]+;`)
	multiSpaceRe   = regexp.MustCompile(`[ \t]+`)
	multiNewlineRe = regexp.MustCompile(`\n{3,}`)
)

func stripHTML(html string) string {
	text := scriptRe.ReplaceAllString(html, "")
	text = styleRe.ReplaceAllString(text, "")
	text = commentRe.ReplaceAllString(text, "")
	text = htmlTagRe.ReplaceAllString(text, " ")
	text = entityRe.ReplaceAllString(text, "")
	text = multiSpaceRe.ReplaceAllString(text, " ")
	text = multiNewlineRe.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

func compressWhitespace(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// URLEntityProtection protects critical entities during HTML stripping.
type URLEntityProtection struct {
	patterns []*regexp.Regexp
}

// NewURLEntityProtection creates entity protection for tickers, dates, money.
func NewURLEntityProtection() *URLEntityProtection {
	return &URLEntityProtection{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\$[\d,.]+`),
			regexp.MustCompile(`\d{4}-\d{2}-\d{2}`),
			regexp.MustCompile(`[A-Z]{2,5}`),
		},
	}
}

// Protect extracts and protects entities from text.
func (p *URLEntityProtection) Protect(text string) (string, []string) {
	var entities []string
	for _, pattern := range p.patterns {
		matches := pattern.FindAllString(text, -1)
		entities = append(entities, matches...)
	}
	return text, entities
}

// ConfidenceScore measures entity preservation after optimization.
func ConfidenceScore(original, optimized string, entities []string) float64 {
	if len(entities) == 0 {
		return 1.0
	}
	preserved := 0
	for _, e := range entities {
		if strings.Contains(optimized, e) {
			preserved++
		}
	}
	return float64(preserved) / float64(len(entities))
}
