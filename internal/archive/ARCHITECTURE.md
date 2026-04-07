# RewindStore Architecture Design

## Overview
RewindStore provides zero-information-loss archiving for TokMan filtered content, enabling retrieval by SHA-256 hash.

## Architecture Components

### 1. Data Model

```
ArchiveEntry
├── ID (auto-increment)
├── Hash (SHA-256, unique, indexed)
├── OriginalContent (compressed)
├── FilteredContent (what was shown to user)
├── Command (the command that generated it)
├── WorkingDirectory (where command was run)
├── Timestamp (when archived)
├── ExpiresAt (optional expiration)
├── Tags (JSON array)
├── Category (command/session/user)
├── Agent (which AI agent)
├── CompressionType (gzip/brotli/none)
├── OriginalSize (bytes)
├── CompressedSize (bytes)
└── Metadata (JSON: user, project, etc.)
```

### 2. Storage Layer

```
SQLite Database (~/.local/share/tokman/archive.db)
├── archives table (main storage)
├── archive_tags table (many-to-many)
├── archive_access_log (retrieval tracking)
└── archive_stats (aggregated metrics)

File System Backup (~/.local/share/tokman/archive/)
├── <hash_prefix>/<hash_suffix>.bin (large content)
└── manifest.json (file-to-db mapping)
```

### 3. Core Components

```
RewindStore
├── ArchiveManager (orchestration)
├── HashCalculator (SHA-256)
├── CompressionEngine (gzip/brotli)
├── StorageBackend (SQLite + FS)
├── EncryptionLayer (optional AES)
├── IntegrityVerifier (hash validation)
├── ExpirationManager (cleanup)
└── QuotaEnforcer (size limits)
```

### 4. API Design

```go
// Core interface
type ArchiveManager interface {
    Archive(content []byte, metadata ArchiveMetadata) (hash string, err error)
    Retrieve(hash string) (*ArchiveEntry, error)
    Delete(hash string) error
    List(opts ListOptions) ([]ArchiveEntry, error)
    Search(query string) ([]ArchiveEntry, error)
    Verify(hash string) (bool, error)
    Export(hash string, format string) ([]byte, error)
}
```

### 5. Integration Points

```
Pipeline Integration:
  FilterOutput → ArchiveManager.Archive() → Return hash in metadata

CLI Commands:
  tokman archive <file> [--tags] [--category] [--expires]
  tokman retrieve <hash> [--output]
  tokman archive-list [--filter]
  tokman archive-search <query>
  tokman archive-verify <hash>
  tokman archive-export <hash> [--format]
  tokman archive-import <file>
  tokman archive-delete <hash>
  tokman archive-stats
  tokman archive-cleanup

Dashboard Integration:
  REST API endpoints for web UI
  Archive browser with search/filter
  Visual archive timeline
```

### 6. Security

- SHA-256 for content integrity
- Optional AES-256 encryption for sensitive content
- Access control based on user/permissions
- Audit logging for all operations

### 7. Performance

- In-memory LRU cache for hot archives
- Lazy loading for large content
- Background compression
- Streaming for large files
- Parallel batch operations

### 8. Retention

- Configurable TTL per archive
- Automatic cleanup of expired archives
- Size-based quotas
- Manual cleanup commands
