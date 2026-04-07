package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FilterRegistry struct {
	client   *http.Client
	cacheDir string
	builtin  map[string]*FilterMeta
}

type FilterMeta struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Downloads   int64    `json:"downloads"`
	Rating      float64  `json:"rating"`
	Updated     string   `json:"updated"`
}

type SearchOptions struct {
	Query     string
	Category  string
	Tags      []string
	MinRating float64
	Limit     int
}

type InstalledFilter struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	InstalledAt time.Time `json:"installed_at"`
	Source      string    `json:"source"`
}

type FilterStore struct {
	Installed map[string]InstalledFilter `json:"installed"`
	Ratings   map[string]float64         `json:"ratings"`
}

func NewFilterRegistry(cacheDir string) *FilterRegistry {
	return &FilterRegistry{
		client:   &http.Client{Timeout: 10 * time.Second},
		cacheDir: cacheDir,
		builtin: map[string]*FilterMeta{
			"jest":          {Name: "jest", Description: "Jest test runner output filter", Category: "testing", Tags: []string{"javascript", "testing", "jest"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 1000, Rating: 4.5},
			"vitest":        {Name: "vitest", Description: "Vitest test runner output filter", Category: "testing", Tags: []string{"javascript", "testing", "vite"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 800, Rating: 4.3},
			"playwright":    {Name: "playwright", Description: "Playwright E2E test output filter", Category: "testing", Tags: []string{"javascript", "testing", "e2e"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 600, Rating: 4.7},
			"eslint":        {Name: "eslint", Description: "ESLint output filter", Category: "linting", Tags: []string{"javascript", "linting", "es"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 900, Rating: 4.2},
			"biome":         {Name: "biome", Description: "Biome linter output filter", Category: "linting", Tags: []string{"javascript", "linting", "biome"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 400, Rating: 4.6},
			"pytest":        {Name: "pytest", Description: "Pytest output filter", Category: "testing", Tags: []string{"python", "testing", "pytest"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 1200, Rating: 4.8},
			"mypy":          {Name: "mypy", Description: "Mypy type checker output filter", Category: "linting", Tags: []string{"python", "typing", "mypy"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 700, Rating: 4.4},
			"ruff":          {Name: "ruff", Description: "Ruff linter output filter", Category: "linting", Tags: []string{"python", "linting", "fast"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 500, Rating: 4.5},
			"cargo":         {Name: "cargo", Description: "Cargo build output filter", Category: "build", Tags: []string{"rust", "build", "cargo"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 1100, Rating: 4.6},
			"go":            {Name: "go", Description: "Go build/test output filter", Category: "build", Tags: []string{"golang", "build", "go"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 1300, Rating: 4.7},
			"docker":        {Name: "docker", Description: "Docker CLI output filter", Category: "container", Tags: []string{"docker", "container", "cli"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 800, Rating: 4.3},
			"kubectl":       {Name: "kubectl", Description: "Kubernetes CLI output filter", Category: "container", Tags: []string{"kubernetes", "k8s", "cli"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 700, Rating: 4.4},
			"terraform":     {Name: "terraform", Description: "Terraform output filter", Category: "infra", Tags: []string{"terraform", "infra", "iac"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 600, Rating: 4.2},
			"helm":          {Name: "helm", Description: "Helm output filter", Category: "infra", Tags: []string{"helm", "k8s", "charts"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 400, Rating: 4.1},
			"npm":           {Name: "npm", Description: "npm output filter", Category: "package", Tags: []string{"npm", "node", "package"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 900, Rating: 4.3},
			"pnpm":          {Name: "pnpm", Description: "pnpm output filter", Category: "package", Tags: []string{"pnpm", "node", "package"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 500, Rating: 4.5},
			"webpack":       {Name: "webpack", Description: "Webpack build output filter", Category: "build", Tags: []string{"javascript", "build", "webpack"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 600, Rating: 4.0},
			"vite":          {Name: "vite", Description: "Vite build output filter", Category: "build", Tags: []string{"javascript", "build", "vite"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 700, Rating: 4.4},
			"trivy":         {Name: "trivy", Description: "Trivy security scanner output filter", Category: "security", Tags: []string{"security", "scanner", "trivy"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 500, Rating: 4.6},
			"golangci-lint": {Name: "golangci-lint", Description: "golangci-lint output filter", Category: "linting", Tags: []string{"golang", "linting", "ci"}, Version: "1.0.0", Author: "GrayCodeAI", Downloads: 800, Rating: 4.5},
		},
	}
}

func (r *FilterRegistry) Search(ctx context.Context, opts SearchOptions) ([]*FilterMeta, error) {
	results := make([]*FilterMeta, 0)

	for _, filter := range r.builtin {
		if opts.Query != "" && !strings.Contains(strings.ToLower(filter.Name), strings.ToLower(opts.Query)) &&
			!strings.Contains(strings.ToLower(filter.Description), strings.ToLower(opts.Query)) {
			continue
		}

		if opts.Category != "" && filter.Category != opts.Category {
			continue
		}

		if len(opts.Tags) > 0 {
			hasTag := false
			for _, tag := range opts.Tags {
				for _, ftag := range filter.Tags {
					if strings.ToLower(tag) == strings.ToLower(ftag) {
						hasTag = true
						break
					}
				}
			}
			if !hasTag {
				continue
			}
		}

		if opts.MinRating > 0 && filter.Rating < opts.MinRating {
			continue
		}

		results = append(results, filter)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Downloads > results[j].Downloads
	})

	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

func (r *FilterRegistry) GetCategories() []string {
	catSet := make(map[string]bool)
	for _, f := range r.builtin {
		catSet[f.Category] = true
	}
	cats := make([]string, 0, len(catSet))
	for c := range catSet {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	return cats
}

func (r *FilterRegistry) GetRecommendations(ctx context.Context, installed []string) ([]*FilterMeta, error) {
	installedSet := make(map[string]bool)
	for _, name := range installed {
		installedSet[name] = true
	}

	var installedCats []string
	for name := range installedSet {
		if f, ok := r.builtin[name]; ok {
			installedCats = append(installedCats, f.Category)
		}
	}

	catCount := make(map[string]int)
	for _, c := range installedCats {
		catCount[c]++
	}

	var recs []*FilterMeta
	for _, f := range r.builtin {
		if installedSet[f.Name] {
			continue
		}
		if catCount[f.Category] > 0 {
			recs = append(recs, f)
		}
	}

	sort.Slice(recs, func(i, j int) bool {
		return recs[i].Rating > recs[j].Rating
	})

	if len(recs) > 10 {
		recs = recs[:10]
	}

	return recs, nil
}

func (r *FilterRegistry) GetTrending(ctx context.Context) ([]*FilterMeta, error) {
	var trending []*FilterMeta
	for _, f := range r.builtin {
		trending = append(trending, f)
	}

	sort.Slice(trending, func(i, j int) bool {
		return trending[i].Downloads > trending[j].Downloads
	})

	return trending[:5], nil
}

func (r *FilterRegistry) DownloadFilter(ctx context.Context, name string) ([]byte, error) {
	meta, ok := r.builtin[name]
	if !ok {
		return nil, fmt.Errorf("filter not found: %s", name)
	}

	resp, err := r.client.Get(meta.URL)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (r *FilterRegistry) LoadStore() (*FilterStore, error) {
	storePath := filepath.Join(r.cacheDir, "store.json")
	data, err := os.ReadFile(storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &FilterStore{Installed: make(map[string]InstalledFilter), Ratings: make(map[string]float64)}, nil
		}
		return nil, err
	}

	var store FilterStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *FilterRegistry) SaveStore(store *FilterStore) error {
	storePath := filepath.Join(r.cacheDir, "store.json")
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(storePath, data, 0600)
}

func (r *FilterRegistry) RateFilter(name string, rating float64) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	store, err := r.LoadStore()
	if err != nil {
		return err
	}

	store.Ratings[name] = rating
	return r.SaveStore(store)
}

func (r *FilterRegistry) GetRating(name string) float64 {
	store, err := r.LoadStore()
	if err != nil {
		return 0
	}
	return store.Ratings[name]
}
