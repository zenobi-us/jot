package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

type scriptedBenchmarkMeasure struct {
	values map[string][]time.Duration
	calls  map[string]int
}

func newScriptedBenchmarkMeasure(values map[string][]time.Duration) *scriptedBenchmarkMeasure {
	return &scriptedBenchmarkMeasure{
		values: values,
		calls:  make(map[string]int),
	}
}

func (m *scriptedBenchmarkMeasure) measure(dataset SemanticBenchmarkDataset, mode RetrievalMode, run int, warmup bool) (time.Duration, error) {
	key := fmt.Sprintf("%s/%s", dataset.Name, mode)
	series, ok := m.values[key]
	if !ok {
		return 0, fmt.Errorf("no scripted values for %s", key)
	}

	idx := m.calls[key]
	if idx >= len(series) {
		return 0, fmt.Errorf("not enough scripted values for %s", key)
	}

	m.calls[key] = idx + 1
	return series[idx], nil
}

func TestRunSemanticBenchmark_DefaultModesAndPercentiles(t *testing.T) {
	dataset := SemanticBenchmarkDataset{Name: "small", NoteCount: 1000}
	measure := newScriptedBenchmarkMeasure(map[string][]time.Duration{
		"small/keyword":  {10 * time.Millisecond, 100 * time.Millisecond, 120 * time.Millisecond, 110 * time.Millisecond, 130 * time.Millisecond, 140 * time.Millisecond},
		"small/semantic": {12 * time.Millisecond, 200 * time.Millisecond, 210 * time.Millisecond, 220 * time.Millisecond, 230 * time.Millisecond, 240 * time.Millisecond},
		"small/hybrid":   {14 * time.Millisecond, 300 * time.Millisecond, 310 * time.Millisecond, 320 * time.Millisecond, 330 * time.Millisecond, 340 * time.Millisecond},
	})

	report, err := RunSemanticBenchmark(SemanticBenchmarkConfig{
		Datasets:   []SemanticBenchmarkDataset{dataset},
		WarmupRuns: 1,
		Runs:       5,
		Thresholds: SemanticBenchmarkThresholds{MaxP50: 500 * time.Millisecond, MaxP95: 700 * time.Millisecond, MaxDatasetSize: 50000},
		Now:        func() time.Time { return time.Date(2026, 2, 15, 1, 30, 0, 0, time.UTC) },
	}, measure.measure)
	if err != nil {
		t.Fatalf("RunSemanticBenchmark() error = %v", err)
	}

	if len(report.Results) != 3 {
		t.Fatalf("expected 3 mode results, got %d", len(report.Results))
	}

	byMode := make(map[RetrievalMode]SemanticBenchmarkResult)
	for _, result := range report.Results {
		byMode[result.Mode] = result
	}

	keyword := byMode[RetrievalModeKeyword]
	if keyword.P50 != 120*time.Millisecond {
		t.Fatalf("keyword P50 = %s, want 120ms", keyword.P50)
	}
	if keyword.P95 != 140*time.Millisecond {
		t.Fatalf("keyword P95 = %s, want 140ms", keyword.P95)
	}
	if !keyword.ThresholdPassed {
		t.Fatalf("keyword threshold should pass: %+v", keyword)
	}

	if measure.calls["small/keyword"] != 6 || measure.calls["small/semantic"] != 6 || measure.calls["small/hybrid"] != 6 {
		t.Fatalf("expected warmup + run calls per mode (6), got %+v", measure.calls)
	}
}

func TestRunSemanticBenchmark_ThresholdFailure(t *testing.T) {
	dataset := SemanticBenchmarkDataset{Name: "target", NoteCount: 50000}
	measure := newScriptedBenchmarkMeasure(map[string][]time.Duration{
		"target/keyword": {5 * time.Millisecond, 100 * time.Millisecond, 120 * time.Millisecond, 260 * time.Millisecond, 300 * time.Millisecond, 900 * time.Millisecond},
	})

	report, err := RunSemanticBenchmark(SemanticBenchmarkConfig{
		Datasets:   []SemanticBenchmarkDataset{dataset},
		Modes:      []RetrievalMode{RetrievalModeKeyword},
		WarmupRuns: 1,
		Runs:       5,
		Thresholds: SemanticBenchmarkThresholds{MaxP50: 250 * time.Millisecond, MaxP95: 750 * time.Millisecond, MaxDatasetSize: 50000},
	}, measure.measure)
	if err != nil {
		t.Fatalf("RunSemanticBenchmark() error = %v", err)
	}

	result := report.Results[0]
	if result.ThresholdPassed {
		t.Fatalf("expected threshold failure, got pass")
	}
	if len(result.ThresholdFailures) != 2 {
		t.Fatalf("expected two threshold failure reasons, got %v", result.ThresholdFailures)
	}
}

func TestRunSemanticBenchmark_ThresholdNotAppliedBeyondDatasetLimit(t *testing.T) {
	dataset := SemanticBenchmarkDataset{Name: "large", NoteCount: 75000}
	measure := newScriptedBenchmarkMeasure(map[string][]time.Duration{
		"large/hybrid": {5 * time.Millisecond, 900 * time.Millisecond, 910 * time.Millisecond, 920 * time.Millisecond, 930 * time.Millisecond, 940 * time.Millisecond},
	})

	report, err := RunSemanticBenchmark(SemanticBenchmarkConfig{
		Datasets:   []SemanticBenchmarkDataset{dataset},
		Modes:      []RetrievalMode{RetrievalModeHybrid},
		WarmupRuns: 1,
		Runs:       5,
		Thresholds: SemanticBenchmarkThresholds{MaxP50: 250 * time.Millisecond, MaxP95: 750 * time.Millisecond, MaxDatasetSize: 50000},
	}, measure.measure)
	if err != nil {
		t.Fatalf("RunSemanticBenchmark() error = %v", err)
	}

	result := report.Results[0]
	if result.ThresholdApplied {
		t.Fatalf("threshold should not apply for dataset > max size")
	}
}

func TestSemanticBenchmarkReport_ToMarkdownAndJSON(t *testing.T) {
	report := SemanticBenchmarkReport{
		GeneratedAt: time.Date(2026, 2, 15, 1, 45, 0, 0, time.UTC),
		Environment: SemanticBenchmarkEnvironment{GoVersion: "go1.x", GOOS: "linux", GOARCH: "amd64", NumCPU: 8},
		Results: []SemanticBenchmarkResult{
			{
				Dataset:          SemanticBenchmarkDataset{Name: "small", NoteCount: 1000},
				Mode:             RetrievalModeKeyword,
				Runs:             5,
				P50:              120 * time.Millisecond,
				P95:              180 * time.Millisecond,
				ThresholdApplied: true,
				ThresholdPassed:  true,
			},
			{
				Dataset:           SemanticBenchmarkDataset{Name: "target", NoteCount: 50000},
				Mode:              RetrievalModeHybrid,
				Runs:              5,
				P50:               320 * time.Millisecond,
				P95:               900 * time.Millisecond,
				ThresholdApplied:  true,
				ThresholdPassed:   false,
				ThresholdFailures: []string{"p95 900ms > 750ms"},
			},
		},
	}

	markdown := report.ToMarkdown()
	if !strings.Contains(markdown, "| Dataset | Notes | Mode | Runs | P50 | P95 | Threshold |") {
		t.Fatalf("markdown missing table header: %s", markdown)
	}
	if !strings.Contains(markdown, "PASS") || !strings.Contains(markdown, "FAIL") {
		t.Fatalf("markdown should include pass/fail status: %s", markdown)
	}

	jsonBytes, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("benchmark report JSON should be valid: %v", err)
	}
}

func TestDefaultSemanticBenchmarkQueryCorpus(t *testing.T) {
	queries := DefaultSemanticBenchmarkQueryCorpus()

	if len(queries) != 20 {
		t.Fatalf("expected 20 deterministic benchmark queries, got %d", len(queries))
	}
	if queries[0] != "meeting" {
		t.Fatalf("unexpected first benchmark query: %q", queries[0])
	}
}
