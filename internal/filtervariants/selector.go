package filtervariants

import (
	"os"
	"path/filepath"
	"strings"
)

type VariantType string

const (
	VariantFileDetection VariantType = "file_detection"
	VariantOutputPattern VariantType = "output_pattern"
)

type Variant struct {
	Name           string
	Type           VariantType
	MatchPatterns  []string
	OutputPatterns []string
	Priority       int
	FilterName     string
}

type ProjectType string

const (
	ProjectGo      ProjectType = "go"
	ProjectRust    ProjectType = "rust"
	ProjectNode    ProjectType = "node"
	ProjectPython  ProjectType = "python"
	ProjectRuby    ProjectType = "ruby"
	ProjectJava    ProjectType = "java"
	ProjectUnknown ProjectType = "unknown"
)

type VariantSelector struct {
	variants []Variant
}

func NewVariantSelector() *VariantSelector {
	vs := &VariantSelector{}
	vs.loadBuiltInVariants()
	return vs
}

func (vs *VariantSelector) loadBuiltInVariants() {
	vs.variants = []Variant{
		{Name: "go_test", Type: VariantFileDetection, MatchPatterns: []string{"go.mod", "go.sum"}, OutputPatterns: []string{"^--- FAIL", "^ok", "^FAIL"}, Priority: 10, FilterName: "go_test"},
		{Name: "rust_test", Type: VariantFileDetection, MatchPatterns: []string{"Cargo.toml"}, OutputPatterns: []string{"^test result", "^running"}, Priority: 10, FilterName: "rust_test"},
		{Name: "node_test", Type: VariantFileDetection, MatchPatterns: []string{"package.json"}, OutputPatterns: []string{"^PASS", "^FAIL", "^Test Suites"}, Priority: 10, FilterName: "node_test"},
		{Name: "python_test", Type: VariantFileDetection, MatchPatterns: []string{"requirements.txt", "pyproject.toml"}, OutputPatterns: []string{"^passed", "^failed", "^ERROR"}, Priority: 10, FilterName: "python_test"},
		{Name: "generic", Type: VariantOutputPattern, OutputPatterns: []string{".*"}, Priority: 0, FilterName: "generic"},
	}
}

func (vs *VariantSelector) DetectProjectType(workDir string) ProjectType {
	detectFiles := map[ProjectType][]string{
		ProjectGo:     {"go.mod", "go.sum"},
		ProjectRust:   {"Cargo.toml", "Cargo.lock"},
		ProjectNode:   {"package.json", "package-lock.json"},
		ProjectPython: {"requirements.txt", "pyproject.toml", "setup.py"},
		ProjectRuby:   {"Gemfile", "Gemfile.lock"},
		ProjectJava:   {"pom.xml", "build.gradle"},
	}

	for projType, files := range detectFiles {
		for _, f := range files {
			path := filepath.Join(workDir, f)
			if _, err := os.Stat(path); err == nil {
				return projType
			}
		}
	}
	return ProjectUnknown
}

func (vs *VariantSelector) SelectVariant(command string, output string, workDir string) *Variant {
	projectType := vs.DetectProjectType(workDir)

	var best *Variant
	for i := range vs.variants {
		v := &vs.variants[i]
		score := 0

		if v.Type == VariantFileDetection {
			for _, pattern := range v.MatchPatterns {
				if strings.Contains(string(projectType), pattern) {
					score += 10
				}
			}
		}

		if v.Type == VariantOutputPattern {
			for _, pattern := range v.OutputPatterns {
				if strings.Contains(output, pattern) {
					score += 5
				}
			}
		}

		if score > 0 {
			if best == nil || v.Priority > best.Priority || (v.Priority == best.Priority && score > 0) {
				best = v
			}
		}
	}

	if best == nil {
		for i := range vs.variants {
			if vs.variants[i].Name == "generic" {
				return &vs.variants[i]
			}
		}
	}

	return best
}

func (vs *VariantSelector) GetVariants() []Variant {
	return vs.variants
}

func (vs *VariantSelector) AddVariant(v Variant) {
	vs.variants = append(vs.variants, v)
}
