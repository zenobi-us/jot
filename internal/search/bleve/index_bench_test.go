package bleve

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zenobi-us/jot/internal/search"
)

// BenchmarkIndex_Add measures single document indexing performance
func BenchmarkIndex_Add(b *testing.B) {
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	if err != nil {
		b.Fatalf("failed to create index: %v", err)
	}
	defer func() { _ = idx.Close() }()

	ctx := context.Background()
	doc := search.Document{
		Path:     "benchmark/test.md",
		Title:    "Benchmark Document",
		Body:     "This is a benchmark test document with some content to index.",
		Lead:     "This is a benchmark test",
		Tags:     []string{"benchmark", "test"},
		Created:  time.Now(),
		Modified: time.Now(),
		Checksum: "bench123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.Path = fmt.Sprintf("benchmark/test-%d.md", i)
		if err := idx.Add(ctx, doc); err != nil {
			b.Fatalf("failed to add document: %v", err)
		}
	}
}

// BenchmarkIndex_BulkAdd measures bulk indexing performance
// Target: 10k documents in <500ms
func BenchmarkIndex_BulkAdd(b *testing.B) {
	const docCount = 10000

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		storage := MemStorage()
		idx, err := NewIndex(storage, Options{InMemory: true})
		if err != nil {
			b.Fatalf("failed to create index: %v", err)
		}

		ctx := context.Background()
		baseDoc := search.Document{
			Title:    "Bulk Document",
			Body:     "This is a bulk indexing test with some content.",
			Lead:     "This is a bulk indexing test",
			Tags:     []string{"bulk", "test"},
			Created:  time.Now(),
			Modified: time.Now(),
		}

		b.StartTimer()
		for j := 0; j < docCount; j++ {
			baseDoc.Path = fmt.Sprintf("bulk/doc-%d.md", j)
			baseDoc.Checksum = fmt.Sprintf("bulk%d", j)
			if err := idx.Add(ctx, baseDoc); err != nil {
				b.Fatalf("failed to add document %d: %v", j, err)
			}
		}
		b.StopTimer()

		_ = idx.Close()
	}

	b.ReportMetric(float64(docCount)/b.Elapsed().Seconds(), "docs/sec")
}

// BenchmarkIndex_Find_Simple measures simple search query performance
// Target: <25ms
func BenchmarkIndex_Find_Simple(b *testing.B) {
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	if err != nil {
		b.Fatalf("failed to create index: %v", err)
	}
	defer func() { _ = idx.Close() }()

	// Populate with test documents
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		doc := search.Document{
			Path:     fmt.Sprintf("docs/test-%d.md", i),
			Title:    fmt.Sprintf("Document %d", i),
			Body:     fmt.Sprintf("This is document number %d with some searchable content.", i),
			Tags:     []string{"test", "benchmark"},
			Modified: time.Now(),
		}
		if err := idx.Add(ctx, doc); err != nil {
			b.Fatalf("failed to add document: %v", err)
		}
	}

	query := &search.Query{
		Expressions: []search.Expr{
			search.TermExpr{Value: "searchable"},
		},
	}

	opts := search.FindOpts{}.WithQuery(query).WithLimit(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := idx.Find(ctx, opts)
		if err != nil {
			b.Fatalf("failed to find: %v", err)
		}
		if results.Total == 0 {
			b.Fatal("expected results")
		}
	}
}

// BenchmarkIndex_Find_Complex measures complex query performance
func BenchmarkIndex_Find_Complex(b *testing.B) {
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	if err != nil {
		b.Fatalf("failed to create index: %v", err)
	}
	defer func() { _ = idx.Close() }()

	// Populate with diverse documents
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		tags := []string{"general"}
		tagNum := i % 10
		tags = append(tags, fmt.Sprintf("tag%d", tagNum))

		doc := search.Document{
			Path:     fmt.Sprintf("notes/%d.md", i),
			Title:    fmt.Sprintf("Note %d", i),
			Body:     fmt.Sprintf("Content for note %d with various tags and metadata.", i),
			Tags:     tags,
			Created:  time.Now().Add(-time.Duration(i) * time.Hour),
			Modified: time.Now(),
		}
		if err := idx.Add(ctx, doc); err != nil {
			b.Fatalf("failed to add document: %v", err)
		}
	}

	// Complex query with multiple conditions
	opts := search.FindOpts{}.
		WithTags("tag5").
		WithPath("notes/").
		WithLimit(20)

	// Verify we have data before benchmarking
	results, err := idx.Find(ctx, opts)
	if err != nil {
		b.Fatalf("failed to verify query: %v", err)
	}
	if results.Total == 0 {
		b.Fatalf("no results found for query - expected documents with tag5")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := idx.Find(ctx, opts)
		if err != nil {
			b.Fatalf("failed to find: %v", err)
		}
		if results.Total == 0 {
			b.Fatal("expected results")
		}
	}
}

// BenchmarkIndex_FindByPath measures exact path lookup performance
func BenchmarkIndex_FindByPath(b *testing.B) {
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	if err != nil {
		b.Fatalf("failed to create index: %v", err)
	}
	defer func() { _ = idx.Close() }()

	// Add test documents
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		doc := search.Document{
			Path:     fmt.Sprintf("path/to/doc-%d.md", i),
			Title:    fmt.Sprintf("Document %d", i),
			Modified: time.Now(),
		}
		if err := idx.Add(ctx, doc); err != nil {
			b.Fatalf("failed to add document: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("path/to/doc-%d.md", i%1000)
		doc, err := idx.FindByPath(ctx, path)
		if err != nil {
			b.Fatalf("failed to find by path: %v", err)
		}
		if doc.Path != path {
			b.Fatalf("wrong document returned")
		}
	}
}

// BenchmarkIndex_Count measures count query performance
func BenchmarkIndex_Count(b *testing.B) {
	storage := MemStorage()
	idx, err := NewIndex(storage, Options{InMemory: true})
	if err != nil {
		b.Fatalf("failed to create index: %v", err)
	}
	defer func() { _ = idx.Close() }()

	// Add test documents
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		doc := search.Document{
			Path:     fmt.Sprintf("docs/%d.md", i),
			Tags:     []string{fmt.Sprintf("category%d", i%5)},
			Modified: time.Now(),
		}
		if err := idx.Add(ctx, doc); err != nil {
			b.Fatalf("failed to add document: %v", err)
		}
	}

	opts := search.FindOpts{}.WithTags("category2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count, err := idx.Count(ctx, opts)
		if err != nil {
			b.Fatalf("failed to count: %v", err)
		}
		if count == 0 {
			b.Fatal("expected non-zero count")
		}
	}
}
