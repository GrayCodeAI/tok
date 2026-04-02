package langcompress

import (
	"os"
	"strings"
)

type LanguageCompressor struct{}

func NewLanguageCompressor() *LanguageCompressor {
	return &LanguageCompressor{}
}

func (c *LanguageCompressor) Compress(content, language string) string {
	switch language {
	case "flutter", "dart":
		return c.compressDart(content)
	case "swift":
		return c.compressSwift(content)
	case "zig":
		return c.compressZig(content)
	case "deno", "typescript":
		return c.compressDeno(content)
	case "bun":
		return c.compressBun(content)
	case "rust":
		return c.compressRust(content)
	case "cmake":
		return c.compressCMake(content)
	case "composer", "php":
		return c.compressComposer(content)
	case "bazel":
		return c.compressBazel(content)
	case "python", "poetry", "uv":
		return c.compressPython(content)
	default:
		return content
	}
}

func (c *LanguageCompressor) compressDart(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "void ") || strings.HasPrefix(trimmed, "Widget ") ||
			strings.HasPrefix(trimmed, "final ") || strings.HasPrefix(trimmed, "const ") ||
			strings.Contains(trimmed, "setState") || strings.Contains(trimmed, "build(") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (c *LanguageCompressor) compressSwift(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "func ") ||
			strings.HasPrefix(trimmed, "class ") || strings.HasPrefix(trimmed, "struct ") ||
			strings.HasPrefix(trimmed, "protocol ") || strings.HasPrefix(trimmed, "extension ") ||
			strings.HasPrefix(trimmed, "var ") || strings.HasPrefix(trimmed, "let ") ||
			strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "override ") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (c *LanguageCompressor) compressZig(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "const ") || strings.HasPrefix(trimmed, "pub ") ||
			strings.HasPrefix(trimmed, "fn ") || strings.HasPrefix(trimmed, "test ") ||
			strings.HasPrefix(trimmed, "var ") || strings.HasPrefix(trimmed, "import ") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (c *LanguageCompressor) compressDeno(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "export ") ||
			strings.HasPrefix(trimmed, "function ") || strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "interface ") || strings.HasPrefix(trimmed, "type ") ||
			strings.HasPrefix(trimmed, "const ") || strings.HasPrefix(trimmed, "Deno.") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (c *LanguageCompressor) compressBun(content string) string {
	return c.compressDeno(content)
}

func (c *LanguageCompressor) compressRust(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "use ") || strings.HasPrefix(trimmed, "fn ") ||
			strings.HasPrefix(trimmed, "pub ") || strings.HasPrefix(trimmed, "struct ") ||
			strings.HasPrefix(trimmed, "enum ") || strings.HasPrefix(trimmed, "impl ") ||
			strings.HasPrefix(trimmed, "trait ") || strings.HasPrefix(trimmed, "mod ") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (c *LanguageCompressor) compressCMake(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "add_executable") || strings.Contains(trimmed, "target_link") ||
			strings.Contains(trimmed, "find_package") || strings.Contains(trimmed, "include_directories") ||
			strings.Contains(trimmed, "set(") || strings.Contains(trimmed, "project(") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (c *LanguageCompressor) compressComposer(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "require") || strings.Contains(trimmed, "autoload") ||
			strings.Contains(trimmed, "namespace") || strings.Contains(trimmed, "use ") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (c *LanguageCompressor) compressBazel(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "cc_library") || strings.Contains(trimmed, "cc_binary") ||
			strings.Contains(trimmed, "java_library") || strings.Contains(trimmed, "py_library") ||
			strings.Contains(trimmed, "deps") || strings.Contains(trimmed, "srcs") ||
			strings.Contains(trimmed, "name") || strings.Contains(trimmed, "load(") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (c *LanguageCompressor) compressPython(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") ||
			strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "async def ") || strings.HasPrefix(trimmed, "class ") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func (c *LanguageCompressor) DetectLanguage(filePath string) string {
	ext := ""
	if idx := strings.LastIndex(filePath, "."); idx >= 0 {
		ext = strings.ToLower(filePath[idx+1:])
	}

	switch ext {
	case "dart":
		return "dart"
	case "swift":
		return "swift"
	case "zig":
		return "zig"
	case "ts":
		return "typescript"
	case "rs":
		return "rust"
	case "cmake":
		return "cmake"
	case "php":
		return "php"
	case "bzl":
		return "bazel"
	case "py":
		return "python"
	default:
		return "unknown"
	}
}

func (c *LanguageCompressor) IsPackageFile(filePath string) bool {
	base := ""
	if idx := strings.LastIndex(filePath, "/"); idx >= 0 {
		base = filePath[idx+1:]
	}
	packageFiles := []string{
		"pubspec.yaml", "pubspec.lock", "Package.swift", "Package.resolved",
		"build.zig", "build.zig.lock", "package.json", "deno.json", "bun.lockb",
		"Cargo.toml", "Cargo.lock", "CMakeLists.txt", "composer.json",
		"composer.lock", "BUILD", "WORKSPACE", "MODULE.bazel",
		"requirements.txt", "pyproject.toml", "setup.py", "poetry.lock",
		"uv.lock", "bun.lockb", "deno.lock",
	}
	for _, f := range packageFiles {
		if f == base {
			return true
		}
	}
	return false
}

func (c *LanguageCompressor) StripLockFileContent(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "---" || trimmed == "..." {
			continue
		}
		if strings.Contains(trimmed, "resolved") || strings.Contains(trimmed, "integrity") ||
			strings.Contains(trimmed, "shasum") || strings.Contains(trimmed, "tarball") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

func ScanLanguageDir(dir string, lang string) []string {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := ""
		if idx := strings.LastIndex(entry.Name(), "."); idx >= 0 {
			ext = strings.ToLower(entry.Name()[idx+1:])
		}
		langExts := map[string][]string{
			"dart":     {"dart"},
			"swift":    {"swift"},
			"zig":      {"zig"},
			"deno":     {"ts", "tsx"},
			"bun":      {"ts", "tsx", "js", "jsx"},
			"rust":     {"rs"},
			"cmake":    {"cmake"},
			"composer": {"php"},
			"bazel":    {"bzl"},
			"python":   {"py"},
		}
		if exts, ok := langExts[lang]; ok {
			for _, e := range exts {
				if ext == e {
					files = append(files, entry.Name())
				}
			}
		}
	}
	return files
}
