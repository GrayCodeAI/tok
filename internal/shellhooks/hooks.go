package shellhooks

type HookCategory string

const (
	HookGit      HookCategory = "git"
	HookBuild    HookCategory = "build"
	HookTest     HookCategory = "test"
	HookLint     HookCategory = "lint"
	HookDocker   HookCategory = "docker"
	HookK8s      HookCategory = "k8s"
	HookAWS      HookCategory = "aws"
	HookSSH      HookCategory = "ssh"
	HookHTTP     HookCategory = "http"
	HookFile     HookCategory = "file"
	HookProcess  HookCategory = "process"
	HookNetwork  HookCategory = "network"
	HookPackage  HookCategory = "package"
	HookMonitor  HookCategory = "monitor"
	HookDeploy   HookCategory = "deploy"
	HookDebug    HookCategory = "debug"
	HookSecurity HookCategory = "security"
	HookData     HookCategory = "data"
)

type ShellHook struct {
	ID          string       `json:"id"`
	Command     string       `json:"command"`
	Regex       string       `json:"regex"`
	Category    HookCategory `json:"category"`
	FilterFunc  string       `json:"filter_func"`
	TokensSaved int          `json:"tokens_saved"`
}

type ShellHookRegistry struct {
	hooks map[string]*ShellHook
}

func NewShellHookRegistry() *ShellHookRegistry {
	reg := &ShellHookRegistry{
		hooks: make(map[string]*ShellHook),
	}
	reg.registerBuiltInHooks()
	return reg
}

func (r *ShellHookRegistry) registerBuiltInHooks() {
	hooks := []ShellHook{
		// Git hooks (15)
		{ID: "git_status", Command: "git status", Regex: "git status", Category: HookGit, FilterFunc: "git_status", TokensSaved: 50},
		{ID: "git_log", Command: "git log", Regex: "git log", Category: HookGit, FilterFunc: "git_log", TokensSaved: 200},
		{ID: "git_diff", Command: "git diff", Regex: "git diff", Category: HookGit, FilterFunc: "git_diff", TokensSaved: 100},
		{ID: "git_branch", Command: "git branch", Regex: "git branch", Category: HookGit, FilterFunc: "git_branch", TokensSaved: 20},
		{ID: "git_remote", Command: "git remote", Regex: "git remote", Category: HookGit, FilterFunc: "git_remote", TokensSaved: 10},
		{ID: "git_show", Command: "git show", Regex: "git show", Category: HookGit, FilterFunc: "git_show", TokensSaved: 80},
		{ID: "git_reflog", Command: "git reflog", Regex: "git reflog", Category: HookGit, FilterFunc: "git_reflog", TokensSaved: 100},
		{ID: "git_stash", Command: "git stash", Regex: "git stash", Category: HookGit, FilterFunc: "git_stash", TokensSaved: 30},
		{ID: "git_tag", Command: "git tag", Regex: "git tag", Category: HookGit, FilterFunc: "git_tag", TokensSaved: 10},
		{ID: "git_bisect", Command: "git bisect", Regex: "git bisect", Category: HookGit, FilterFunc: "git_bisect", TokensSaved: 40},
		{ID: "git_blame", Command: "git blame", Regex: "git blame", Category: HookGit, FilterFunc: "git_blame", TokensSaved: 100},
		{ID: "git_fetch", Command: "git fetch", Regex: "git fetch", Category: HookGit, FilterFunc: "git_fetch", TokensSaved: 50},
		{ID: "git_merge", Command: "git merge", Regex: "git merge", Category: HookGit, FilterFunc: "git_merge", TokensSaved: 30},
		{ID: "git_rebase", Command: "git rebase", Regex: "git rebase", Category: HookGit, FilterFunc: "git_rebase", TokensSaved: 40},
		{ID: "git_cherry_pick", Command: "git cherry-pick", Regex: "git cherry-pick", Category: HookGit, FilterFunc: "git_cherry_pick", TokensSaved: 30},

		// Build hooks (10)
		{ID: "cargo_build", Command: "cargo build", Regex: "cargo build", Category: HookBuild, FilterFunc: "cargo_build", TokensSaved: 100},
		{ID: "make_build", Command: "make", Regex: "^make", Category: HookBuild, FilterFunc: "make_build", TokensSaved: 80},
		{ID: "gradle_build", Command: "gradle build", Regex: "gradle build", Category: HookBuild, FilterFunc: "gradle_build", TokensSaved: 100},
		{ID: "mvn_build", Command: "mvn build", Regex: "mvn", Category: HookBuild, FilterFunc: "mvn_build", TokensSaved: 100},
		{ID: "npm_build", Command: "npm run build", Regex: "npm run build", Category: HookBuild, FilterFunc: "npm_build", TokensSaved: 80},
		{ID: "yarn_build", Command: "yarn build", Regex: "yarn build", Category: HookBuild, FilterFunc: "yarn_build", TokensSaved: 80},
		{ID: "pnpm_build", Command: "pnpm build", Regex: "pnpm build", Category: HookBuild, FilterFunc: "pnpm_build", TokensSaved: 80},
		{ID: "go_build", Command: "go build", Regex: "go build", Category: HookBuild, FilterFunc: "go_build", TokensSaved: 60},
		{ID: "python_build", Command: "python setup.py", Regex: "python setup.py", Category: HookBuild, FilterFunc: "python_build", TokensSaved: 50},
		{ID: "ruby_build", Command: "bundle install", Regex: "bundle install", Category: HookBuild, FilterFunc: "bundle_install", TokensSaved: 60},

		// Test hooks (10)
		{ID: "cargo_test", Command: "cargo test", Regex: "cargo test", Category: HookTest, FilterFunc: "cargo_test", TokensSaved: 150},
		{ID: "jest_test", Command: "jest", Regex: "^jest|npx jest", Category: HookTest, FilterFunc: "jest_test", TokensSaved: 100},
		{ID: "pytest", Command: "pytest", Regex: "^pytest", Category: HookTest, FilterFunc: "pytest_test", TokensSaved: 100},
		{ID: "vitest", Command: "vitest", Regex: "^vitest|npx vitest", Category: HookTest, FilterFunc: "vitest_test", TokensSaved: 100},
		{ID: "go_test", Command: "go test", Regex: "go test", Category: HookTest, FilterFunc: "go_test", TokensSaved: 100},
		{ID: "rspec", Command: "rspec", Regex: "^rspec", Category: HookTest, FilterFunc: "rspec_test", TokensSaved: 100},
		{ID: "playwright", Command: "playwright test", Regex: "playwright test", Category: HookTest, FilterFunc: "playwright_test", TokensSaved: 80},
		{ID: "mocha", Command: "mocha", Regex: "^mocha|npx mocha", Category: HookTest, FilterFunc: "mocha_test", TokensSaved: 100},
		{ID: "cypress", Command: "cypress", Regex: "^cypress|npx cypress", Category: HookTest, FilterFunc: "cypress_test", TokensSaved: 80},
		{ID: "junit", Command: "junit", Regex: "junit", Category: HookTest, FilterFunc: "junit_test", TokensSaved: 100},

		// Lint hooks (8)
		{ID: "eslint", Command: "eslint", Regex: "^eslint|npx eslint", Category: HookLint, FilterFunc: "eslint", TokensSaved: 80},
		{ID: "prettier", Command: "prettier", Regex: "^prettier|npx prettier", Category: HookLint, FilterFunc: "prettier", TokensSaved: 50},
		{ID: "ruff", Command: "ruff", Regex: "^ruff", Category: HookLint, FilterFunc: "ruff", TokensSaved: 60},
		{ID: "tsc", Command: "tsc", Regex: "^tsc|npx tsc", Category: HookLint, FilterFunc: "tsc", TokensSaved: 100},
		{ID: "golangci", Command: "golangci-lint", Regex: "golangci-lint", Category: HookLint, FilterFunc: "golangci", TokensSaved: 80},
		{ID: "rubocop", Command: "rubocop", Regex: "^rubocop", Category: HookLint, FilterFunc: "rubocop", TokensSaved: 80},
		{ID: "mypy", Command: "mypy", Regex: "^mypy", Category: HookLint, FilterFunc: "mypy", TokensSaved: 80},
		{ID: "shellcheck", Command: "shellcheck", Regex: "^shellcheck", Category: HookLint, FilterFunc: "shellcheck", TokensSaved: 50},

		// Docker hooks (8)
		{ID: "docker_ps", Command: "docker ps", Regex: "docker ps", Category: HookDocker, FilterFunc: "docker_ps", TokensSaved: 50},
		{ID: "docker_images", Command: "docker images", Regex: "docker images", Category: HookDocker, FilterFunc: "docker_images", TokensSaved: 50},
		{ID: "docker_logs", Command: "docker logs", Regex: "docker logs", Category: HookDocker, FilterFunc: "docker_logs", TokensSaved: 100},
		{ID: "docker_compose", Command: "docker compose", Regex: "docker compose", Category: HookDocker, FilterFunc: "docker_compose", TokensSaved: 80},
		{ID: "docker_build", Command: "docker build", Regex: "docker build", Category: HookDocker, FilterFunc: "docker_build", TokensSaved: 60},
		{ID: "docker_exec", Command: "docker exec", Regex: "docker exec", Category: HookDocker, FilterFunc: "docker_exec", TokensSaved: 30},
		{ID: "docker_network", Command: "docker network", Regex: "docker network", Category: HookDocker, FilterFunc: "docker_network", TokensSaved: 40},
		{ID: "docker_volume", Command: "docker volume", Regex: "docker volume", Category: HookDocker, FilterFunc: "docker_volume", TokensSaved: 40},

		// K8s hooks (8)
		{ID: "kubectl_pods", Command: "kubectl get pods", Regex: "kubectl get pods", Category: HookK8s, FilterFunc: "kubectl_pods", TokensSaved: 100},
		{ID: "kubectl_services", Command: "kubectl get services", Regex: "kubectl get services", Category: HookK8s, FilterFunc: "kubectl_services", TokensSaved: 80},
		{ID: "kubectl_logs", Command: "kubectl logs", Regex: "kubectl logs", Category: HookK8s, FilterFunc: "kubectl_logs", TokensSaved: 150},
		{ID: "kubectl_describe", Command: "kubectl describe", Regex: "kubectl describe", Category: HookK8s, FilterFunc: "kubectl_describe", TokensSaved: 100},
		{ID: "kubectl_get", Command: "kubectl get", Regex: "kubectl get", Category: HookK8s, FilterFunc: "kubectl_get", TokensSaved: 80},
		{ID: "kubectl_apply", Command: "kubectl apply", Regex: "kubectl apply", Category: HookK8s, FilterFunc: "kubectl_apply", TokensSaved: 60},
		{ID: "kubectl_rollout", Command: "kubectl rollout", Regex: "kubectl rollout", Category: HookK8s, FilterFunc: "kubectl_rollout", TokensSaved: 50},
		{ID: "kubectl_exec", Command: "kubectl exec", Regex: "kubectl exec", Category: HookK8s, FilterFunc: "kubectl_exec", TokensSaved: 30},

		// HTTP hooks (8)
		{ID: "curl", Command: "curl", Regex: "^curl", Category: HookHTTP, FilterFunc: "curl", TokensSaved: 80},
		{ID: "wget", Command: "wget", Regex: "^wget", Category: HookHTTP, FilterFunc: "wget", TokensSaved: 60},
		{ID: "httpie", Command: "http", Regex: "^http", Category: HookHTTP, FilterFunc: "httpie", TokensSaved: 60},
		{ID: "gh_api", Command: "gh api", Regex: "gh api", Category: HookHTTP, FilterFunc: "gh_api", TokensSaved: 100},
		{ID: "gh_pr", Command: "gh pr", Regex: "gh pr", Category: HookHTTP, FilterFunc: "gh_pr", TokensSaved: 80},
		{ID: "gh_issue", Command: "gh issue", Regex: "gh issue", Category: HookHTTP, FilterFunc: "gh_issue", TokensSaved: 80},
		{ID: "gh_run", Command: "gh run", Regex: "gh run", Category: HookHTTP, FilterFunc: "gh_run", TokensSaved: 60},
		{ID: "gh_actions", Command: "gh workflow", Regex: "gh workflow", Category: HookHTTP, FilterFunc: "gh_actions", TokensSaved: 60},

		// File hooks (8)
		{ID: "ls", Command: "ls", Regex: "^ls", Category: HookFile, FilterFunc: "ls", TokensSaved: 30},
		{ID: "find", Command: "find", Regex: "^find", Category: HookFile, FilterFunc: "find", TokensSaved: 50},
		{ID: "grep", Command: "grep", Regex: "^grep", Category: HookFile, FilterFunc: "grep", TokensSaved: 80},
		{ID: "rg", Command: "rg", Regex: "^rg", Category: HookFile, FilterFunc: "rg", TokensSaved: 80},
		{ID: "tree", Command: "tree", Regex: "^tree", Category: HookFile, FilterFunc: "tree", TokensSaved: 40},
		{ID: "diff", Command: "diff", Regex: "^diff", Category: HookFile, FilterFunc: "diff", TokensSaved: 60},
		{ID: "sort", Command: "sort", Regex: "^sort", Category: HookFile, FilterFunc: "sort", TokensSaved: 30},
		{ID: "wc", Command: "wc", Regex: "^wc", Category: HookFile, FilterFunc: "wc", TokensSaved: 10},

		// Monitor hooks (8)
		{ID: "ps", Command: "ps", Regex: "^ps", Category: HookMonitor, FilterFunc: "ps", TokensSaved: 40},
		{ID: "top", Command: "top", Regex: "^top", Category: HookMonitor, FilterFunc: "top", TokensSaved: 60},
		{ID: "df", Command: "df", Regex: "^df", Category: HookMonitor, FilterFunc: "df", TokensSaved: 20},
		{ID: "du", Command: "du", Regex: "^du", Category: HookMonitor, FilterFunc: "du", TokensSaved: 30},
		{ID: "free", Command: "free", Regex: "^free", Category: HookMonitor, FilterFunc: "free", TokensSaved: 10},
		{ID: "netstat", Command: "netstat", Regex: "^netstat", Category: HookMonitor, FilterFunc: "netstat", TokensSaved: 50},
		{ID: "ss", Command: "ss", Regex: "^ss", Category: HookMonitor, FilterFunc: "ss", TokensSaved: 50},
		{ID: "systemctl", Command: "systemctl", Regex: "^systemctl", Category: HookMonitor, FilterFunc: "systemctl", TokensSaved: 40},

		// Package hooks (7)
		{ID: "npm_list", Command: "npm list", Regex: "npm list", Category: HookPackage, FilterFunc: "npm_list", TokensSaved: 60},
		{ID: "pip_list", Command: "pip list", Regex: "pip list", Category: HookPackage, FilterFunc: "pip_list", TokensSaved: 40},
		{ID: "pip_outdated", Command: "pip list --outdated", Regex: "pip list.*outdated", Category: HookPackage, FilterFunc: "pip_outdated", TokensSaved: 40},
		{ID: "cargo_list", Command: "cargo tree", Regex: "cargo tree", Category: HookPackage, FilterFunc: "cargo_tree", TokensSaved: 80},
		{ID: "go_mod", Command: "go mod", Regex: "go mod", Category: HookPackage, FilterFunc: "go_mod", TokensSaved: 30},
		{ID: "composer", Command: "composer show", Regex: "composer show", Category: HookPackage, FilterFunc: "composer", TokensSaved: 40},
		{ID: "gem_list", Command: "gem list", Regex: "gem list", Category: HookPackage, FilterFunc: "gem_list", TokensSaved: 30},

		// Infra hooks (8)
		{ID: "terraform_plan", Command: "terraform plan", Regex: "terraform plan", Category: HookDeploy, FilterFunc: "terraform_plan", TokensSaved: 200},
		{ID: "terraform_apply", Command: "terraform apply", Regex: "terraform apply", Category: HookDeploy, FilterFunc: "terraform_apply", TokensSaved: 150},
		{ID: "helm_ls", Command: "helm list", Regex: "helm list", Category: HookDeploy, FilterFunc: "helm_ls", TokensSaved: 60},
		{ID: "helm_status", Command: "helm status", Regex: "helm status", Category: HookDeploy, FilterFunc: "helm_status", TokensSaved: 80},
		{ID: "ansible", Command: "ansible", Regex: "^ansible", Category: HookDeploy, FilterFunc: "ansible", TokensSaved: 100},
		{ID: "aws_s3", Command: "aws s3", Regex: "aws s3", Category: HookDeploy, FilterFunc: "aws_s3", TokensSaved: 40},
		{ID: "aws_ec2", Command: "aws ec2", Regex: "aws ec2", Category: HookDeploy, FilterFunc: "aws_ec2", TokensSaved: 100},
		{ID: "gcloud", Command: "gcloud", Regex: "^gcloud", Category: HookDeploy, FilterFunc: "gcloud", TokensSaved: 80},
	}

	for i := range hooks {
		r.hooks[hooks[i].ID] = &hooks[i]
	}
}

func (r *ShellHookRegistry) Get(command string) *ShellHook {
	if h, ok := r.hooks[command]; ok {
		return h
	}
	return nil
}

func (r *ShellHookRegistry) List() []*ShellHook {
	var result []*ShellHook
	for _, h := range r.hooks {
		result = append(result, h)
	}
	return result
}

func (r *ShellHookRegistry) GetByCategory(cat HookCategory) []*ShellHook {
	var result []*ShellHook
	for _, h := range r.hooks {
		if h.Category == cat {
			result = append(result, h)
		}
	}
	return result
}

func (r *ShellHookRegistry) Count() int {
	return len(r.hooks)
}
