package deps

import (
	"os"
	"path/filepath"
	"strings"
)

type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
}

type DepSummary struct {
	Language     string       `json:"language"`
	TotalDeps    int          `json:"total_deps"`
	DevDeps      int          `json:"dev_deps"`
	Dependencies []Dependency `json:"dependencies"`
}

func ScanDependencies(workDir string) *DepSummary {
	summary := &DepSummary{}

	for _, file := range []string{"package.json", "go.mod", "Cargo.toml", "requirements.txt", "Gemfile", "pom.xml", "build.gradle", "composer.json", "pubspec.yaml", "mix.exs"} {
		path := filepath.Join(workDir, file)
		if _, err := os.Stat(path); err == nil {
			summary.Language = detectLanguage(file)
			summary.Dependencies = parseDeps(path, summary.Language)
			summary.TotalDeps = len(summary.Dependencies)
			break
		}
	}

	return summary
}

func detectLanguage(file string) string {
	switch file {
	case "package.json":
		return "node"
	case "go.mod":
		return "go"
	case "Cargo.toml":
		return "rust"
	case "requirements.txt":
		return "python"
	case "Gemfile":
		return "ruby"
	case "pom.xml":
		return "java"
	case "build.gradle":
		return "gradle"
	case "composer.json":
		return "php"
	case "pubspec.yaml":
		return "dart"
	case "mix.exs":
		return "elixir"
	default:
		return "unknown"
	}
}

func parseDeps(path, language string) []Dependency {
	var deps []Dependency
	data, err := os.ReadFile(path)
	if err != nil {
		return deps
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		switch language {
		case "go":
			if strings.Contains(trimmed, " ") && !strings.Contains(trimmed, "module") && !strings.Contains(trimmed, "go ") {
				parts := strings.Fields(trimmed)
				if len(parts) >= 2 {
					deps = append(deps, Dependency{Name: parts[0], Version: parts[1], Type: "dependency"})
				}
			}
		case "python":
			if strings.Contains(trimmed, "==") || strings.Contains(trimmed, ">=") {
				parts := strings.Split(trimmed, "==")
				if len(parts) == 2 {
					deps = append(deps, Dependency{Name: parts[0], Version: parts[1], Type: "dependency"})
				}
			}
		default:
			deps = append(deps, Dependency{Name: trimmed, Type: "dependency"})
		}
	}

	return deps
}
