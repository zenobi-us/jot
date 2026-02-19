package search

import (
	"io"
	"io/fs"
	"time"
)

// Storage defines the filesystem abstraction for note storage.
//
// This interface is designed to be compatible with spf13/afero.Fs,
// allowing tests to use in-memory filesystems and production to use
// the real OS filesystem.
//
// Example with afero:
//
//	// Production
//	storage := NewAferoStorage(afero.NewOsFs(), notebookRoot)
//
//	// Testing
//	storage := NewAferoStorage(afero.NewMemMapFs(), "/")
type Storage interface {
	// Read reads the content of a file.
	// Path is relative to the notebook root.
	Read(path string) ([]byte, error)

	// ReadStream opens a file for reading.
	// The caller is responsible for closing the reader.
	ReadStream(path string) (io.ReadCloser, error)

	// Write writes content to a file, creating it if necessary.
	// Parent directories are created automatically.
	Write(path string, content []byte) error

	// Exists returns true if the path exists.
	Exists(path string) bool

	// Stat returns file metadata.
	Stat(path string) (FileInfo, error)

	// Walk traverses the directory tree starting at root.
	// The walk function is called for each file and directory.
	Walk(root string, walkFn WalkFunc) error

	// List returns entries in a directory (non-recursive).
	List(path string) ([]FileInfo, error)

	// Remove deletes a file.
	Remove(path string) error

	// Rename moves a file from oldPath to newPath.
	Rename(oldPath, newPath string) error

	// Root returns the absolute path to the notebook root.
	Root() string
}

// WalkFunc is the type of function called by Storage.Walk.
// It mirrors filepath.WalkFunc but uses our FileInfo type.
type WalkFunc func(path string, info FileInfo, err error) error

// FileInfo contains metadata about a file.
// This mirrors fs.FileInfo but is a concrete type for easier testing.
type FileInfo struct {
	// Path is the relative path from notebook root
	Path string

	// Name is the base name of the file
	Name string

	// Size is the file size in bytes
	Size int64

	// ModTime is the modification time
	ModTime time.Time

	// IsDir is true if this is a directory
	IsDir bool

	// Mode is the file mode bits
	Mode fs.FileMode
}

// SkipDir is used as a return value from WalkFunc to indicate
// that the directory named in the call is to be skipped.
var SkipDir = fs.SkipDir

// SkipAll is used as a return value from WalkFunc to indicate
// that all remaining files and directories are to be skipped.
var SkipAll = fs.SkipAll
