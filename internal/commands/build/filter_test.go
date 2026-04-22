package build

import (
	"strings"
	"testing"

	"github.com/GrayCodeAI/tok/internal/commands/shared"
)

func TestFilterTscOutput_NoErrors(t *testing.T) {
	got := filterTscOutput("Found 0 errors. Watching for file changes.")
	if !strings.Contains(got, "No errors") {
		t.Errorf("expected 'No errors' in output, got %q", got)
	}
}

func TestFilterTscOutput_WithErrors(t *testing.T) {
	input := `src/index.ts(10,5): error TS2322: Type 'string' is not assignable to type 'number'.
src/index.ts(20,3): error TS2304: Cannot find name 'foo'.`
	got := filterTscOutput(input)
	if !strings.Contains(got, "errors") {
		t.Errorf("expected 'errors' in output, got %q", got)
	}
}

func TestFilterTscOutput_UltraCompact(t *testing.T) {
	orig := shared.UltraCompact
	defer func() { shared.UltraCompact = orig }()
	shared.UltraCompact = true

	input := `src/index.ts(10,5): error TS2322: Type 'string' is not assignable.`
	got := filterTscOutput(input)
	if !strings.Contains(got, "errors") {
		t.Errorf("expected 'errors' in output, got %q", got)
	}
}

func TestFilterTscOutput_UltraCompact_NoErrors(t *testing.T) {
	orig := shared.UltraCompact
	defer func() { shared.UltraCompact = orig }()
	shared.UltraCompact = true

	got := filterTscOutput("Found 0 errors.")
	if !strings.Contains(got, "ok") {
		t.Errorf("expected 'ok' in output, got %q", got)
	}
}

func TestFilterPrismaGenerate(t *testing.T) {
	input := `Environment variables loaded from .env
Prisma schema loaded from prisma/schema.prisma
✔ Generated Prisma Client (v5.0.0) to ./node_modules/@prisma/client`
	got := filterPrismaGenerate(input)
	if !strings.Contains(got, "Generated") {
		t.Errorf("expected 'Generated' in output, got %q", got)
	}
}

func TestFilterPrismaMigrate(t *testing.T) {
	input := `The following migration(s) have been created and applied from new schema changes:
migrations/
  20240101000000_init/
    migration.sql`
	got := filterPrismaMigrate(input)
	if got == "" {
		t.Error("expected non-empty output")
	}
}

func TestFilterPrismaDb(t *testing.T) {
	input := `🚀  Your database is now in sync with your schema.`
	got := filterPrismaDb(input)
	if got == "" {
		t.Error("expected non-empty output")
	}
}

func TestFilterPrismaStudio(t *testing.T) {
	input := `Prisma Studio is running at: http://localhost:5555`
	got := filterPrismaStudio(input)
	if !strings.Contains(got, "Studio") {
		t.Errorf("expected 'Studio' in output, got %q", got)
	}
}

func TestFilterPrismaValidate(t *testing.T) {
	got := filterPrismaValidate("The schema at prisma/schema.prisma is valid")
	if !strings.Contains(got, "valid") {
		t.Errorf("expected 'valid' in output, got %q", got)
	}
}

func TestFilterPrismaOutputCompact(t *testing.T) {
	got := filterPrismaOutputCompact("Prisma output")
	if !strings.Contains(got, "Prisma") {
		t.Errorf("expected 'Prisma' in output, got %q", got)
	}
}

func TestFilterNextOutputCompact(t *testing.T) {
	input := `✓ Compiled successfully in 2.3s
  ○ /                            1.2 kB
  ○ /about                       800 B`
	got := filterNextOutputCompact(input)
	if !strings.Contains(got, "static") {
		t.Errorf("expected 'static' in output, got %q", got)
	}
}

func TestFilterNextOutputCompact_Error(t *testing.T) {
	input := `Error: Cannot find module 'react'
  at Object.<anonymous> (/app/pages/index.tsx:1:0)`
	got := filterNextOutputCompact(input)
	if !strings.Contains(got, "Error") {
		t.Errorf("expected 'Error' in output, got %q", got)
	}
}
