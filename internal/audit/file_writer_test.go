package audit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/audit"
)

func TestFileWriter_WritesAndClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	fw, err := audit.NewFileWriter(path)
	if err != nil {
		t.Fatalf("NewFileWriter: %v", err)
	}
	defer fw.Close()

	l := audit.NewLogger(fw)
	a := makeAlert("lease/xyz", "test entry", 60*time.Second, alert.LevelCritical)
	if err := l.Record(a); err != nil {
		t.Fatalf("Record: %v", err)
	}
	fw.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "lease/xyz") {
		t.Errorf("expected lease/xyz in log, got: %s", data)
	}
}

func TestFileWriter_AppendsOnReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	for i := 0; i < 2; i++ {
		fw, err := audit.NewFileWriter(path)
		if err != nil {
			t.Fatalf("open %d: %v", i, err)
		}
		l := audit.NewLogger(fw)
		_ = l.Record(makeAlert("lease/loop", "msg", 5*time.Second, alert.LevelInfo))
		fw.Close()
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines after two opens, got %d", len(lines))
	}
}

func TestNewFileWriter_InvalidPath(t *testing.T) {
	_, err := audit.NewFileWriter("/nonexistent-dir/audit.log")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestFileWriter_Path(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	fw, _ := audit.NewFileWriter(path)
	defer fw.Close()
	if fw.Path() != path {
		t.Errorf("expected path %s, got %s", path, fw.Path())
	}
}
