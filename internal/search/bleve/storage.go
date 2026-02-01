package bleve

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
	"github.com/zenobi-us/opennotes/internal/search"
)

// AferoStorage adapts an afero.Fs to the search.Storage interface.
type AferoStorage struct {
	fs   afero.Fs
	root string
}

// NewAferoStorage creates a new AferoStorage with the given filesystem and root path.
func NewAferoStorage(fs afero.Fs, root string) *AferoStorage {
	return &AferoStorage{
		fs:   fs,
		root: root,
	}
}

// fullPath joins the root with a relative path.
func (s *AferoStorage) fullPath(path string) string {
	return filepath.Join(s.root, path)
}

// Read reads the content of a file.
func (s *AferoStorage) Read(path string) ([]byte, error) {
	return afero.ReadFile(s.fs, s.fullPath(path))
}

// ReadStream opens a file for reading.
func (s *AferoStorage) ReadStream(path string) (io.ReadCloser, error) {
	return s.fs.Open(s.fullPath(path))
}

// Write writes content to a file, creating it if necessary.
func (s *AferoStorage) Write(path string, content []byte) error {
	fullPath := s.fullPath(path)
	// Create parent directories if needed
	dir := filepath.Dir(fullPath)
	if err := s.fs.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return afero.WriteFile(s.fs, fullPath, content, 0644)
}

// Exists returns true if the path exists.
func (s *AferoStorage) Exists(path string) bool {
	exists, _ := afero.Exists(s.fs, s.fullPath(path))
	return exists
}

// Stat returns file metadata.
func (s *AferoStorage) Stat(path string) (search.FileInfo, error) {
	info, err := s.fs.Stat(s.fullPath(path))
	if err != nil {
		return search.FileInfo{}, err
	}
	return toFileInfo(path, info), nil
}

// Walk traverses the directory tree starting at root.
func (s *AferoStorage) Walk(root string, walkFn search.WalkFunc) error {
	return afero.Walk(s.fs, s.fullPath(root), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return walkFn(path, search.FileInfo{}, err)
		}
		// Make path relative to storage root
		relPath, relErr := filepath.Rel(s.root, path)
		if relErr != nil {
			relPath = path
		}
		return walkFn(relPath, toFileInfo(relPath, info), nil)
	})
}

// List returns entries in a directory (non-recursive).
func (s *AferoStorage) List(path string) ([]search.FileInfo, error) {
	entries, err := afero.ReadDir(s.fs, s.fullPath(path))
	if err != nil {
		return nil, err
	}
	result := make([]search.FileInfo, len(entries))
	for i, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		result[i] = toFileInfo(entryPath, entry)
	}
	return result, nil
}

// Remove deletes a file.
func (s *AferoStorage) Remove(path string) error {
	return s.fs.Remove(s.fullPath(path))
}

// Rename moves a file from oldPath to newPath.
func (s *AferoStorage) Rename(oldPath, newPath string) error {
	return s.fs.Rename(s.fullPath(oldPath), s.fullPath(newPath))
}

// Root returns the absolute path to the notebook root.
func (s *AferoStorage) Root() string {
	return s.root
}

// toFileInfo converts os.FileInfo to search.FileInfo.
func toFileInfo(path string, info fs.FileInfo) search.FileInfo {
	return search.FileInfo{
		Path:    path,
		Name:    info.Name(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
		Mode:    info.Mode(),
	}
}

// Ensure AferoStorage implements search.Storage.
var _ search.Storage = (*AferoStorage)(nil)

// MemStorage creates an in-memory storage for testing.
func MemStorage() *AferoStorage {
	return NewAferoStorage(afero.NewMemMapFs(), "/")
}

// OsStorage creates a real filesystem storage.
func OsStorage(root string) *AferoStorage {
	return NewAferoStorage(afero.NewOsFs(), root)
}

// IndexPath returns the default index path within a notebook.
const IndexDir = ".opennotes/index"

// CreateTestDocument creates a test document with the given path and content.
// Useful for testing.
func CreateTestDocument(path, title, body string, tags []string) search.Document {
	now := time.Now()
	return search.Document{
		Path:     path,
		Title:    title,
		Body:     body,
		Lead:     body[:min(len(body), 100)],
		Tags:     tags,
		Created:  now,
		Modified: now,
		Checksum: path, // Simple checksum for tests
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
