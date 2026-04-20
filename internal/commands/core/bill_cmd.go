package core

// tok bill — attribute tok's tracked savings against a real provider
// invoice, not an estimator.
//
// `tok gain` already reports token savings, but the $ figure it shows is
// computed from a hardcoded price-per-token estimate. That's fine for
// trend, wrong for claims like "tok saved me $X last month." If a user
// exports their Anthropic or OpenAI usage CSV, this command parses it,
// matches by date range against tok-tracked savings records, and
// reports: actual amount billed by the provider, and the $ tok would
// have avoided at the same rate applied to the saved tokens.
//
// Import format detection is by column headers since neither provider
// ships a machine-discoverable schema version in the CSV itself.

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	out "github.com/lakshmanpatel/tok/internal/output"
)

var (
	billProvider string
	billStart    string
	billEnd      string
)

var billCmd = &cobra.Command{
	Use:   "bill",
	Short: "Attribute tok savings against a real provider invoice",
	Long: `Parse a provider usage CSV and match against tok-tracked savings.

Supported formats (auto-detected by column headers):
  - Anthropic console usage export
  - OpenAI usage export

Usage:
  tok bill import anthropic-usage.csv
  tok bill import openai-usage.csv --provider openai
  tok bill import usage.csv --start 2026-03-01 --end 2026-03-31

Reports:
  - Total amount billed by the provider in the date range
  - Tokens tok filtered / compressed in the same range
  - Dollars tok would have saved at the provider's effective rate
  - Effective savings % against the real bill`,
}

var billImportCmd = &cobra.Command{
	Use:   "import <csv-file>",
	Short: "Import a provider usage CSV and attribute tok savings",
	Args:  cobra.ExactArgs(1),
	RunE:  runBillImport,
}

func init() {
	billImportCmd.Flags().StringVar(&billProvider, "provider", "",
		"force provider format (anthropic|openai). Default: auto-detect")
	billImportCmd.Flags().StringVar(&billStart, "start", "",
		"start date YYYY-MM-DD (default: earliest row in CSV)")
	billImportCmd.Flags().StringVar(&billEnd, "end", "",
		"end date YYYY-MM-DD (default: latest row in CSV)")
	billCmd.AddCommand(billImportCmd)
	registry.Add(func() { registry.Register(billCmd) })
}

// billRow is the normalized shape we convert Anthropic and OpenAI rows
// into after parsing. Keeps the downstream matching logic format-agnostic.
type billRow struct {
	date         time.Time
	inputTokens  int64
	outputTokens int64
	amountUSD    float64
}

func runBillImport(cmd *cobra.Command, args []string) error {
	path := args[0]
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err == io.EOF {
		return fmt.Errorf("csv is empty")
	}
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	provider := billProvider
	if provider == "" {
		provider = detectProvider(header)
		if provider == "" {
			return fmt.Errorf("could not auto-detect provider from columns %q — pass --provider anthropic|openai", header)
		}
	}

	var parse func([]string, []string) (billRow, error)
	switch strings.ToLower(provider) {
	case "anthropic":
		parse = parseAnthropicRow
	case "openai":
		parse = parseOpenAIRow
	default:
		return fmt.Errorf("unknown provider %q", provider)
	}

	var rows []billRow
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read row: %w", err)
		}
		row, err := parse(header, rec)
		if err != nil {
			// Skip malformed individual rows rather than aborting the
			// whole import — CSV exports frequently contain summary
			// rows at the bottom that don't match the main schema.
			continue
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return fmt.Errorf("no usable rows parsed from %s", path)
	}

	start, end := computeDateRange(rows, billStart, billEnd)
	rows = filterByDateRange(rows, start, end)

	if len(rows) == 0 {
		return fmt.Errorf("no rows in date range %s..%s", start, end)
	}

	// Aggregate provider bill.
	var totalBilled float64
	var totalInputTok, totalOutputTok int64
	for _, r := range rows {
		totalBilled += r.amountUSD
		totalInputTok += r.inputTokens
		totalOutputTok += r.outputTokens
	}
	totalBilledTok := totalInputTok + totalOutputTok

	// Effective $/token actually charged — no estimator involved.
	effectiveRate := 0.0
	if totalBilledTok > 0 {
		effectiveRate = totalBilled / float64(totalBilledTok)
	}

	// tok-tracked savings from the local database, same window.
	tokSavedTokens, err := savingsTokensInRange(start, end)
	if err != nil {
		return fmt.Errorf("read tok savings: %w", err)
	}
	tokSavedUSD := float64(tokSavedTokens) * effectiveRate

	effectiveReduction := 0.0
	if totalBilled > 0 {
		effectiveReduction = tokSavedUSD / (totalBilled + tokSavedUSD) * 100
	}

	out.Global().Println("tok bill — provider-invoice attribution")
	out.Global().Println("=======================================")
	out.Global().Printf("  provider:     %s\n", strings.ToLower(provider))
	out.Global().Printf("  window:       %s..%s\n", start, end)
	out.Global().Printf("  billed:       $%.2f across %d tokens\n", totalBilled, totalBilledTok)
	out.Global().Printf("  effective:    $%.8f per token (real, from invoice)\n", effectiveRate)
	out.Global().Println()
	out.Global().Printf("  tok saved:    %d tokens\n", tokSavedTokens)
	out.Global().Printf("  would-cost:   $%.2f at the same effective rate\n", tokSavedUSD)
	out.Global().Printf("  effective %%:  %.1f%% cost reduction vs counterfactual\n", effectiveReduction)
	return nil
}

// detectProvider returns "anthropic" or "openai" based on header columns
// that are unique to each export. Intentionally strict — a match on
// neither returns "" so the caller can ask the user to pick via --provider.
func detectProvider(header []string) string {
	h := strings.ToLower(strings.Join(header, ","))
	switch {
	case strings.Contains(h, "input_tokens") && strings.Contains(h, "model"):
		if strings.Contains(h, "cache_creation") || strings.Contains(h, "api_key_name") {
			return "anthropic"
		}
		return "openai"
	case strings.Contains(h, "prompt_tokens") || strings.Contains(h, "completion_tokens"):
		return "openai"
	}
	return ""
}

func parseAnthropicRow(header, rec []string) (billRow, error) {
	m := headerMap(header, rec)
	date, err := parseDateFlexible(firstNonEmpty(m["date"], m["timestamp"], m["created_at"]))
	if err != nil {
		return billRow{}, err
	}
	in := parseInt64(firstNonEmpty(m["input_tokens"], m["input"]))
	out := parseInt64(firstNonEmpty(m["output_tokens"], m["output"]))
	amount := parseFloat(firstNonEmpty(m["cost_usd"], m["cost"], m["amount"], m["total_cost"]))
	return billRow{date: date, inputTokens: in, outputTokens: out, amountUSD: amount}, nil
}

func parseOpenAIRow(header, rec []string) (billRow, error) {
	m := headerMap(header, rec)
	date, err := parseDateFlexible(firstNonEmpty(m["date"], m["timestamp"], m["created_at"]))
	if err != nil {
		return billRow{}, err
	}
	in := parseInt64(firstNonEmpty(m["prompt_tokens"], m["input_tokens"], m["n_context_tokens_total"]))
	out := parseInt64(firstNonEmpty(m["completion_tokens"], m["output_tokens"], m["n_generated_tokens_total"]))
	amount := parseFloat(firstNonEmpty(m["cost"], m["amount"], m["usd"]))
	return billRow{date: date, inputTokens: in, outputTokens: out, amountUSD: amount}, nil
}

func headerMap(header, row []string) map[string]string {
	m := make(map[string]string, len(header))
	for i, h := range header {
		if i >= len(row) {
			break
		}
		m[strings.ToLower(strings.TrimSpace(h))] = strings.TrimSpace(row[i])
	}
	return m
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseInt64(s string) int64 {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func parseFloat(s string) float64 {
	s = strings.TrimPrefix(strings.ReplaceAll(s, ",", ""), "$")
	n, _ := strconv.ParseFloat(s, 64)
	return n
}

// parseDateFlexible accepts the common CSV date formats both providers
// emit (full timestamp, date-only, RFC3339) without forcing the user to
// convert first.
func parseDateFlexible(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	for _, layout := range []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"01/02/2006",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %q", s)
}

func computeDateRange(rows []billRow, startFlag, endFlag string) (string, string) {
	var minT, maxT time.Time
	for i, r := range rows {
		if i == 0 || r.date.Before(minT) {
			minT = r.date
		}
		if i == 0 || r.date.After(maxT) {
			maxT = r.date
		}
	}
	start := minT.Format("2006-01-02")
	end := maxT.Format("2006-01-02")
	if startFlag != "" {
		start = startFlag
	}
	if endFlag != "" {
		end = endFlag
	}
	return start, end
}

func filterByDateRange(rows []billRow, start, end string) []billRow {
	s, _ := time.Parse("2006-01-02", start)
	e, _ := time.Parse("2006-01-02", end)
	e = e.Add(24 * time.Hour) // include full end day
	var out []billRow
	for _, r := range rows {
		if !r.date.Before(s) && r.date.Before(e) {
			out = append(out, r)
		}
	}
	return out
}

// savingsTokensInRange queries the local tracking store for total tokens
// tok filtered/compressed in the window. Returns 0 (not an error) if the
// store doesn't exist yet — a user may be importing a bill before they've
// used tok, which is a valid "what-if" query.
func savingsTokensInRange(start, end string) (int64, error) {
	tracker, err := shared.OpenTracker()
	if err != nil {
		return 0, nil
	}
	defer tracker.Close()

	// tracking stores one row per filtered command; sum the savings column.
	query := `SELECT COALESCE(SUM(saved_tokens), 0) FROM commands
	          WHERE DATE(timestamp) >= ? AND DATE(timestamp) <= ?`
	row := tracker.QueryRow(query, start, end)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, fmt.Errorf("sum savings: %w", err)
	}
	return total, nil
}
