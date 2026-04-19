package system

// ProjectMapCommand implements `tok project_map` - generate a compact
// tree-style map of an entire codebase with token usage estimates.
//
// Inspired by claude-context-optimizer's project_map tool (98% reduction:
// 95,000 → 815 tokens) and Mycelium's project_map command.
//
// The project_map answers "how many tokens does my codebase cost to read?"
// by generating a structured summary: directory tree with file types, line
// counts, sizes, and estimated token counts — in a fraction of the tokens
// of raw `find` or `tree` output.

import (
	"encoding/json"
	"fmt"
	out "github.com/lakshmanpatel/tok/internal/output"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/core"
)

var (
	pmMaxDepth       int
	pmIncludeHidden  bool
	pmTokenEstimate  bool
	pmShowSignatures bool
	pmOutputFormat   string // tree, json, compact
	pmIgnore         string
)

var projectMapCmd = &cobra.Command{
	Use:   "project_map [dir]",
	Short: "Generate a compact map of a codebase with token estimates",
	Long: `Generate a structured, token-optimized map of an entire directory tree.

Shows file types, line counts, sizes, and estimated token counts in a
fraction of the tokens of raw 'find' or 'tree' output.

This answers: "How many tokens would it cost to read my entire codebase?"

Examples:
  tok project_map              # Map current directory
  tok project_map /path/to/repo --depth 2
  tok project_map . --tokens   # Include token estimates
  tok project_map . --signatures  # Include function/class signatures
  tok project_map . --format json # JSON output for processing`,
	Aliases: []string{"pmap", "repo-map"},
	RunE:    runProjectMap,
}

func init() {
	projectMapCmd.Flags().IntVarP(&pmMaxDepth, "depth", "d", 4, "Maximum directory depth")
	projectMapCmd.Flags().BoolVarP(&pmTokenEstimate, "tokens", "t", true, "Estimate tokens per file")
	projectMapCmd.Flags().BoolVarP(&pmIncludeHidden, "hidden", "H", false, "Include hidden files/dirs")
	projectMapCmd.Flags().BoolVarP(&pmShowSignatures, "signatures", "s", false, "Include function/class signatures for code files")
	projectMapCmd.Flags().StringVarP(&pmOutputFormat, "format", "f", "tree", "Output format: tree, compact, json")
	projectMapCmd.Flags().StringVarP(&pmIgnore, "ignore", "i", "", "Comma-separated patterns to ignore (e.g., node_modules,.git)")

	registry.Add(func() { registry.Register(projectMapCmd) })
}

// DirNode represents a directory in the project map.
type DirNode struct {
	Name    string
	Path    string
	Files   []FileInfo
	SubDirs []*DirNode
	Depth   int
}

// FileInfo represents a single file's metadata.
type FileInfo struct {
	Name        string `json:"name"`
	Ext         string `json:"ext"`
	Size        int64  `json:"size"`
	Lines       int    `json:"lines"`
	Tokens      int    `json:"tokens,omitempty"`
	Signatures  int    `json:"signatures,omitempty"`
	IsGenerated bool   `json:"-"`
}

func runProjectMap(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	if info, err := os.Stat(absDir); err != nil || !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	root, err := buildProjectTree(absDir, 0)
	if err != nil {
		return err
	}
	computeStats(root)

	switch pmOutputFormat {
	case "json":
		return printProjectMapJSON(root, absDir)
	case "compact":
		printProjectMapCompact(root, absDir)
	default:
		printProjectMapTree(root, absDir)
	}

	return nil
}

// buildProjectTree recursively walks the directory and builds a tree.
func buildProjectTree(path string, depth int) (*DirNode, error) {
	if depth > pmMaxDepth {
		return nil, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	node := &DirNode{
		Name:  filepath.Base(path),
		Path:  path,
		Depth: depth,
	}

	for _, entry := range entries {
		name := entry.Name()

		if !pmIncludeHidden && strings.HasPrefix(name, ".") {
			continue
		}

		if shouldIgnore(name) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.IsDir() {
			child, childErr := buildProjectTree(filepath.Join(path, name), depth+1)
			if childErr != nil {
				continue
			}
			if child != nil {
				node.SubDirs = append(node.SubDirs, child)
			}
		} else {
			fi := FileInfo{
				Name: name,
				Ext:  trimmedExt(name),
				Size: info.Size(),
			}
			if !isGeneratedFile(name, info.Size()) {
				node.Files = append(node.Files, fi)
			}
		}
	}

	// Sort for deterministic output
	sort.Slice(node.Files, func(i, j int) bool { return node.Files[i].Name < node.Files[j].Name })
	sort.Slice(node.SubDirs, func(i, j int) bool { return node.SubDirs[i].Name < node.SubDirs[j].Name })

	return node, nil
}

// computeStats computes line counts, token estimates, and signatures.
func computeStats(node *DirNode) {
	for i := range node.Files {
		f := &node.Files[i]
		fPath := filepath.Join(node.Path, f.Name)

		// Line count
		if data, err := os.ReadFile(fPath); err == nil {
			f.Lines = strings.Count(string(data), "\n")
			if len(data) > 0 && data[len(data)-1] != '\n' {
				f.Lines++
			}
		}

		// Token estimate
		if pmTokenEstimate {
			if data, err := os.ReadFile(fPath); err == nil {
				f.Tokens = core.EstimateTokens(string(data))
			}
		}

		// Signatures (for code files)
		if pmShowSignatures && isCodeFile(f.Ext) {
			if data, err := os.ReadFile(fPath); err == nil {
				f.Signatures = countSignatures(string(data))
			}
		}
	}

	for _, sub := range node.SubDirs {
		computeStats(sub)
	}
}

// printProjectMapTree prints the project map in tree format.
func printProjectMapTree(root *DirNode, basePath string) {
	totalFiles := 0
	totalLines := 0
	totalTokens := 0
	totalSize := int64(0)
	fileCountByExt := make(map[string]int)
	fileCountByDir := make(map[string]int)

	// First pass: collect totals
	collectStats(root, &totalFiles, &totalLines, &totalTokens, &totalSize, fileCountByExt, fileCountByDir)

	// Header
	color.New(color.Bold).Fprintf(os.Stderr, "%s\n", relativePath(root.Path, basePath))

	if shared.IsVerbose() {
		out.Global().Errorf("\n")
	}

	// Print tree
	printDirTree(root, "", true, fileCountByDir)

	// Summary
	out.Global().Errorf("\n")
	color.New(color.Bold).Printf("Summary\n")
	out.Global().Errorf("  Files:     %d\n", totalFiles)
	out.Global().Errorf("  Lines:     %s\n", formatNumber(totalLines))
	if pmTokenEstimate {
		out.Global().Errorf("  Tokens:    %s (~%d KB BPE estimate)\n", formatNumber(totalTokens), totalTokens/4)
		out.Global().Errorf("  Cost:      $%.2f (at $3/MTok input)\n", float64(totalTokens)/1e6*3.0)
	}
	out.Global().Errorf("  Size:      %s\n", formatBytes(totalSize))
	out.Global().Errorf("  Dirs:      %d\n", countDirs(root))

	if len(fileCountByExt) > 0 {
		out.Global().Errorf("\n  By extension:\n")
		// Sort by count descending
		type extCount struct {
			ext   string
			count int
		}
		var sorted []extCount
		for ext, count := range fileCountByExt {
			sorted = append(sorted, extCount{ext, count})
		}
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })
		for _, ec := range sorted {
			ext := ec.ext
			if ext == "" {
				ext = "(no ext)"
			}
			out.Global().Errorf("    %-12s %d files\n", ext, ec.count)
		}
	}
}

// printDirTree prints a directory tree with ANSI formatting.
func printDirTree(node *DirNode, prefix string, isLast bool, fileCountByDir map[string]int) {
	isRoot := node.Depth == 0

	if !isRoot {
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		if isCodeDir(node.Name) {
			color.New(color.FgCyan, color.Bold).Printf("%s%s%s/\n", prefix, connector, node.Name)
		} else {
			color.New(color.FgYellow, color.Bold).Printf("%s%s%s/\n", prefix, connector, node.Name)
		}

		extension := prefix
		if isLast {
			extension += "    "
		} else {
			extension += "│   "
		}

		printDirContents(node, extension, fileCountByDir)
	} else {
		// Root directory - print directly
		printDirContents(node, "", fileCountByDir)
	}

	// Subdirectories
	for i, sub := range node.SubDirs {
		isLastSub := i == len(node.SubDirs)-1
		printDirTree(sub, prefix+extensionOrRoot(isRoot), isLastSub, fileCountByDir)
	}
}

// printDirContents prints the files in a directory with compact formatting.
func printDirContents(node *DirNode, indent string, fileCountByDir map[string]int) {
	if len(node.Files) == 0 && len(node.SubDirs) == 0 {
		color.New(color.Faint).Printf("%s(empty)\n", indent)
		return
	}

	// Group files by extension
	groups := make(map[string][]FileInfo)
	for _, f := range node.Files {
		ext := f.Ext
		if ext == "" {
			ext = "(none)"
		}
		groups[ext] = append(groups[ext], f)
	}

	// Print file groups
	exts := sortedKeys(groups)
	for _, ext := range exts {
		files := groups[ext]
		if len(files) == 1 {
			f := files[0]
			printFile(f, indent, false)
		} else {
			// Summary format: 8 .go files (2.4K lines, 12K tokens)
			totalLines := 0
			totalTokens := 0
			totalSize := int64(0)
			for _, f := range files {
				totalLines += f.Lines
				totalTokens += f.Tokens
				totalSize += f.Size
			}

			extLabel := ext
			if ext == "(none)" {
				extLabel = "files"
			}

			if shared.IsUltraCompact() {
				out.Global().Errorf("%s%d %s (%s, %s tokens, %s)...\n",
					indent, len(files), extLabel,
					formatNumber(totalLines), formatNumber(totalTokens), formatBytes(totalSize))
			} else {
				color.New(color.FgGreen).Fprintf(os.Stderr, "%s%d %s file", indent, len(files), extLabel)
				if len(files) > 1 {
					out.Global().Errorf("s")
				}
				out.Global().Errorf(" (%s lines, %s tokens, %s)...\n",
					formatNumber(totalLines), formatNumber(totalTokens), formatBytes(totalSize))
			}

			// Show individual files in verbose mode
			if shared.IsVerbose() {
				for _, f := range files {
					printFile(f, indent+"  ", false)
				}
			}
		}
	}
}

// printFile prints a single file with compact formatting.
func printFile(f FileInfo, indent string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	if isCodeFile(f.Ext) {
		color.New(color.FgCyan).Printf("%s%s%s", indent, connector, f.Name)
	} else if isTextFile(f.Ext) {
		out.Global().Errorf("%s%s%s", indent, connector, f.Name)
	} else {
		color.New(color.Faint).Printf("%s%s%s", indent, connector, f.Name)
	}

	if shared.IsVerbose() {
		out.Global().Errorf(" (%d lines", f.Lines)
		if f.Tokens > 0 {
			out.Global().Errorf(", %d tokens", f.Tokens)
		}
		if f.Signatures > 0 {
			out.Global().Errorf(", %d signatures", f.Signatures)
		}
		out.Global().Errorf(")")
	}

	out.Global().Errorf("\n")
}

// extensionOrRoot returns the proper indentation for root vs sub-directories.
func extensionOrRoot(isRoot bool) string {
	if isRoot {
		return ""
	}
	return "    "
}

// printProjectMapCompact prints a minimal one-line-per-directory summary.
func printProjectMapCompact(root *DirNode, basePath string) {
	printCompactDir(root, basePath, "")
}

// printCompactDir prints a directory in compact format.
func printCompactDir(node *DirNode, basePath string, prefix string) {
	totalFiles := 0
	totalLines := 0
	totalTokens := 0

	for _, f := range node.Files {
		totalFiles++
		totalLines += f.Lines
		totalTokens += f.Tokens
	}

	for _, sub := range node.SubDirs {
		sFiles, sLines, sTokens := countCompact(sub)
		totalFiles += sFiles
		totalLines += sLines
		totalTokens += sTokens
	}

	if totalFiles > 0 {
		relPath := relativePath(node.Path, basePath)
		if relPath == "." {
			relPath = basePath
		}

		out.Global().Printf("%-50s %6d files %8d lines %10d tokens\n",
			relPath, totalFiles, totalLines, totalTokens)
	}

	for _, sub := range node.SubDirs {
		printCompactDir(sub, basePath, prefix+"  ")
	}
}

// countCompact counts files, lines, and tokens in a subtree.
func countCompact(node *DirNode) (int, int, int) {
	files := len(node.Files)
	lines := 0
	tokens := 0

	for _, f := range node.Files {
		lines += f.Lines
		tokens += f.Tokens
	}

	for _, sub := range node.SubDirs {
		sf, sl, st := countCompact(sub)
		files += sf
		lines += sl
		tokens += st
	}

	return files, lines, tokens
}

// printProjectMapJSON prints the project map as JSON.
func printProjectMapJSON(root *DirNode, basePath string) error {
	type extCount struct {
		Ext   string `json:"ext"`
		Count int    `json:"count"`
		Lines int    `json:"lines"`
	}

	type treeNode struct {
		Name       string     `json:"name"`
		Path       string     `json:"path"`
		Files      int        `json:"file_count"`
		Lines      int        `json:"total_lines"`
		Tokens     int        `json:"total_tokens"`
		Size       int64      `json:"total_size"`
		SubDirs    int        `json:"subdir_count"`
		Extensions []extCount `json:"extensions,omitempty"`
		Children   []treeNode `json:"children,omitempty"`
	}

	var walkTree func(*DirNode) treeNode
	walkTree = func(n *DirNode) treeNode {
		totalLines := 0
		totalTokens := 0
		totalSize := int64(0)
		extMap := make(map[string]*extCount)

		for _, f := range n.Files {
			totalLines += f.Lines
			totalTokens += f.Tokens
			totalSize += f.Size
			if _, ok := extMap[f.Ext]; !ok {
				extMap[f.Ext] = &extCount{Ext: f.Ext}
			}
			extMap[f.Ext].Count++
			extMap[f.Ext].Lines += f.Lines
		}

		var exts []extCount
		for _, ec := range extMap {
			exts = append(exts, *ec)
		}
		sort.Slice(exts, func(i, j int) bool { return exts[i].Count > exts[j].Count })

		tj := treeNode{
			Name:       n.Name,
			Path:       relativePath(n.Path, basePath),
			Files:      len(n.Files),
			Lines:      totalLines,
			Tokens:     totalTokens,
			Size:       totalSize,
			SubDirs:    len(n.SubDirs),
			Extensions: exts,
		}

		for _, sub := range n.SubDirs {
			tj.Children = append(tj.Children, walkTree(sub))
		}

		return tj
	}

	result := walkTree(root)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// Helper functions

func countDirs(node *DirNode) int {
	count := 0
	for _, sub := range node.SubDirs {
		count += 1 + countDirs(sub)
	}
	return count
}

func collectStats(node *DirNode, totalFiles, totalLines, totalTokens *int, totalSize *int64, fileCountByExt, fileCountByDir map[string]int) {
	for _, f := range node.Files {
		*totalFiles++
		*totalLines += f.Lines
		*totalTokens += f.Tokens
		*totalSize += f.Size
		ext := f.Ext
		if ext == "" {
			ext = "(no ext)"
		}
		fileCountByExt[ext]++
		fileCountByDir[node.Path]++
	}
	for _, sub := range node.SubDirs {
		collectStats(sub, totalFiles, totalLines, totalTokens, totalSize, fileCountByExt, fileCountByDir)
	}
}

func trimmedExt(name string) string {
	ext := filepath.Ext(name)
	if ext != "" {
		return strings.ToLower(ext[1:])
	}
	return ""
}

func isCodeFile(ext string) bool {
	codeExts := map[string]bool{
		"go": true, "rs": true, "py": true, "js": true, "ts": true, "tsx": true,
		"jsx": true, "java": true, "c": true, "cpp": true, "h": true, "hpp": true,
		"rb": true, "php": true, "swift": true, "kt": true, "scala": true,
		"sh": true, "bash": true, "zsh": true, "fish": true,
		"html": true, "css": true, "scss": true, "sass": true,
		"vue": true, "svelte": true, "sol": true, "toml": true, "yaml": true, "yml": true,
		"json": true, "xml": true, "md": true, "rst": true, "lua": true,
	}
	return codeExts[ext]
}

func isTextFile(ext string) bool {
	textExts := map[string]bool{
		"txt": true, "log": true, "cfg": true, "conf": true, "ini": true,
		"env": true, "lock": true, "csv": true, "tsv": true, "sql": true,
	}
	return textExts[ext]
}

func isCodeDir(name string) bool {
	codeDirs := map[string]bool{
		"src": true, "lib": true, "internal": true, "pkg": true, "cmd": true,
		"tests": true, "test": true, "__tests__": true,
	}
	return codeDirs[name]
}

func isGeneratedFile(name string, size int64) bool {
	// Skip very large files (>500KB) - likely generated or vendor
	if size > 500*1024 {
		return true
	}

	generated := map[string]bool{
		"go.sum": true, "package-lock.json": true, "yarn.lock": true,
		"pnpm-lock.yaml": true, "Cargo.lock": true, "poetry.lock": true,
		"Pipfile.lock": true, "vendor": true, "node_modules": true,
	}
	return generated[name]
}

func countSignatures(content string) int {
	count := 0
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "func ") ||
			strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "def ") ||
			strings.HasPrefix(trimmed, "struct ") ||
			strings.HasPrefix(trimmed, "interface ") ||
			strings.HasPrefix(trimmed, "type ") ||
			strings.HasPrefix(trimmed, "enum ") ||
			strings.HasPrefix(trimmed, "pub fn ") ||
			strings.HasPrefix(trimmed, "impl ") ||
			strings.HasPrefix(trimmed, "mod ") {
			count++
		}
	}
	return count
}

func formatNumber(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n/div >= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	if exp < len(units) {
		return fmt.Sprintf("%.1f %s", float64(n)/float64(div), units[exp])
	}
	return fmt.Sprintf("%.1f %s", float64(n)/float64(div), "PB")
}

func relativePath(path, base string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return rel
}

func shouldIgnore(name string) bool {
	ignoreDirs := map[string]bool{
		".git": true, ".svn": true, ".hg": true,
		"node_modules": true, "vendor": true, ".venv": true, "venv": true,
		"__pycache__": true, ".tox": true, ".mypy_cache": true,
		"dist": true, "build": true, "target": true, "out": true,
		".next": true, ".nuxt": true, ".svelte-kit": true,
		".idea": true, ".vscode": true, ".vs": true,
	}

	if pmIgnore != "" {
		patterns := strings.Split(pmIgnore, ",")
		for _, p := range patterns {
			p = strings.TrimSpace(p)
			if p != "" && (name == p || strings.Contains(name, p)) {
				return true
			}
		}
	}

	return ignoreDirs[name]
}

func sortedKeys(m map[string][]FileInfo) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
