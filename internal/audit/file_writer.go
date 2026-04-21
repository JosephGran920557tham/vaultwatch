package audit

import (
	"fmt"
	"os"
)

// FileWriter wraps an *os.File so it satisfies io.Writer and can be closed.
type FileWriter struct {
	f *os.File
}

// NewFileWriter opens (or creates) the file at path for append-only writes.
func NewFileWriter(path string) (*FileWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return nil, fmt.Errorf("audit: open file %q: %w", path, err)
	}
	return &FileWriter{f: f}, nil
}

// Write implements io.Writer.
func (fw *FileWriter) Write(p []byte) (int, error) {
	n, err := fw.f.Write(p)
	if err != nil {
		return n, fmt.Errorf("audit: write to file %q: %w", fw.f.Name(), err)
	}
	return n, nil
}

// Close closes the underlying file.
func (fw *FileWriter) Close() error {
	if err := fw.f.Close(); err != nil {
		return fmt.Errorf("audit: close file %q: %w", fw.f.Name(), err)
	}
	return nil
}

// Path returns the path of the underlying file.
func (fw *FileWriter) Path() string {
	return fw.f.Name()
}
