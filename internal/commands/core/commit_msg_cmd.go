package core

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
)

var commitMsgCmd = &cobra.Command{
	Use:   "commit-msg",
	Short: "Generate a caveman-style commit message from staged diff",
	Long: `Read the staged git diff (or stdin) and emit a Conventional Commits
subject plus an optional short body. Rule-based — no LLM required.

Detection:
  - feat: new file(s) added
  - fix:  single-file edit under bug-like paths or messages
  - docs: only *.md / *.rst / docs/
  - test: only *_test.* / tests/
  - chore: build / ci / deps / lockfiles
  - refactor: rename/move without net additions

Subject is <= 50 chars. Body is emitted only when multiple scopes are touched.`,
	RunE: runCommitMsg,
}

func runCommitMsg(cmd *cobra.Command, args []string) error {
	diff, err := readDiff()
	if err != nil {
		return err
	}
	if strings.TrimSpace(diff) == "" {
		return fmt.Errorf("no staged changes")
	}

	files := parseChangedFiles(diff)
	if len(files) == 0 {
		return fmt.Errorf("no file changes detected")
	}

	kind, scope := classifyChange(files)
	subject := buildSubject(kind, scope, files)
	body := buildBody(files)

	fmt.Println(subject)
	if body != "" {
		fmt.Println()
		fmt.Println(body)
	}
	return nil
}

func readDiff() (string, error) {
	stat, _ := os.Stdin.Stat()
	if stat != nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		var sb strings.Builder
		s := bufio.NewScanner(os.Stdin)
		s.Buffer(make([]byte, 64*1024), 4*1024*1024)
		for s.Scan() {
			sb.WriteString(s.Text())
			sb.WriteByte('\n')
		}
		return sb.String(), s.Err()
	}

	c := exec.Command("git", "diff", "--cached")
	out, err := c.Output()
	if err != nil {
		return "", fmt.Errorf("git diff --cached: %w", err)
	}
	return string(out), nil
}

type changedFile struct {
	path    string
	added   int
	deleted int
	renamed bool
	isNew   bool
}

var (
	reDiffGit = regexp.MustCompile(`^diff --git a/(\S+) b/(\S+)`)
	reNewFile = regexp.MustCompile(`^new file mode`)
	reRename  = regexp.MustCompile(`^rename (from|to) `)
	reFileHdr = regexp.MustCompile(`^(\+\+\+|---) `)
)

func parseChangedFiles(diff string) []changedFile {
	var files []changedFile
	var cur *changedFile

	for _, line := range strings.Split(diff, "\n") {
		if m := reDiffGit.FindStringSubmatch(line); m != nil {
			if cur != nil {
				files = append(files, *cur)
			}
			cur = &changedFile{path: m[2]}
			continue
		}
		if cur == nil {
			continue
		}
		if reNewFile.MatchString(line) {
			cur.isNew = true
		}
		if reRename.MatchString(line) {
			cur.renamed = true
		}
		if reFileHdr.MatchString(line) {
			continue
		}
		if strings.HasPrefix(line, "+") {
			cur.added++
		} else if strings.HasPrefix(line, "-") {
			cur.deleted++
		}
	}
	if cur != nil {
		files = append(files, *cur)
	}
	return files
}

func classifyChange(files []changedFile) (kind, scope string) {
	var onlyDocs, onlyTests, onlyChore, anyNew, anyRename = true, true, true, false, false
	scopes := map[string]int{}

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.path))
		base := strings.ToLower(filepath.Base(f.path))
		dir := strings.ToLower(filepath.Dir(f.path))

		isDoc := ext == ".md" || ext == ".rst" || strings.HasPrefix(dir, "docs") ||
			base == "readme" || base == "changelog"
		isTest := strings.Contains(base, "_test.") || strings.HasSuffix(base, ".test.ts") ||
			strings.HasSuffix(base, ".test.js") || strings.HasPrefix(dir, "tests") ||
			strings.HasPrefix(dir, "test")
		isChore := base == "go.sum" || base == "go.mod" || base == "package-lock.json" ||
			base == "yarn.lock" || base == "pnpm-lock.yaml" ||
			strings.HasPrefix(dir, ".github") || strings.HasPrefix(dir, "ci") ||
			strings.HasPrefix(dir, "dist") || ext == ".yml" || ext == ".yaml"

		if !isDoc {
			onlyDocs = false
		}
		if !isTest {
			onlyTests = false
		}
		if !isChore {
			onlyChore = false
		}
		if f.isNew {
			anyNew = true
		}
		if f.renamed {
			anyRename = true
		}

		// scope = top-level dir
		parts := strings.Split(f.path, "/")
		if len(parts) > 1 {
			scopes[parts[0]]++
		}
	}

	switch {
	case onlyDocs:
		kind = "docs"
	case onlyTests:
		kind = "test"
	case onlyChore:
		kind = "chore"
	case anyRename && !anyNew:
		kind = "refactor"
	case anyNew:
		kind = "feat"
	default:
		kind = "fix"
	}

	// pick dominant scope if ≥60% of files share it
	if len(scopes) > 0 {
		type kv struct {
			k string
			v int
		}
		var list []kv
		for k, v := range scopes {
			list = append(list, kv{k, v})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].v > list[j].v })
		if float64(list[0].v) >= 0.6*float64(len(files)) {
			scope = list[0].k
		}
	}
	return
}

func buildSubject(kind, scope string, files []changedFile) string {
	verb := "update"
	switch kind {
	case "feat":
		verb = "add"
	case "fix":
		verb = "fix"
	case "docs":
		verb = "update"
	case "test":
		verb = "add"
	case "chore":
		verb = "bump"
	case "refactor":
		verb = "refactor"
	}
	target := "changes"
	if len(files) == 1 {
		target = trimExt(filepath.Base(files[0].path))
	} else if scope != "" {
		target = scope
	} else {
		target = fmt.Sprintf("%d files", len(files))
	}
	subj := kind
	if scope != "" && len(files) > 1 {
		subj = fmt.Sprintf("%s(%s)", kind, scope)
	}
	subj = fmt.Sprintf("%s: %s %s", subj, verb, target)
	if len(subj) > 50 {
		subj = subj[:50]
	}
	return subj
}

func trimExt(s string) string {
	ext := filepath.Ext(s)
	if ext == "" {
		return s
	}
	return strings.TrimSuffix(s, ext)
}

func buildBody(files []changedFile) string {
	if len(files) < 3 {
		return ""
	}
	topDirs := map[string]bool{}
	for _, f := range files {
		parts := strings.SplitN(f.path, "/", 2)
		if len(parts) > 0 {
			topDirs[parts[0]] = true
		}
	}
	if len(topDirs) < 2 {
		return ""
	}
	dirs := make([]string, 0, len(topDirs))
	for d := range topDirs {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return "touches: " + strings.Join(dirs, ", ")
}

func init() {
	registry.Add(func() { registry.Register(commitMsgCmd) })
}
