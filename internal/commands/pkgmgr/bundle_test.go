package pkgmgr

import (
	"strings"
	"testing"
)

func TestFilterBundleInstall_AllCached(t *testing.T) {
	input := `Using bundler 2.5.6
Using rake 13.1.0
Using ast 2.4.2
Using base64 0.2.0
Using minitest 5.22.2
Bundle complete! 85 Gemfile dependencies, 200 gems now installed.
Use 'bundle info [gemname]' to see where a bundled gem is installed.
`
	result := filterBundleInstall(input)
	if !strings.Contains(result, "ok bundle install") {
		t.Errorf("expected 'ok bundle install', got: %s", result)
	}
	if !strings.Contains(result, "all up to date") {
		t.Errorf("expected 'all up to date', got: %s", result)
	}
}

func TestFilterBundleInstall_WithInstalls(t *testing.T) {
	input := `Fetching gem metadata from https://rubygems.org/.........
Resolving dependencies...
Using rake 13.1.0
Using ast 2.4.2
Fetching rspec 3.13.0
Installing rspec 3.13.0
Using rubocop 1.62.0
Fetching simplecov 0.22.0
Installing simplecov 0.22.0
Bundle complete! 85 Gemfile dependencies, 202 gems now installed.
`
	result := filterBundleInstall(input)
	if !strings.Contains(result, "Installing rspec") {
		t.Errorf("expected install line, got: %s", result)
	}
	if !strings.Contains(result, "Installing simplecov") {
		t.Errorf("expected install line, got: %s", result)
	}
	if strings.Contains(result, "Using rake") {
		t.Errorf("should strip Using lines, got: %s", result)
	}
}

func TestFilterBundleInstall_Error(t *testing.T) {
	input := `Fetching gem metadata from https://rubygems.org/.........
Could not find gem 'rails (~> 7.0)' in any of the relevant sources:
  the local rubygems cache
Bundler::GemNotFound
`
	result := filterBundleInstall(input)
	if !strings.Contains(result, "FAILED") {
		t.Errorf("expected FAILED, got: %s", result)
	}
	if !strings.Contains(result, "Could not find gem") {
		t.Errorf("expected error message, got: %s", result)
	}
}

func TestFilterBundleInstall_WithRemovals(t *testing.T) {
	input := `Using rake 13.1.0
Removing old-rspec 3.12.0
Installing rspec 3.13.0
Bundle complete! 85 Gemfile dependencies, 200 gems now installed.
`
	result := filterBundleInstall(input)
	if !strings.Contains(result, "Removing old-rspec") {
		t.Errorf("expected removal line, got: %s", result)
	}
}

func TestFilterBundleUpdate_NoChanges(t *testing.T) {
	input := `Using bundler 2.5.6
Using rake 13.1.0
Using rails 7.1.0
`
	result := filterBundleUpdate(input)
	if !strings.Contains(result, "ok bundle update") {
		t.Errorf("expected 'ok bundle update', got: %s", result)
	}
	if !strings.Contains(result, "no changes") {
		t.Errorf("expected 'no changes', got: %s", result)
	}
}

func TestFilterBundleUpdate_WithChanges(t *testing.T) {
	input := `Fetching gem metadata from https://rubygems.org/.........
Resolving dependencies...
Using rake 13.1.0
Fetching rspec 3.14.0 (was 3.13.0)
Installing rspec 3.14.0 (was 3.13.0)
Bundle updated!
`
	result := filterBundleUpdate(input)
	if !strings.Contains(result, "bundle update:") {
		t.Errorf("expected 'bundle update:', got: %s", result)
	}
	if !strings.Contains(result, "Installing rspec") {
		t.Errorf("expected install line, got: %s", result)
	}
}

func TestFilterBundleList(t *testing.T) {
	input := `Gems included by the bundle:
  * actioncable (7.1.0)
  * actionmailer (7.1.0)
  * actionpack (7.1.0)
  * actionview (7.1.0)
  * activejob (7.1.0)
  * activemodel (7.1.0)
  * activerecord (7.1.0)
  * railties (7.1.0)
Use ` + "`bundle info [gemname]`" + ` to see where a bundled gem is installed.
`
	result := filterBundleList(input)
	if !strings.Contains(result, "bundle list:") {
		t.Errorf("expected 'bundle list:', got: %s", result)
	}
	if !strings.Contains(result, "8 gems") {
		t.Errorf("expected '8 gems', got: %s", result)
	}
}

func TestFilterBundleList_Overflow(t *testing.T) {
	var gems []string
	for i := 0; i < 25; i++ {
		gems = append(gems, "* gem"+string(rune('a'+i%26))+" (1.0.0)")
	}
	input := strings.Join(gems, "\n")

	result := filterBundleList(input)
	if !strings.Contains(result, "... +5 more") {
		t.Errorf("expected overflow message, got: %s", result)
	}
}

func TestFilterBundleOutdated(t *testing.T) {
	input := `Fetching gem metadata from https://rubygems.org/.........
Resolving dependencies...

Outdated gems included in the bundle:
  * rails (newest 7.1.3, installed 7.1.0)
  * rake (newest 13.2.0, installed 13.1.0)
`
	result := filterBundleOutdated(input)
	if !strings.Contains(result, "bundle outdated:") {
		t.Errorf("expected 'bundle outdated:', got: %s", result)
	}
	if !strings.Contains(result, "2 gems") {
		t.Errorf("expected '2 gems', got: %s", result)
	}
}

func TestFilterBundleOutdated_AllUpToDate(t *testing.T) {
	input := `Fetching gem metadata from https://rubygems.org/.........
Resolving dependencies...
`
	result := filterBundleOutdated(input)
	if !strings.Contains(result, "ok bundle outdated") {
		t.Errorf("expected 'ok bundle outdated', got: %s", result)
	}
}
