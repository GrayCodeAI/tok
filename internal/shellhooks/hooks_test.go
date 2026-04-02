package shellhooks

import "testing"

func TestShellHookRegistry(t *testing.T) {
	reg := NewShellHookRegistry()
	if reg.Count() < 90 {
		t.Errorf("Expected at least 90 hooks, got %d", reg.Count())
	}

	hook := reg.Get("git_status")
	if hook == nil {
		t.Error("Expected git_status hook")
	}

	gitHooks := reg.GetByCategory(HookGit)
	if len(gitHooks) < 10 {
		t.Errorf("Expected at least 10 git hooks, got %d", len(gitHooks))
	}

	all := reg.List()
	if len(all) < 90 {
		t.Errorf("Expected at least 90 hooks in list, got %d", len(all))
	}
}
