package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRotatingFileWriter_DefaultMaxSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	rw, err := NewRotatingFileWriter(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rw.Close()
	if rw.maxSize != 10*1024*1024 {
		t.Errorf("expected default maxSize 10 MiB, got %d", rw.maxSize)
	}
}

func TestRotatingFileWriter_WritesData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	rw, err := NewRotatingFileWriter(path, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rw.Close()

	msg := "hello rotation\n"
	if _, err := rw.Write([]byte(msg)); err != nil {
		t.Fatalf("write error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if !strings.Contains(string(data), "hello rotation") {
		t.Errorf("expected written content in file, got: %s", data)
	}
}

func TestRotatingFileWriter_RotatesOnSizeExceeded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	// maxSize of 10 bytes forces rotation after first write
	rw, err := NewRotatingFileWriter(path, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rw.Close()

	if _, err := rw.Write([]byte("first-entry\n")); err != nil {
		t.Fatalf("first write error: %v", err)
	}
	if _, err := rw.Write([]byte("second-entry\n")); err != nil {
		t.Fatalf("second write error: %v", err)
	}

	entries, err := filepath.Glob(filepath.Join(dir, "audit*.log"))
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	// original + at least one rotated file
	if len(entries) < 2 {
		t.Errorf("expected at least 2 log files after rotation, found %d: %v", len(entries), entries)
	}
}

func TestNewRotatingFileWriter_InvalidPath(t *testing.T) {
	_, err := NewRotatingFileWriter("/nonexistent/dir/audit.log", 1024)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestRotatingFileWriter_Path(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	rw, err := NewRotatingFileWriter(path, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rw.Close()
	if rw.Path() != path {
		t.Errorf("expected path %s, got %s", path, rw.Path())
	}
}
