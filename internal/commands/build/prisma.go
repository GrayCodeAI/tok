package build

import (
	"fmt"
	"os/exec"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

var prismaCmd = &cobra.Command{
	Use:   "prisma [args...]",
	Short: "Prisma commands with compact output",
	Long: `Execute Prisma commands with token-optimized output.

Specialized filters for:
  - prisma generate: Compact schema generation
  - prisma migrate dev/status/reset: Compact migration output
  - prisma db push/pull: Compact database sync
  - prisma studio: Note studio URL
  - prisma validate: Compact validation

Examples:
  tok prisma generate
  tok prisma migrate dev --name init
  tok prisma db push
  tok prisma validate`,
	DisableFlagParsing: true,
	RunE:               runPrisma,
}

func init() {
	registry.Add(func() { registry.Register(prismaCmd) })
}

func runPrisma(cmd *cobra.Command, args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		args = []string{"--help"}
	}

	if shared.Verbose > 0 {
		out.Global().Errorf("Running: prisma %s\n", strings.Join(args, " "))
	}

	execCmd := exec.Command("prisma", args...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	var filtered string
	if len(args) > 0 {
		switch args[0] {
		case "generate":
			filtered = filterPrismaGenerate(raw)
		case "migrate":
			filtered = filterPrismaMigrate(raw)
		case "db":
			filtered = filterPrismaDb(raw)
		case "studio":
			filtered = filterPrismaStudio(raw)
		case "validate":
			filtered = filterPrismaValidate(raw)
		default:
			filtered = filterPrismaOutputCompact(raw)
		}
	} else {
		filtered = filterPrismaOutputCompact(raw)
	}

	if err != nil {
		if hint := shared.TeeOnFailure(raw, "prisma", err); hint != "" {
			filtered = filtered + "\n" + hint
		}
	}

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("prisma %s", strings.Join(args, " ")), "tok prisma", originalTokens, filteredTokens)

	return err
}

func filterPrismaGenerate(raw string) string {
	if shared.UltraCompact {
		return filterPrismaCompact(raw, "generate")
	}

	var result strings.Builder
	var models, enums, fields int
	var errors []string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "model") && strings.Contains(trimmed, "generated") {
			models++
		}
		if strings.Contains(trimmed, "enum") && strings.Contains(trimmed, "generated") {
			enums++
		}
		if strings.Contains(trimmed, "field") && strings.Contains(trimmed, "generated") {
			fields++
		}
		if strings.Contains(strings.ToLower(trimmed), "error") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
		if strings.Contains(trimmed, "Generated") || strings.Contains(trimmed, "generated") {
			result.WriteString(trimmed + "\n")
		}
	}

	if models > 0 || enums > 0 || fields > 0 {
		result.WriteString(fmt.Sprintf("Generated: %d models, %d enums, %d fields\n", models, enums, fields))
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("Errors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-5))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if result.Len() == 0 {
		return "Generate: completed\n"
	}
	return result.String()
}

func filterPrismaMigrate(raw string) string {
	if shared.UltraCompact {
		return filterPrismaCompact(raw, "migrate")
	}

	var result strings.Builder
	var migrations []string
	var applied, rolledBack int
	var errors []string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "Applied") && strings.Contains(trimmed, "migration") {
			applied++
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(trimmed, "Rolled back") || strings.Contains(trimmed, "rolled back") {
			rolledBack++
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(trimmed, "migration") && strings.Contains(trimmed, "...") {
			migrations = append(migrations, shared.TruncateLine(trimmed, 80))
		}
		if strings.Contains(strings.ToLower(trimmed), "error") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
		if strings.Contains(trimmed, "Database") || strings.Contains(trimmed, "database") {
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(trimmed, "The following migration") {
			result.WriteString(trimmed + "\n")
		}
	}

	if applied > 0 || rolledBack > 0 {
		result.WriteString(fmt.Sprintf("Migrations: %d applied", applied))
		if rolledBack > 0 {
			result.WriteString(fmt.Sprintf(", %d rolled back", rolledBack))
		}
		result.WriteString("\n")
	}

	if len(migrations) > 0 {
		result.WriteString("Pending migrations:\n")
		for i, m := range migrations {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(migrations)-10))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", m))
		}
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("Errors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-5))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if result.Len() == 0 {
		return "Migrate: completed\n"
	}
	return result.String()
}

func filterPrismaDb(raw string) string {
	if shared.UltraCompact {
		return filterPrismaCompact(raw, "db")
	}

	var result strings.Builder
	var errors []string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "pushed") || strings.Contains(trimmed, "pulled") ||
			strings.Contains(trimmed, "changes") || strings.Contains(trimmed, "synced") ||
			strings.Contains(trimmed, "applied") || strings.Contains(trimmed, "created") ||
			strings.Contains(trimmed, "Database") || strings.Contains(trimmed, "database") ||
			strings.Contains(trimmed, "Your database") || strings.Contains(trimmed, "is now") {
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(strings.ToLower(trimmed), "error") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("Errors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 5 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-5))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if result.Len() == 0 {
		return "Database: sync completed\n"
	}
	return result.String()
}

func filterPrismaStudio(raw string) string {
	var result strings.Builder

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "http") || strings.Contains(trimmed, "started") ||
			strings.Contains(trimmed, "studio") || strings.Contains(trimmed, "Studio") ||
			strings.Contains(trimmed, "http://") || strings.Contains(trimmed, "https://") {
			result.WriteString(trimmed + "\n")
		}
	}

	if result.Len() == 0 {
		return "Prisma Studio: started\n"
	}
	return result.String()
}

func filterPrismaValidate(raw string) string {
	if shared.UltraCompact {
		for _, line := range strings.Split(raw, "\n") {
			if strings.Contains(strings.ToLower(line), "error") {
				return "validate: failed\n"
			}
		}
		return "validate: ok\n"
	}

	var result strings.Builder
	var errors []string

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "valid") || strings.Contains(trimmed, "Valid") {
			result.WriteString(trimmed + "\n")
		}
		if strings.Contains(strings.ToLower(trimmed), "error") {
			errors = append(errors, shared.TruncateLine(trimmed, 100))
		}
	}

	if len(errors) > 0 {
		result.WriteString(fmt.Sprintf("Validation errors (%d):\n", len(errors)))
		for i, e := range errors {
			if i >= 10 {
				result.WriteString(fmt.Sprintf("  ... +%d more\n", len(errors)-10))
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", e))
		}
	}

	if result.Len() == 0 {
		return "Schema: valid\n"
	}
	return result.String()
}

func filterPrismaCompact(raw string, cmd string) string {
	var errors []string
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(trimmed), "error") {
			errors = append(errors, shared.TruncateLine(trimmed, 80))
		}
	}
	if len(errors) > 0 {
		return fmt.Sprintf("prisma %s: %d errors\n", cmd, len(errors))
	}
	return fmt.Sprintf("prisma %s: ok\n", cmd)
}

func filterPrismaOutputCompact(raw string) string {
	if shared.UltraCompact {
		return filterPrismaCompact(raw, "")
	}

	lines := strings.Split(raw, "\n")
	var result []string
	var inBanner bool

	for _, line := range lines {
		if strings.Contains(line, "│") || strings.Contains(line, "─") ||
			strings.Contains(line, "╭") || strings.Contains(line, "╰") ||
			strings.Contains(line, "▼") || strings.Contains(line, "▲") {
			inBanner = true
			continue
		}

		if inBanner && strings.TrimSpace(line) == "" {
			inBanner = false
			continue
		}

		inBanner = false
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "✔") || strings.HasPrefix(line, "✅") {
			continue
		}

		if line != "" {
			result = append(result, shared.TruncateLine(line, 100))
		}
	}

	if len(result) == 0 {
		return "Prisma command completed"
	}

	var compact []string
	for i, line := range result {
		if i > 0 && line == "" && result[i-1] == "" {
			continue
		}
		compact = append(compact, line)
	}

	if len(compact) > 20 {
		return strings.Join(compact[:20], "\n") + fmt.Sprintf("\n... (%d more lines)", len(compact)-20)
	}
	return strings.Join(compact, "\n")
}
