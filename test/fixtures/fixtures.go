package fixtures

import "fmt"

// GitStatusOutput contains sample git status output for benchmarking
var GitStatusOutput = `On branch main
Your branch is up to date with 'origin/main'.

Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   internal/filter/pipeline.go
	modified:   internal/commands/root.go

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	internal/filter/new_layer.go
	docs/NEW_FEATURE.md

no changes added to commit (use "git add" and/or "git commit -a")`

// GitLogOutput contains sample git log output for benchmarking
var GitLogOutput = `commit a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
Author: Developer <dev@example.com>
Date:   Mon Jan 15 10:30:00 2025 +0000

    feat: add new filter layer for semantic compression
    
    This layer implements semantic-aware token pruning
    based on relevance scoring.

commit b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3
Author: Contributor <contrib@example.com>
Date:   Sun Jan 14 15:45:00 2025 +0000

    fix: resolve memory leak in pipeline coordinator
    
    The pipeline coordinator was not properly releasing
    resources when processing large inputs.

commit c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4
Author: Developer <dev@example.com>
Date:   Sat Jan 13 09:15:00 2025 +0000

    refactor: simplify entropy calculation
    
    Optimized the entropy calculation to use SIMD
    instructions for better performance.`

// GitDiffOutput contains sample git diff output for benchmarking
var GitDiffOutput = `diff --git a/internal/filter/pipeline.go b/internal/filter/pipeline.go
index 1234567..abcdefg 100644
--- a/internal/filter/pipeline.go
+++ b/internal/filter/pipeline.go
@@ -10,6 +10,7 @@ import (
 	"strings"
 	"time"
 )
+
 // PipelineCoordinator orchestrates the 31-layer compression pipeline
 type PipelineCoordinator struct {
 	layers []Filter
@@ -25,6 +26,7 @@ type PipelineCoordinator struct {
 	EnableH2O        bool
 	EnableCompaction bool
 	EnableTFIDF      bool
+	EnableAST        bool
 }
 
 func NewPipelineCoordinator(cfg PipelineConfig) *PipelineCoordinator {
-	return &PipelineCoordinator{}
+	pc := &PipelineCoordinator{
+		config: cfg,
+	}
+	pc.initLayers()
+	return pc
 }`

// DockerPsOutput contains sample docker ps output for benchmarking
var DockerPsOutput = `CONTAINER ID   IMAGE                  COMMAND                  CREATED        STATUS        PORTS                    NAMES
a1b2c3d4e5f6   nginx:latest           "/docker-entrypoint.…"   2 hours ago    Up 2 hours    0.0.0.0:80->80/tcp       web-server
b2c3d4e5f6a1   postgres:15            "docker-entrypoint.s…"   3 days ago     Up 3 days     0.0.0.0:5432->5432/tcp   database
c3d4e5f6a1b2   redis:7-alpine         "docker-entrypoint.s…"   3 days ago     Up 3 days     0.0.0.0:6379->6379/tcp   cache
d4e5f6a1b2c3   app:latest             "node server.js"         1 week ago     Up 1 week     0.0.0.0:3000->3000/tcp   api-server`

// KubectlGetPodsOutput contains sample kubectl get pods output for benchmarking
var KubectlGetPodsOutput = `NAME                                READY   STATUS    RESTARTS   AGE
web-deployment-5d8f7b6c4-x2k9m    1/1     Running   0          2d
api-deployment-7f9a8b5c3-j4n7p    1/1     Running   0          2d
redis-master-0                    1/1     Running   0          5d
postgres-0                        1/1     Running   0          5d
celery-worker-6c8d9e4f2-m3q5r     1/1     Running   1          3d
celery-beat-5b7c8d3e1-k2j4n       1/1     Running   0          3d`

// NpmInstallOutput contains sample npm install output for benchmarking
var NpmInstallOutput = `npm WARN deprecated inflight@1.0.6: This module is not supported
npm WARN deprecated glob@7.2.3: Glob versions prior to v9 are no longer supported

added 147 packages, and audited 148 packages in 12s

23 packages are looking for funding
  run ` + "`npm fund`" + ` for details

found 0 vulnerabilities`

// PytestOutput contains sample pytest output for benchmarking
var PytestOutput = `============================= test session starts ==============================
platform linux -- Python 3.11.5, pytest-7.4.3, pluggy-1.3.0
rootdir: /home/user/project
collected 42 items

tests/test_auth.py ........                                              [ 19%]
tests/test_api.py ............                                           [ 47%]
tests/test_models.py ..........                                          [ 71%]
tests/test_utils.py ............                                         [100%]

============================== 42 passed in 3.45s ==============================`

// GoTestOutput contains sample go test output for benchmarking
var GoTestOutput = `=== RUN   TestPipelineCoordinator
--- PASS: TestPipelineCoordinator (0.02s)
=== RUN   TestEntropyFilter
--- PASS: TestEntropyFilter (0.01s)
=== RUN   TestH2OFilter
--- PASS: TestH2OFilter (0.03s)
=== RUN   TestCompaction
--- PASS: TestCompaction (0.05s)
=== RUN   TestPipelineIntegration
--- PASS: TestPipelineIntegration (0.12s)
PASS
ok  	github.com/lakshmanpatel/tok/internal/filter	0.234s`

// LargeLogFile contains a large simulated log file for benchmarking
var LargeLogFile = generateLargeLog()

func generateLargeLog() string {
	var log string
	for i := 0; i < 1000; i++ {
		log += `[2025-01-15T10:30:` + fmt.Sprintf("%02d", i%60) + `Z] INFO server.handler: Request processed method=GET path=/api/v1/users status=200 duration=45ms client=192.168.1.` + fmt.Sprintf("%d", i%255) + "\n"
	}
	return log
}

// AllFixtures returns a map of all fixture names to their content
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
