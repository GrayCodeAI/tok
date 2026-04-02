package filterregistry

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type FilterEntry struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags"`
	Rating      float64   `json:"rating"`
	Downloads   int       `json:"downloads"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FilterRegistry struct {
	filters map[string]*FilterEntry
	local   map[string]*FilterEntry
}

func NewFilterRegistry() *FilterRegistry {
	return &FilterRegistry{
		filters: make(map[string]*FilterEntry),
		local:   make(map[string]*FilterEntry),
	}
}

func (r *FilterRegistry) Publish(filter *FilterEntry) error {
	if filter.ID == "" {
		return fmt.Errorf("filter ID is required")
	}
	if filter.Name == "" {
		return fmt.Errorf("filter name is required")
	}
	if filter.Content == "" {
		return fmt.Errorf("filter content is required")
	}

	safetyErrors := SafetyCheck(filter.Content)
	if len(safetyErrors) > 0 {
		return fmt.Errorf("safety check failed: %s", strings.Join(safetyErrors, ", "))
	}

	filter.CreatedAt = time.Now()
	filter.UpdatedAt = time.Now()
	r.filters[filter.ID] = filter
	return nil
}

func (r *FilterRegistry) Install(id string) (*FilterEntry, error) {
	if filter, ok := r.filters[id]; ok {
		r.local[id] = filter
		filter.Downloads++
		return filter, nil
	}
	return nil, fmt.Errorf("filter not found: %s", id)
}

func (r *FilterRegistry) Search(query string) []*FilterEntry {
	var results []*FilterEntry
	query = strings.ToLower(query)
	for _, f := range r.filters {
		if strings.Contains(strings.ToLower(f.Name), query) ||
			strings.Contains(strings.ToLower(f.Description), query) ||
			strings.Contains(strings.ToLower(f.Author), query) {
			results = append(results, f)
		}
		for _, tag := range f.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, f)
				break
			}
		}
	}
	return results
}

func (r *FilterRegistry) List() []*FilterEntry {
	var result []*FilterEntry
	for _, f := range r.filters {
		result = append(result, f)
	}
	return result
}

func (r *FilterRegistry) GetLocal() map[string]*FilterEntry {
	return r.local
}

func (r *FilterRegistry) Export() ([]byte, error) {
	return json.Marshal(r.filters)
}

func SafetyCheck(content string) []string {
	var errors []string
	dangerousPatterns := []string{
		"os.execute", "io.popen", "os.remove", "os.rename",
		"rm -rf", "curl | sh", "wget | bash",
		"<script>", "javascript:", "eval(",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(content, pattern) {
			errors = append(errors, fmt.Sprintf("dangerous pattern detected: %s", pattern))
		}
	}

	if containsHiddenUnicode(content) {
		errors = append(errors, "hidden unicode characters detected")
	}

	return errors
}

func containsHiddenUnicode(content string) bool {
	for _, r := range content {
		if (r >= 0x200B && r <= 0x200F) || (r >= 0x202A && r <= 0x202E) || r == 0xFEFF {
			return true
		}
	}
	return false
}
