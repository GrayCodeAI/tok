package core

import (
	"strings"
	"testing"
)

const sampleDiffNewFile = `diff --git a/src/auth.go b/src/auth.go
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/src/auth.go
@@ -0,0 +1,3 @@
+package src
+
+func Login() {}
`

const sampleDiffEdit = `diff --git a/src/bug.go b/src/bug.go
index 111..222 100644
--- a/src/bug.go
+++ b/src/bug.go
@@ -1,3 +1,3 @@
 package src
-func Foo() int { return 0 }
+func Foo() int { return 42 }
`

const sampleDiffDocs = `diff --git a/docs/intro.md b/docs/intro.md
index 111..222 100644
--- a/docs/intro.md
+++ b/docs/intro.md
@@ -1 +1,2 @@
 # Intro
+Added line.
`

const sampleDiffTests = `diff --git a/internal/foo/foo_test.go b/internal/foo/foo_test.go
new file mode 100644
--- /dev/null
+++ b/internal/foo/foo_test.go
@@ -0,0 +1,3 @@
+package foo
+import "testing"
+func TestX(t *testing.T) {}
`

const sampleDiffChore = `diff --git a/go.sum b/go.sum
index 111..222 100644
--- a/go.sum
+++ b/go.sum
@@ -1 +1,2 @@
 foo
+bar
`

const sampleDiffMulti = `diff --git a/src/a.go b/src/a.go
index 1..2 100644
--- a/src/a.go
+++ b/src/a.go
@@ -1,1 +1,1 @@
-old
+new
diff --git a/src/b.go b/src/b.go
index 1..2 100644
--- a/src/b.go
+++ b/src/b.go
@@ -1,1 +1,1 @@
-x
+y
diff --git a/src/c.go b/src/c.go
index 1..2 100644
--- a/src/c.go
+++ b/src/c.go
@@ -1,1 +1,1 @@
-p
+q
`

func TestParseChangedFiles_DetectsNewFile(t *testing.T) {
	files := parseChangedFiles(sampleDiffNewFile)
	if len(files) != 1 {
		t.Fatalf("want 1 file, got %d", len(files))
	}
	if !files[0].isNew {
		t.Errorf("expected isNew=true for %+v", files[0])
	}
	if files[0].added != 3 {
		t.Errorf("expected 3 adds, got %d", files[0].added)
	}
}

func TestParseChangedFiles_MultiFile(t *testing.T) {
	files := parseChangedFiles(sampleDiffMulti)
	if len(files) != 3 {
		t.Fatalf("want 3 files, got %d", len(files))
	}
	for _, f := range files {
		if f.isNew {
			t.Errorf("edit diff should not mark files new: %+v", f)
		}
	}
}

func TestClassifyChange_Kinds(t *testing.T) {
	cases := []struct {
		name string
		diff string
		want string
	}{
		{"feat", sampleDiffNewFile, "feat"},
		{"fix", sampleDiffEdit, "fix"},
		{"docs", sampleDiffDocs, "docs"},
		{"test", sampleDiffTests, "test"},
		{"chore", sampleDiffChore, "chore"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := parseChangedFiles(c.diff)
			kind, _ := classifyChange(files)
			if kind != c.want {
				t.Errorf("kind=%q, want %q", kind, c.want)
			}
		})
	}
}

func TestClassifyChange_Scope(t *testing.T) {
	files := parseChangedFiles(sampleDiffMulti)
	_, scope := classifyChange(files)
	if scope != "src" {
		t.Errorf("scope=%q, want src", scope)
	}
}

func TestBuildSubject_UnderLimit(t *testing.T) {
	files := parseChangedFiles(sampleDiffNewFile)
	kind, scope := classifyChange(files)
	subj := buildSubject(kind, scope, files)
	if len(subj) > 50 {
		t.Errorf("subject > 50 chars: %q", subj)
	}
	if !strings.HasPrefix(subj, "feat:") {
		t.Errorf("expected feat prefix, got %q", subj)
	}
}

func TestBuildSubject_MultiFileIncludesScope(t *testing.T) {
	files := parseChangedFiles(sampleDiffMulti)
	kind, scope := classifyChange(files)
	subj := buildSubject(kind, scope, files)
	if !strings.Contains(subj, "(src)") {
		t.Errorf("expected (src) scope in subject, got %q", subj)
	}
}

func TestBuildBody_EmptyForSingleDir(t *testing.T) {
	files := parseChangedFiles(sampleDiffMulti)
	if body := buildBody(files); body != "" {
		t.Errorf("single-dir diff should yield empty body, got %q", body)
	}
}

func TestBuildBody_MultiDir(t *testing.T) {
	multi := sampleDiffMulti + `diff --git a/docs/intro.md b/docs/intro.md
index 1..2 100644
--- a/docs/intro.md
+++ b/docs/intro.md
@@ -1,1 +1,1 @@
-a
+b
diff --git a/tests/foo.go b/tests/foo.go
index 1..2 100644
--- a/tests/foo.go
+++ b/tests/foo.go
@@ -1,1 +1,1 @@
-a
+b
`
	files := parseChangedFiles(multi)
	body := buildBody(files)
	if !strings.Contains(body, "docs") || !strings.Contains(body, "src") || !strings.Contains(body, "tests") {
		t.Errorf("body missing expected dirs: %q", body)
	}
}
