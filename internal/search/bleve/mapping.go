package bleve

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/v2/mapping"
)

// Field weight constants for BM25 ranking.
// Higher weights mean stronger signal for relevance scoring.
const (
	WeightPath  = 1000.0 // Exact path matches are strongest
	WeightTitle = 500.0  // Title matches are very important
	WeightTags  = 300.0  // Tag matches are medium-high
	WeightLead  = 50.0   // First paragraph has elevated importance
	WeightBody  = 1.0    // Body is baseline
)

// Field names in the Bleve index.
const (
	FieldPath     = "path"
	FieldTitle    = "title"
	FieldBody     = "body"
	FieldLead     = "lead"
	FieldTags     = "tags"
	FieldCreated  = "created"
	FieldModified = "modified"
	FieldChecksum = "checksum"
	FieldMetadata = "metadata"
)

// BuildDocumentMapping creates the Bleve document mapping for notes.
//
// The mapping defines how each field is indexed and its relative weight
// in BM25 scoring.
func BuildDocumentMapping() mapping.IndexMapping {
	// Create the index mapping
	indexMapping := bleve.NewIndexMapping()

	// Create the document mapping for notes
	noteMapping := bleve.NewDocumentMapping()

	// Path field - keyword analyzer for exact matching, highest weight
	pathField := bleve.NewTextFieldMapping()
	pathField.Analyzer = keyword.Name
	pathField.Store = true
	pathField.IncludeInAll = true
	noteMapping.AddFieldMappingsAt(FieldPath, pathField)

	// Title field - standard analyzer, high weight
	titleField := bleve.NewTextFieldMapping()
	titleField.Analyzer = standard.Name
	titleField.Store = true
	titleField.IncludeInAll = true
	noteMapping.AddFieldMappingsAt(FieldTitle, titleField)

	// Body field - standard analyzer, baseline weight
	bodyField := bleve.NewTextFieldMapping()
	bodyField.Analyzer = standard.Name
	bodyField.Store = true // Store body for retrieval (needed during migration)
	bodyField.IncludeInAll = true
	noteMapping.AddFieldMappingsAt(FieldBody, bodyField)

	// Lead field - standard analyzer, medium weight
	leadField := bleve.NewTextFieldMapping()
	leadField.Analyzer = standard.Name
	leadField.Store = true
	leadField.IncludeInAll = true
	noteMapping.AddFieldMappingsAt(FieldLead, leadField)

	// Tags field - simple analyzer (lowercase, no stemming)
	tagsField := bleve.NewTextFieldMapping()
	tagsField.Analyzer = simple.Name
	tagsField.Store = true
	tagsField.IncludeInAll = true
	noteMapping.AddFieldMappingsAt(FieldTags, tagsField)

	// Created field - datetime
	createdField := bleve.NewDateTimeFieldMapping()
	createdField.Store = true
	noteMapping.AddFieldMappingsAt(FieldCreated, createdField)

	// Modified field - datetime
	modifiedField := bleve.NewDateTimeFieldMapping()
	modifiedField.Store = true
	noteMapping.AddFieldMappingsAt(FieldModified, modifiedField)

	// Checksum field - keyword (exact match only)
	checksumField := bleve.NewTextFieldMapping()
	checksumField.Analyzer = keyword.Name
	checksumField.Store = true
	checksumField.Index = false // Not searchable, just stored
	noteMapping.AddFieldMappingsAt(FieldChecksum, checksumField)

	// Metadata field - dynamic for arbitrary frontmatter
	metadataMapping := bleve.NewDocumentMapping()
	metadataMapping.Dynamic = true
	noteMapping.AddSubDocumentMapping(FieldMetadata, metadataMapping)

	// Set the default document type
	indexMapping.DefaultMapping = noteMapping

	// Use standard analyzer by default
	indexMapping.DefaultAnalyzer = standard.Name

	return indexMapping
}

// BleveDocument is the internal representation for indexing.
// It mirrors search.Document but uses types compatible with Bleve.
type BleveDocument struct {
	Path     string         `json:"path"`
	Title    string         `json:"title"`
	Body     string         `json:"body"`
	Lead     string         `json:"lead"`
	Tags     []string       `json:"tags"`
	Created  string         `json:"created"`  // ISO8601 format
	Modified string         `json:"modified"` // ISO8601 format
	Checksum string         `json:"checksum"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// TimeFormat is the ISO8601 format used for date fields.
const TimeFormat = "2006-01-02T15:04:05Z07:00"
