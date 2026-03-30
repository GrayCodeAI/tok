package fixtures

// CLIOutputFixtures contains real-world CLI outputs for benchmarking
// These are representative samples from common development tools

// GitStatusOutput simulates typical git status output
var GitStatusOutput = `On branch main
Your branch is up to date with 'origin/main'.

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   internal/filter/pipeline.go
	modified:   internal/commands/analysis/stats.go
	modified:   Makefile

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	PERFORMANCE_REPORT.md
	OPTIMIZATION_PLAN.md

no changes added to commit (use "git add" and/or "git commit -a")
`

// GitLogOutput simulates typical git log output
var GitLogOutput = `commit abc123def456789012345678901234567890abcd (HEAD -> main, origin/main)
Author: Developer <dev@example.com>
Date:   Mon Mar 30 10:15:22 2026 -0700

    feat: add cache monitoring to stats command
    
    - Add --cache flag to show cache statistics
    - Display hit rate and efficiency metrics
    - Integrate with FingerprintCache

commit def456abc123789012345678901234567890abcd
Author: Developer <dev@example.com>
Date:   Mon Mar 30 09:30:15 2026 -0700

    fix: resolve import cycle in filter/toml integration
    
    - Use setter pattern for TOML filter wrapper
    - Pipeline no longer imports toml package directly

commit ghi789jkl012345678901234567890abcd
Author: Developer <dev@example.com>
Date:   Sun Mar 29 16:45:00 2026 -0700

    feat: integrate TOML filters into pipeline
    
    - Add TOML filter as Layer 0
    - Support 94 built-in filter templates
    - Command-aware pattern matching
`

// GitDiffOutput simulates typical git diff output
var GitDiffOutput = `diff --git a/internal/filter/pipeline.go b/internal/filter/pipeline.go
index abc1234..def5678 100644
--- a/internal/filter/pipeline.go
+++ b/internal/filter/pipeline.go
@@ -664,6 +664,23 @@ func (p *PipelineCoordinator) Process(input string) (string, *PipelineStats) {
 	stats := &PipelineStats{
 		OriginalTokens: core.EstimateTokens(input),
 		LayerStats:     make(map[string]LayerStat),
 	}
 
+	// TOML Filter: Apply first if configured
+	if p.tomlFilterWrapper != nil && p.config.EnableTOMLFilter {
+		filtered, saved := p.tomlFilterWrapper.Apply(output, ModeMinimal)
+		if saved > 0 {
+			stats.LayerStats["0_toml_filter"] = LayerStat{
+				TokensSaved: saved,
+				Duration:    0,
+			}
+			output = filtered
+			stats.TotalSaved += saved
+		}
+	}
+
 	output := input
`

// DockerPsOutput simulates docker ps output
var DockerPsOutput = `CONTAINER ID   IMAGE                    COMMAND                  CREATED       STATUS       PORTS                    NAMES
abc123def456   nginx:latest             "/docker-entrypoint.…"   2 hours ago   Up 2 hours   0.0.0.0:80->80/tcp       web-server
def789ghi012   postgres:15              "docker-entrypoint.s…"   3 days ago    Up 3 days    0.0.0.0:5432->5432/tcp   database
ghi345jkl678   redis:7-alpine           "docker-entrypoint.s…"   1 week ago    Up 7 days    0.0.0.0:6379->6379/tcp   cache
jkl901mno234   node:18-alpine           "npm start"              5 days ago    Up 5 days    0.0.0.0:3000->3000/tcp   frontend
mno567pqr890   python:3.11-slim         "python app.py"          1 day ago     Up 1 day     0.0.0.0:8000->8000/tcp   api-server
`

// KubectlGetPodsOutput simulates kubectl get pods output
var KubectlGetPodsOutput = `NAME                          READY   STATUS    RESTARTS   AGE
nginx-deployment-abc123       1/1     Running   0          2d
nginx-def456                  1/1     Running   0          2d
postgres-ghi789               1/1     Running   0          5d
redis-jkl012                  1/1     Running   0          5d
api-server-mno345             1/1     Running   0          1d
worker-pqr678                 0/1     CrashLoopBackOff   5   3h
scheduler-stu901              1/1     Running   0          1d
`

// NpmInstallOutput simulates npm install output
var NpmInstallOutput = `npm WARN deprecated request@2.88.2: request has been deprecated, see https://github.com/request/request/issues/3142
npm WARN deprecated har-validator@5.1.5: this library is no longer supported
npm WARN deprecated uuid@3.4.0: Please upgrade  to version 7 or higher

added 1234 packages, and audited 1235 packages in 15s

152 packages are looking for funding
  run 'npm fund' for details

8 vulnerabilities (2 moderate, 4 high, 2 critical)

To address issues that do not require attention, run:
  npm audit fix

To address all issues (including breaking changes), run:
  npm audit fix --force
`

// PytestOutput simulates pytest output
var PytestOutput = `============================= test session starts ==============================
platform darwin -- Python 3.11.4, pytest-7.4.0, pluggy-1.3.0
rootdir: /Users/dev/project
configfile: pyproject.toml
plugins: anyio-3.7.1, cov-4.1.0
collected 150 items                                                            

tests/test_api.py ....................................................     [33%]
tests/test_auth.py ........................................             [66%]
tests/test_database.py ..............................                   [88%]
tests/test_utils.py ..................                                  [100%]

=============================== warnings summary ===============================
tests/test_api.py::test_endpoint
  /Users/dev/project/tests/test_api.py:45: UserWarning: Deprecated endpoint

-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html
======================== 150 passed, 1 warning in 12.34s ========================
`

// GoTestOutput simulates go test output
var GoTestOutput = `?   	github.com/GrayCodeAI/tokman/internal/cache	[no test files]
ok  	github.com/GrayCodeAI/tokman/internal/commands/analysis	1.234s
ok  	github.com/GrayCodeAI/tokman/internal/commands/filtercmd	2.345s
ok  	github.com/GrayCodeAI/tokman/internal/config	0.456s
ok  	github.com/GrayCodeAI/tokman/internal/core	0.789s
ok  	github.com/GrayCodeAI/tokman/internal/filter	5.678s
ok  	github.com/GrayCodeAI/tokman/internal/tracking	1.234s
PASS
`

// LargeLogFile simulates a large log output (100KB)
var LargeLogFile = generateLargeLog()

func generateLargeLog() string {
	const lines = 2000
	log := ""
	for i := 0; i < lines; i++ {
		log += "2026-03-30 10:30:45 INFO  [main] Processing request from user_id=12345 action=api_call latency=45ms\n"
		log += "2026-03-30 10:30:46 DEBUG [worker] Executing query took 45ms rows=123\n"
		log += "2026-03-30 10:30:47 INFO  [api] Response sent successfully status=200\n"
	}
	return log
}

// AllFixtures returns all CLI output fixtures
func AllFixtures() map[string]string {
	return map[string]string{
		"git_status":   GitStatusOutput,
		"git_log":      GitLogOutput,
		"git_diff":     GitDiffOutput,
		"docker_ps":    DockerPsOutput,
		"kubectl_pods": KubectlGetPodsOutput,
		"npm_install":  NpmInstallOutput,
		"pytest":       PytestOutput,
		"go_test":      GoTestOutput,
		"large_log":    LargeLogFile,
	}
}
