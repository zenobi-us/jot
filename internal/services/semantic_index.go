package services

import (
	"context"
	"errors"

	"github.com/zenobi-us/opennotes/internal/search"
)

// ErrSemanticUnavailable is returned when semantic search is requested
// but no semantic backend is configured or available.
var ErrSemanticUnavailable = errors.New("semantic backend unavailable")

// SemanticFindOpts controls semantic retrieval behavior.
type SemanticFindOpts struct {
	TopK int
}

// SemanticResult is a semantic candidate document and its similarity score.
type SemanticResult struct {
	Document search.Document
	Score    float64
}

// SemanticIndex is the contract for semantic retrieval backends.
type SemanticIndex interface {
	FindSimilar(ctx context.Context, query string, opts SemanticFindOpts) ([]SemanticResult, error)
	Close() error
	IsAvailable() bool
}

// NoopSemanticIndex is a safe fallback backend used when semantic retrieval
// is not configured yet.
type NoopSemanticIndex struct{}

// NewNoopSemanticIndex creates a no-op semantic backend.
func NewNoopSemanticIndex() SemanticIndex {
	return &NoopSemanticIndex{}
}

// FindSimilar returns ErrSemanticUnavailable because this backend is disabled.
func (n *NoopSemanticIndex) FindSimilar(ctx context.Context, query string, opts SemanticFindOpts) ([]SemanticResult, error) {
	return nil, ErrSemanticUnavailable
}

// Close is a no-op for disabled semantic backend.
func (n *NoopSemanticIndex) Close() error {
	return nil
}

// IsAvailable returns false because this backend is intentionally disabled.
func (n *NoopSemanticIndex) IsAvailable() bool {
	return false
}
