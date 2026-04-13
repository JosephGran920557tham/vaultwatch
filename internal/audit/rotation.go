package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RotatingFileWriter wraps FileWriter with size-based log rotation.
type RotatingFileWriter struct {
	mu      sync.Mutex
	path    string
	maxSize int64
	file    *os.File
	size    int64
}

// NewRotatingFileWriter creates a RotatingFileWriter that rotates when the
// log file exceeds maxBytes bytes. If maxBytes <= 0, a default of 10 MiB is used.
func NewRotatingFileWriter(path string, maxBytes int64) (*RotatingFileWriter, error) {
	if maxBytes <= 0 {
		maxBytes = 10 * 1024 * 1024
	}
	rw := &RotatingFileWriter{path: path, maxSize: maxBytes}
	if err := rw.open(); err != nil {
		return nil, err
	}
	return rw, nil
}

func (rw *RotatingFileWriter) open() error {
	f, err := os.OpenFile(rw.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("audit rotation: open %s: %w", rw.path, err)
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return fmt.Errorf("audit rotation: stat %s: %w", rw.path, err)
	}
	rw.file = f
	rw.size = info.Size()
	return nil
}

// Write implements io.Writer. It rotates the file if writing p would exceed maxSize.
func (rw *RotatingFileWriter) Write(p []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.size+int64(len(p)) > rw.maxSize {
		if err := rw.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := rw.file.Write(p)
	rw.size += int64(n)
	return n, err
}

func (rw *RotatingFileWriter) rotate() error {
	rw.file.Close()
	ts := time.Now().UTC().Format("20060102T150405Z")
	ext := filepath.Ext(rw.path)
	base := rw.path[:len(rw.path)-len(ext)]
	dest := fmt.Sprintf("%s.%s%s", base, ts, ext)
	if err := os.Rename(rw.path, dest); err != nil {
		return fmt.Errorf("audit rotation: rename: %w", err)
	}
	return rw.open()
}

// Close closes the underlying file.
func (rw *RotatingFileWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.file.Close()
}

// Path returns the active log file path.
func (rw *RotatingFileWriter) Path() string { return rw.path }
