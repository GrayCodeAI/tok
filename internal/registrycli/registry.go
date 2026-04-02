package registrycli

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

type RegistryFilter struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags"`
	Downloads   int       `json:"downloads"`
	Rating      float64   `json:"rating"`
	Installed   bool      `json:"installed"`
	CreatedAt   time.Time `json:"created_at"`
}

type FilterRegistryCLI struct {
	remote map[string]*RegistryFilter
	local  map[string]*RegistryFilter
	mu     sync.RWMutex
}

func NewFilterRegistryCLI() *FilterRegistryCLI {
	reg := &FilterRegistryCLI{
		remote: make(map[string]*RegistryFilter),
		local:  make(map[string]*RegistryFilter),
	}
	reg.seedDefaults()
	return reg
}

func (r *FilterRegistryCLI) seedDefaults() {
	defaults := []RegistryFilter{
		{ID: "git-compressed", Name: "Git Compressed", Author: "tokman", Description: "Compressed git output", Tags: []string{"git"}, Downloads: 1000},
		{ID: "docker-compressed", Name: "Docker Compressed", Author: "tokman", Description: "Compressed docker output", Tags: []string{"docker"}, Downloads: 800},
		{ID: "kubectl-compressed", Name: "Kubectl Compressed", Author: "tokman", Description: "Compressed kubectl output", Tags: []string{"k8s"}, Downloads: 600},
		{ID: "npm-compressed", Name: "NPM Compressed", Author: "tokman", Description: "Compressed npm output", Tags: []string{"npm"}, Downloads: 500},
		{ID: "cargo-compressed", Name: "Cargo Compressed", Author: "tokman", Description: "Compressed cargo output", Tags: []string{"rust"}, Downloads: 400},
		{ID: "go-test-compressed", Name: "Go Test Compressed", Author: "tokman", Description: "Compressed go test output", Tags: []string{"go"}, Downloads: 300},
	}
	for _, f := range defaults {
		r.remote[f.ID] = &f
	}
}

func (r *FilterRegistryCLI) Search(query string) []*RegistryFilter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	query = strings.ToLower(query)
	var results []*RegistryFilter
	for _, f := range r.remote {
		if strings.Contains(strings.ToLower(f.Name), query) ||
			strings.Contains(strings.ToLower(f.Description), query) ||
			strings.Contains(strings.ToLower(f.Author), query) {
			results = append(results, f)
		}
	}
	return results
}

func (r *FilterRegistryCLI) Install(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	f, ok := r.remote[id]
	if !ok {
		return nil
	}
	f.Installed = true
	r.local[id] = f
	f.Downloads++
	return nil
}

func (r *FilterRegistryCLI) Uninstall(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if f, ok := r.local[id]; ok {
		f.Installed = false
		delete(r.local, id)
	}
}

func (r *FilterRegistryCLI) Publish(filter *RegistryFilter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	filter.CreatedAt = time.Now()
	r.remote[filter.ID] = filter
	return nil
}

func (r *FilterRegistryCLI) ListInstalled() []*RegistryFilter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*RegistryFilter
	for _, f := range r.local {
		result = append(result, f)
	}
	return result
}

func (r *FilterRegistryCLI) ListRemote() []*RegistryFilter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*RegistryFilter
	for _, f := range r.remote {
		result = append(result, f)
	}
	return result
}

func (r *FilterRegistryCLI) ExportJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return json.MarshalIndent(r.local, "", "  ")
}

type ShellCompletion struct {
	Shell  string `json:"shell"`
	Script string `json:"script"`
}

func GenerateCompletions() []*ShellCompletion {
	return []*ShellCompletion{
		{Shell: "bash", Script: `complete -W "init filter gateway tui gain" tokman`},
		{Shell: "zsh", Script: `compdef _tokman tokman; _tokman() { _arguments '1:init filter gateway tui gain' }`},
		{Shell: "fish", Script: `complete -c tokman -a "init filter gateway tui gain"`},
		{Shell: "powershell", Script: `Register-ArgumentCompleter -Native -CommandName tokman -ScriptBlock { param($wordToComplete); @("init","filter","gateway","tui","gain") | Where-Object { $_ -like "$wordToComplete*" } }`},
	}
}
