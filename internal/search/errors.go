package search

import "errors"

// Common errors returned by search operations.
var (
	// ErrNotFound is returned when a document is not in the index.
	ErrNotFound = errors.New("document not found")

	// ErrIndexClosed is returned when operating on a closed index.
	ErrIndexClosed = errors.New("index is closed")

	// ErrInvalidQuery is returned when a query cannot be parsed.
	ErrInvalidQuery = errors.New("invalid query")

	// ErrIndexCorrupted is returned when the index is in an invalid state.
	ErrIndexCorrupted = errors.New("index is corrupted")

	// ErrIndexLocked is returned when another process holds the index lock.
	ErrIndexLocked = errors.New("index is locked by another process")

	// ErrNotebookNotFound is returned when the notebook directory doesn't exist.
	ErrNotebookNotFound = errors.New("notebook not found")

	// ErrStorageError is returned for filesystem-related errors.
	ErrStorageError = errors.New("storage error")
)

// IndexError wraps an error with additional context.
type IndexError struct {
	// Op is the operation that failed (e.g., "add", "find", "reindex")
	Op string

	// Path is the document path involved (if any)
	Path string

	// Err is the underlying error
	Err error
}

// Error implements the error interface.
func (e *IndexError) Error() string {
	if e.Path != "" {
		return e.Op + " " + e.Path + ": " + e.Err.Error()
	}
	return e.Op + ": " + e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *IndexError) Unwrap() error {
	return e.Err
}

// Is reports whether the error matches the target.
func (e *IndexError) Is(target error) bool {
	return errors.Is(e.Err, target)
}
