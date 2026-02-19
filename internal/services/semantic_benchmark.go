package services

import (
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	defaultSemanticBenchmarkWarmupRuns = 3
	defaultSemanticBenchmarkRuns       = 5
)

// SemanticBenchmarkDataset describes one benchmark dataset profile.
type SemanticBenchmarkDataset struct {
	Name      string `json:"name"`
	NoteCount int    `json:"note_count"`
}

// SemanticBenchmarkThresholds defines latency limits for benchmark pass/fail checks.
type SemanticBenchmarkThresholds struct {
	MaxP50         time.Duration `json:"max_p50"`
	MaxP95         time.Duration `json:"max_p95"`
	MaxDatasetSize int           `json:"max_dataset_size"`
}

// SemanticBenchmarkEnvironment captures host metadata for reproducible benchmark reports.
type SemanticBenchmarkEnvironment struct {
	GoVersion string `json:"go_version"`
	GOOS      string `json:"go_os"`
	GOARCH    string `json:"go_arch"`
	NumCPU    int    `json:"num_cpu"`
}

// SemanticBenchmarkConfig configures deterministic semantic benchmark execution.
type SemanticBenchmarkConfig struct {
	Datasets   []SemanticBenchmarkDataset
	Modes      []RetrievalMode
	WarmupRuns int
	Runs       int
	Thresholds SemanticBenchmarkThresholds
	Now        func() time.Time
}

// SemanticBenchmarkResult captures percentile and threshold status for one mode + dataset.
type SemanticBenchmarkResult struct {
	Dataset           SemanticBenchmarkDataset    `json:"dataset"`
	Mode              RetrievalMode               `json:"mode"`
	Runs              int                         `json:"runs"`
	Min               time.Duration               `json:"min"`
	Max               time.Duration               `json:"max"`
	Mean              time.Duration               `json:"mean"`
	P50               time.Duration               `json:"p50"`
	P95               time.Duration               `json:"p95"`
	ThresholdApplied  bool                        `json:"threshold_applied"`
	ThresholdPassed   bool                        `json:"threshold_passed"`
	ThresholdFailures []string                    `json:"threshold_failures,omitempty"`
	Thresholds        SemanticBenchmarkThresholds `json:"thresholds"`
}

// SemanticBenchmarkReport is the full benchmark output for all configured datasets and modes.
type SemanticBenchmarkReport struct {
	GeneratedAt time.Time                    `json:"generated_at"`
	WarmupRuns  int                          `json:"warmup_runs"`
	Runs        int                          `json:"runs"`
	Environment SemanticBenchmarkEnvironment `json:"environment"`
	Results     []SemanticBenchmarkResult    `json:"results"`
}

// SemanticBenchmarkMeasureFunc returns one measured duration for a mode and dataset.
type SemanticBenchmarkMeasureFunc func(dataset SemanticBenchmarkDataset, mode RetrievalMode, run int, warmup bool) (time.Duration, error)

// RunSemanticBenchmark executes deterministic benchmark runs for keyword/semantic/hybrid retrieval modes.
func RunSemanticBenchmark(cfg SemanticBenchmarkConfig, measure SemanticBenchmarkMeasureFunc) (SemanticBenchmarkReport, error) {
	if measure == nil {
		return SemanticBenchmarkReport{}, fmt.Errorf("benchmark measure function is required")
	}
	if len(cfg.Datasets) == 0 {
		return SemanticBenchmarkReport{}, fmt.Errorf("at least one benchmark dataset is required")
	}

	modes := cfg.Modes
	if len(modes) == 0 {
		modes = []RetrievalMode{RetrievalModeKeyword, RetrievalModeSemantic, RetrievalModeHybrid}
	}

	warmups := cfg.WarmupRuns
	if warmups < 0 {
		warmups = 0
	}
	if warmups == 0 {
		warmups = defaultSemanticBenchmarkWarmupRuns
	}

	runs := cfg.Runs
	if runs <= 0 {
		runs = defaultSemanticBenchmarkRuns
	}

	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	report := SemanticBenchmarkReport{
		GeneratedAt: nowFn(),
		WarmupRuns:  warmups,
		Runs:        runs,
		Environment: currentSemanticBenchmarkEnvironment(),
		Results:     make([]SemanticBenchmarkResult, 0, len(cfg.Datasets)*len(modes)),
	}

	for _, dataset := range cfg.Datasets {
		if dataset.Name == "" {
			dataset.Name = fmt.Sprintf("dataset-%d", dataset.NoteCount)
		}

		for _, mode := range modes {
			for i := 0; i < warmups; i++ {
				if _, err := measure(dataset, mode, i, true); err != nil {
					return SemanticBenchmarkReport{}, fmt.Errorf("benchmark warmup failed for %s/%s: %w", dataset.Name, mode, err)
				}
			}

			samples := make([]time.Duration, 0, runs)
			for i := 0; i < runs; i++ {
				duration, err := measure(dataset, mode, i, false)
				if err != nil {
					return SemanticBenchmarkReport{}, fmt.Errorf("benchmark run failed for %s/%s: %w", dataset.Name, mode, err)
				}
				samples = append(samples, duration)
			}

			result := buildSemanticBenchmarkResult(dataset, mode, samples, cfg.Thresholds)
			report.Results = append(report.Results, result)
		}
	}

	return report, nil
}

// ToJSON serializes the benchmark report for CI or regression tooling.
func (r SemanticBenchmarkReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ToMarkdown renders a human-readable benchmark summary with threshold status.
func (r SemanticBenchmarkReport) ToMarkdown() string {
	var b strings.Builder

	b.WriteString("## Semantic Benchmark Report\n\n")
	b.WriteString(fmt.Sprintf("Generated: %s\n", r.GeneratedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf(
		"Environment: %s (%s/%s, cpu=%d)\n\n",
		r.Environment.GoVersion,
		r.Environment.GOOS,
		r.Environment.GOARCH,
		r.Environment.NumCPU,
	))

	b.WriteString("| Dataset | Notes | Mode | Runs | P50 | P95 | Threshold |\n")
	b.WriteString("|---------|-------|------|------|-----|-----|-----------|\n")

	for _, result := range r.Results {
		status := "N/A"
		if result.ThresholdApplied {
			if result.ThresholdPassed {
				status = "PASS"
			} else {
				status = "FAIL"
			}
		}

		b.WriteString(fmt.Sprintf(
			"| %s | %d | %s | %d | %s | %s | %s |\n",
			result.Dataset.Name,
			result.Dataset.NoteCount,
			result.Mode,
			result.Runs,
			result.P50,
			result.P95,
			status,
		))
	}

	return b.String()
}

// DefaultSemanticBenchmarkDatasets returns the canonical dataset sizes used in latency checks.
func DefaultSemanticBenchmarkDatasets() []SemanticBenchmarkDataset {
	return []SemanticBenchmarkDataset{
		{Name: "small", NoteCount: 1000},
		{Name: "medium", NoteCount: 10000},
		{Name: "target", NoteCount: 50000},
	}
}

// DefaultSemanticBenchmarkQueryCorpus returns the deterministic query corpus for benchmark runs.
func DefaultSemanticBenchmarkQueryCorpus() []string {
	return []string{
		"meeting", "project", "workflow", "task", "architecture",
		"retrospective notes", "planning discussion", "design decision", "incident followup", "team alignment",
		"tag:workflow status:active", "tag:project -status:archived", "priority:high sprint:current", "author:alice -tag:archived", "status:in-progress tag:task",
		"meeting notes not archived", "project handoff checklist", "architecture decision timeline", "release blocker mitigation", "customer escalation summary",
	}
}

func buildSemanticBenchmarkResult(
	dataset SemanticBenchmarkDataset,
	mode RetrievalMode,
	samples []time.Duration,
	thresholds SemanticBenchmarkThresholds,
) SemanticBenchmarkResult {
	sorted := make([]time.Duration, len(samples))
	copy(sorted, samples)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	result := SemanticBenchmarkResult{
		Dataset:    dataset,
		Mode:       mode,
		Runs:       len(samples),
		Min:        sorted[0],
		Max:        sorted[len(sorted)-1],
		Mean:       meanDuration(samples),
		P50:        percentileDuration(sorted, 0.50),
		P95:        percentileDuration(sorted, 0.95),
		Thresholds: thresholds,
	}

	if thresholds.MaxDatasetSize > 0 && dataset.NoteCount > thresholds.MaxDatasetSize {
		return result
	}

	result.ThresholdApplied = thresholds.MaxP50 > 0 || thresholds.MaxP95 > 0
	if !result.ThresholdApplied {
		return result
	}

	result.ThresholdPassed = true

	if thresholds.MaxP50 > 0 && result.P50 > thresholds.MaxP50 {
		result.ThresholdPassed = false
		result.ThresholdFailures = append(result.ThresholdFailures, fmt.Sprintf("p50 %s > %s", result.P50, thresholds.MaxP50))
	}
	if thresholds.MaxP95 > 0 && result.P95 > thresholds.MaxP95 {
		result.ThresholdPassed = false
		result.ThresholdFailures = append(result.ThresholdFailures, fmt.Sprintf("p95 %s > %s", result.P95, thresholds.MaxP95))
	}

	return result
}

func percentileDuration(sortedSamples []time.Duration, percentile float64) time.Duration {
	if len(sortedSamples) == 0 {
		return 0
	}
	if percentile <= 0 {
		return sortedSamples[0]
	}
	if percentile >= 1 {
		return sortedSamples[len(sortedSamples)-1]
	}

	rank := int(math.Ceil(percentile*float64(len(sortedSamples)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(sortedSamples) {
		rank = len(sortedSamples) - 1
	}

	return sortedSamples[rank]
}

func meanDuration(samples []time.Duration) time.Duration {
	if len(samples) == 0 {
		return 0
	}

	var total int64
	for _, sample := range samples {
		total += int64(sample)
	}

	return time.Duration(total / int64(len(samples)))
}

func currentSemanticBenchmarkEnvironment() SemanticBenchmarkEnvironment {
	return SemanticBenchmarkEnvironment{
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		NumCPU:    runtime.NumCPU(),
	}
}
