package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportFormat represents the export format type
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatTAR  ExportFormat = "tar"
	ExportFormatZIP  ExportFormat = "zip"
)

// Exporter handles archive export operations
type Exporter struct {
	manager *ArchiveManager
}

// NewExporter creates a new exporter
func NewExporter(manager *ArchiveManager) *Exporter {
	return &Exporter{manager: manager}
}

// ExportOptions contains export configuration
type ExportOptions struct {
	Format      ExportFormat
	Compression bool
	IncludeRaw  bool
}

// DefaultExportOptions returns default export options
func DefaultExportOptions() ExportOptions {
	return ExportOptions{
		Format:      ExportFormatJSON,
		Compression: true,
		IncludeRaw:  true,
	}
}

// Export exports a single archive by hash
func (e *Exporter) Export(ctx context.Context, hash string, opts ExportOptions) ([]byte, string, error) {
	entry, err := e.manager.Retrieve(ctx, hash)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve archive: %w", err)
	}

	switch opts.Format {
	case ExportFormatJSON:
		return e.exportJSON(entry, opts)
	case ExportFormatTAR:
		return e.exportTAR([]*ArchiveEntry{entry}, opts)
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

// ExportMultiple exports multiple archives
func (e *Exporter) ExportMultiple(ctx context.Context, hashes []string, opts ExportOptions) ([]byte, string, error) {
	var entries []*ArchiveEntry

	for _, hash := range hashes {
		entry, err := e.manager.Retrieve(ctx, hash)
		if err != nil {
			return nil, "", fmt.Errorf("failed to retrieve archive %s: %w", hash, err)
		}
		entries = append(entries, entry)
	}

	switch opts.Format {
	case ExportFormatJSON:
		return e.exportJSONMultiple(entries, opts)
	case ExportFormatTAR:
		return e.exportTAR(entries, opts)
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

// ExportAll exports all archives matching filters
func (e *Exporter) ExportAll(ctx context.Context, opts ExportOptions, listOpts ArchiveListOptions) ([]byte, string, error) {
	result, err := e.manager.List(ctx, listOpts)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list archives: %w", err)
	}

	var entries []*ArchiveEntry
	for i := range result.Entries {
		entries = append(entries, &result.Entries[i])
	}

	switch opts.Format {
	case ExportFormatJSON:
		return e.exportJSONMultiple(entries, opts)
	case ExportFormatTAR:
		return e.exportTAR(entries, opts)
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

func (e *Exporter) exportJSON(entry *ArchiveEntry, opts ExportOptions) ([]byte, string, error) {
	data := map[string]interface{}{
		"hash":              entry.Hash,
		"command":           entry.Command,
		"working_directory": entry.WorkingDirectory,
		"category":          entry.Category,
		"agent":             entry.Agent,
		"created_at":        entry.CreatedAt,
		"tags":              entry.Tags,
		"metadata":          entry.Metadata,
	}

	if opts.IncludeRaw {
		data["original_content"] = string(entry.OriginalContent)
		if entry.FilteredContent != nil {
			data["filtered_content"] = string(entry.FilteredContent)
		}
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	filename := fmt.Sprintf("archive_%s.json", entry.Hash[:16])

	if opts.Compression {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, err := gw.Write(jsonData); err != nil {
			return nil, "", err
		}
		gw.Close()
		return buf.Bytes(), filename + ".gz", nil
	}

	return jsonData, filename, nil
}

func (e *Exporter) exportJSONMultiple(entries []*ArchiveEntry, opts ExportOptions) ([]byte, string, error) {
	var data []map[string]interface{}

	for _, entry := range entries {
		entryData := map[string]interface{}{
			"hash":              entry.Hash,
			"command":           entry.Command,
			"working_directory": entry.WorkingDirectory,
			"category":          entry.Category,
			"agent":             entry.Agent,
			"created_at":        entry.CreatedAt,
			"tags":              entry.Tags,
			"metadata":          entry.Metadata,
		}

		if opts.IncludeRaw {
			entryData["original_content"] = string(entry.OriginalContent)
			if entry.FilteredContent != nil {
				entryData["filtered_content"] = string(entry.FilteredContent)
			}
		}

		data = append(data, entryData)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	filename := fmt.Sprintf("archives_%d.json", len(entries))

	if opts.Compression {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, err := gw.Write(jsonData); err != nil {
			return nil, "", err
		}
		gw.Close()
		return buf.Bytes(), filename + ".gz", nil
	}

	return jsonData, filename, nil
}

func (e *Exporter) exportTAR(entries []*ArchiveEntry, opts ExportOptions) ([]byte, string, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for _, entry := range entries {
		// Add metadata file
		metaData := map[string]interface{}{
			"hash":              entry.Hash,
			"command":           entry.Command,
			"working_directory": entry.WorkingDirectory,
			"category":          entry.Category,
			"agent":             entry.Agent,
			"created_at":        entry.CreatedAt,
			"tags":              entry.Tags,
			"metadata":          entry.Metadata,
		}

		metaJSON, _ := json.Marshal(metaData)

		metaHeader := &tar.Header{
			Name:    fmt.Sprintf("%s/meta.json", entry.Hash[:16]),
			Size:    int64(len(metaJSON)),
			Mode:    0644,
			ModTime: time.Now(),
		}

		if err := tw.WriteHeader(metaHeader); err != nil {
			return nil, "", err
		}
		if _, err := tw.Write(metaJSON); err != nil {
			return nil, "", err
		}

		if opts.IncludeRaw {
			// Add original content
			origHeader := &tar.Header{
				Name:    fmt.Sprintf("%s/original", entry.Hash[:16]),
				Size:    int64(len(entry.OriginalContent)),
				Mode:    0644,
				ModTime: time.Now(),
			}

			if err := tw.WriteHeader(origHeader); err != nil {
				return nil, "", err
			}
			if _, err := tw.Write(entry.OriginalContent); err != nil {
				return nil, "", err
			}

			// Add filtered content if different
			if entry.FilteredContent != nil && !bytes.Equal(entry.OriginalContent, entry.FilteredContent) {
				filtHeader := &tar.Header{
					Name:    fmt.Sprintf("%s/filtered", entry.Hash[:16]),
					Size:    int64(len(entry.FilteredContent)),
					Mode:    0644,
					ModTime: time.Now(),
				}

				if err := tw.WriteHeader(filtHeader); err != nil {
					return nil, "", err
				}
				if _, err := tw.Write(entry.FilteredContent); err != nil {
					return nil, "", err
				}
			}
		}
	}

	if err := tw.Close(); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("archives_%d.tar", len(entries))

	if opts.Compression {
		var gzBuf bytes.Buffer
		gw := gzip.NewWriter(&gzBuf)
		if _, err := gw.Write(buf.Bytes()); err != nil {
			return nil, "", err
		}
		gw.Close()
		return gzBuf.Bytes(), filename + ".gz", nil
	}

	return buf.Bytes(), filename, nil
}

// Importer handles archive import operations
type Importer struct {
	manager *ArchiveManager
}

// NewImporter creates a new importer
func NewImporter(manager *ArchiveManager) *Importer {
	return &Importer{manager: manager}
}

// ImportResult contains import statistics
type ImportResult struct {
	Imported  int
	Skipped   int
	Errors    int
	TotalSize int64
}

// ImportFromFile imports archives from a file
func (i *Importer) ImportFromFile(ctx context.Context, filepath string) (*ImportResult, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect format by extension
	ext := filepath[strings.LastIndex(filepath, "."):]

	switch ext {
	case ".json":
		return i.importJSON(ctx, data)
	case ".tar", ".tar.gz":
		return i.importTAR(ctx, data)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func (i *Importer) importJSON(ctx context.Context, data []byte) (*ImportResult, error) {
	result := &ImportResult{}

	// Try single entry first
	var singleEntry map[string]interface{}
	if err := json.Unmarshal(data, &singleEntry); err == nil {
		if _, ok := singleEntry["hash"]; ok {
			return i.importSingleEntry(ctx, singleEntry)
		}
	}

	// Try array
	var entries []map[string]interface{}
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	for _, entryData := range entries {
		r, err := i.importSingleEntry(ctx, entryData)
		if err != nil {
			result.Errors++
			continue
		}
		result.Imported += r.Imported
		result.Skipped += r.Skipped
	}

	return result, nil
}

func (i *Importer) importSingleEntry(ctx context.Context, data map[string]interface{}) (*ImportResult, error) {
	result := &ImportResult{}

	// Check if already exists
	hash, _ := data["hash"].(string)
	if hash != "" {
		_, err := i.manager.Retrieve(ctx, hash)
		if err == nil {
			result.Skipped = 1
			return result, nil
		}
	}

	// Create entry
	entry := &ArchiveEntry{
		Hash:             hash,
		Command:          getString(data, "command"),
		WorkingDirectory: getString(data, "working_directory"),
		Category:         ArchiveCategory(getString(data, "category", "command")),
		Agent:            getString(data, "agent"),
		Tags:             getStringSlice(data, "tags"),
	}

	if content, ok := data["original_content"].(string); ok {
		entry.OriginalContent = []byte(content)
		entry.OriginalSize = int64(len(entry.OriginalContent))
	}

	if content, ok := data["filtered_content"].(string); ok {
		entry.FilteredContent = []byte(content)
	}

	// Parse created_at
	if createdAt, ok := data["created_at"].(string); ok {
		t, _ := time.Parse(time.RFC3339, createdAt)
		entry.CreatedAt = t
	}

	// Import
	newHash, err := i.manager.Archive(ctx, entry)
	if err != nil {
		result.Errors = 1
		return result, err
	}

	result.Imported = 1
	result.TotalSize = entry.OriginalSize
	_ = newHash

	return result, nil
}

func (i *Importer) importTAR(ctx context.Context, data []byte) (*ImportResult, error) {
	result := &ImportResult{}

	tr := tar.NewReader(bytes.NewReader(data))

	entries := make(map[string]map[string]interface{})

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors++
			continue
		}

		dir := filepath.Dir(header.Name)
		filename := filepath.Base(header.Name)

		if _, ok := entries[dir]; !ok {
			entries[dir] = make(map[string]interface{})
		}

		buf := make([]byte, header.Size)
		if _, err := io.ReadFull(tr, buf); err != nil {
			result.Errors++
			continue
		}

		switch filename {
		case "meta.json":
			var meta map[string]interface{}
			json.Unmarshal(buf, &meta)
			for k, v := range meta {
				entries[dir][k] = v
			}
		case "original":
			entries[dir]["original_content"] = string(buf)
		case "filtered":
			entries[dir]["filtered_content"] = string(buf)
		}
	}

	for _, entryData := range entries {
		r, err := i.importSingleEntry(ctx, entryData)
		if err != nil {
			result.Errors++
			continue
		}
		result.Imported += r.Imported
		result.Skipped += r.Skipped
	}

	return result, nil
}

func getString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key].(string); ok {
			return v
		}
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key].([]interface{}); ok {
		var result []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}
