package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	legacyNotebookConfigFile = ".opennotes.json"
)

// MigrationOptions controls migrate behavior.
type MigrationOptions struct {
	Apply    bool
	Registry Registry
}

// FileMigration reports the result for one file migration candidate.
type FileMigration struct {
	From   string
	To     string
	Status string
}

// ProfileReference identifies shell profile files that still mention legacy config/env names.
type ProfileReference struct {
	File    string
	Matches []string
}

// MigrationReport summarizes migration detection and actions.
type MigrationReport struct {
	Applied bool

	LegacyConfigPath string
	JotConfigPath    string

	GlobalConfig    FileMigration
	NotebookConfigs []FileMigration

	LegacyEnvVars     []string
	ProfileReferences []ProfileReference
	Warnings          []string
}

// MigrationService migrates older OpenNotes configuration to Jot.
type MigrationService struct{}

// NewMigrationService creates a migration service.
func NewMigrationService() *MigrationService {
	return &MigrationService{}
}

// MigrateOpenNotesToJot migrates legacy OpenNotes config files and reports env/profile references.
func (s *MigrationService) MigrateOpenNotesToJot(opts MigrationOptions) (*MigrationReport, error) {
	legacyConfigPath, jotConfigPath, err := migrationConfigPaths()
	if err != nil {
		return nil, err
	}

	report := &MigrationReport{
		Applied:          opts.Apply,
		LegacyConfigPath: legacyConfigPath,
		JotConfigPath:    jotConfigPath,
		GlobalConfig: FileMigration{
			From: legacyConfigPath,
			To:   jotConfigPath,
		},
	}

	notebooks, parseWarn, err := s.loadNotebookPaths(legacyConfigPath, jotConfigPath)
	if err != nil {
		return nil, err
	}
	if parseWarn != "" {
		report.Warnings = append(report.Warnings, parseWarn)
	}

	report.GlobalConfig, err = migrateGlobalConfig(legacyConfigPath, jotConfigPath, opts.Apply)
	if err != nil {
		return nil, err
	}

	report.NotebookConfigs, err = migrateNotebookConfigs(notebooks, opts.Apply, opts.Registry)
	if err != nil {
		return nil, err
	}

	report.LegacyEnvVars = detectLegacyEnvVars()
	report.ProfileReferences = detectProfileReferences()

	if len(report.LegacyEnvVars) > 0 {
		report.Warnings = append(report.Warnings, "legacy OPENNOTES_* env vars are set in current shell")
	}
	if len(report.ProfileReferences) > 0 {
		report.Warnings = append(report.Warnings, "legacy OPENNOTES_/opennotes references found in shell profile files")
	}

	return report, nil
}

func migrationConfigPaths() (string, string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve user config dir: %w", err)
	}
	legacy := filepath.Join(configDir, "opennotes", "config.json")
	jot := filepath.Join(configDir, "jot", "config.json")
	return legacy, jot, nil
}

func (s *MigrationService) loadNotebookPaths(legacyConfigPath, jotConfigPath string) ([]string, string, error) {
	paths := []string{}
	warnings := []string{}

	legacyPaths, warn, err := readNotebookPathsFromConfig(legacyConfigPath)
	if err != nil {
		return nil, "", err
	}
	if warn != "" {
		warnings = append(warnings, warn)
	}
	paths = append(paths, legacyPaths...)

	jotPaths, warn, err := readNotebookPathsFromConfig(jotConfigPath)
	if err != nil {
		return nil, "", err
	}
	if warn != "" {
		warnings = append(warnings, warn)
	}
	paths = append(paths, jotPaths...)

	seen := map[string]struct{}{}
	uniq := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		uniq = append(uniq, p)
	}
	sort.Strings(uniq)
	return uniq, strings.Join(warnings, "; "), nil
}

func readNotebookPathsFromConfig(path string) ([]string, string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, "", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Sprintf("could not parse notebooks from %s", path), nil
	}

	return cfg.Notebooks, "", nil
}

func migrateGlobalConfig(legacyConfigPath, jotConfigPath string, apply bool) (FileMigration, error) {
	result := FileMigration{From: legacyConfigPath, To: jotConfigPath}

	if _, err := os.Stat(legacyConfigPath); os.IsNotExist(err) {
		result.Status = "missing-source"
		return result, nil
	}

	if _, err := os.Stat(jotConfigPath); err == nil {
		result.Status = "skipped-target-exists"
		return result, nil
	}

	if !apply {
		result.Status = "would-migrate"
		return result, nil
	}

	data, err := os.ReadFile(legacyConfigPath)
	if err != nil {
		return result, fmt.Errorf("failed to read legacy config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(jotConfigPath), 0755); err != nil {
		return result, fmt.Errorf("failed to create jot config directory: %w", err)
	}
	if err := os.WriteFile(jotConfigPath, data, 0644); err != nil {
		return result, fmt.Errorf("failed to write jot config: %w", err)
	}

	result.Status = "migrated"
	return result, nil
}

func migrateNotebookConfigs(notebooks []string, apply bool, registry Registry) ([]FileMigration, error) {
	if registry == nil {
		return migrateNotebookConfigsLegacy(notebooks, apply)
	}

	results := make([]FileMigration, 0, len(notebooks))
	latest := latestMigrationVersion(registry)
	for _, notebookPath := range notebooks {
		from := filepath.Join(notebookPath, legacyNotebookConfigFile)
		to := filepath.Join(notebookPath, NotebookConfigFile)
		result := FileMigration{From: from, To: to}

		current := detectNotebookVersion(from, to)
		if current == 0 && !fileExists(from) {
			result.Status = "missing-source"
			results = append(results, result)
			continue
		}

		_, err := Execute(context.Background(), registry, Context{NotebookPath: notebookPath, DryRun: !apply}, current, latest)
		if err != nil {
			return nil, fmt.Errorf("failed migration plan for %s: %w", notebookPath, err)
		}

		if apply {
			if current == latest {
				result.Status = "already-current"
			} else {
				result.Status = "migrated"
			}
		} else {
			if current == latest {
				result.Status = "already-current"
			} else {
				result.Status = "would-migrate"
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func migrateNotebookConfigsLegacy(notebooks []string, apply bool) ([]FileMigration, error) {
	results := make([]FileMigration, 0, len(notebooks))

	for _, notebookPath := range notebooks {
		from := filepath.Join(notebookPath, legacyNotebookConfigFile)
		to := filepath.Join(notebookPath, NotebookConfigFile)
		result := FileMigration{From: from, To: to}

		if _, err := os.Stat(from); os.IsNotExist(err) {
			result.Status = "missing-source"
			results = append(results, result)
			continue
		}

		if _, err := os.Stat(to); err == nil {
			result.Status = "skipped-target-exists"
			results = append(results, result)
			continue
		}

		if !apply {
			result.Status = "would-migrate"
			results = append(results, result)
			continue
		}

		if err := os.Rename(from, to); err != nil {
			return nil, fmt.Errorf("failed to rename notebook config %s -> %s: %w", from, to, err)
		}
		result.Status = "migrated"
		results = append(results, result)
	}

	return results, nil
}

func latestMigrationVersion(reg Registry) Version {
	migrations := reg.List()
	if len(migrations) == 0 {
		return 0
	}
	latest := Version(0)
	for _, m := range migrations {
		if m.Metadata().To > latest {
			latest = m.Metadata().To
		}
	}
	return latest
}

func detectNotebookVersion(legacyPath, jotPath string) Version {
	legacy := fileExists(legacyPath)
	current := fileExists(jotPath)
	if current {
		return 1
	}
	if legacy {
		return 0
	}
	return 0
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func detectLegacyEnvVars() []string {
	matches := make([]string, 0)
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 0 {
			continue
		}
		if strings.HasPrefix(parts[0], "OPENNOTES_") {
			matches = append(matches, parts[0])
		}
	}
	sort.Strings(matches)
	return matches
}

func detectProfileReferences() []ProfileReference {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".zprofile"),
		filepath.Join(home, ".profile"),
		filepath.Join(home, ".config", "fish", "config.fish"),
	}

	results := []ProfileReference{}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		matches := []string{}
		for i, line := range lines {
			if strings.Contains(line, "OPENNOTES_") || strings.Contains(line, "opennotes") {
				matches = append(matches, fmt.Sprintf("line %d: %s", i+1, strings.TrimSpace(line)))
			}
		}
		if len(matches) == 0 {
			continue
		}

		results = append(results, ProfileReference{
			File:    path,
			Matches: matches,
		})
	}

	return results
}

var (
	ErrInvalidMigrationID    = errors.New("invalid migration id")
	ErrInvalidVersionRange   = errors.New("invalid migration version range")
	ErrMissingDescription    = errors.New("missing migration description")
	ErrInvalidMigrationChain = errors.New("invalid migration chain")
	ErrMigrationPathNotFound = errors.New("migration path not found")
)

var migrationIDPattern = regexp.MustCompile(`^\d{5}_[a-z0-9_]+$`)

// Version identifies a notebook config schema version.
type Version uint32

// Context carries execution context for migration runs.
type Context struct {
	NotebookPath string
	DryRun       bool
}

// Metadata defines migration identity and version boundaries.
type Metadata struct {
	ID          string
	From        Version
	To          Version
	Description string
}

// Validate enforces metadata contract rules.
func (m Metadata) Validate() error {
	if !migrationIDPattern.MatchString(m.ID) {
		return fmt.Errorf("%w: %q", ErrInvalidMigrationID, m.ID)
	}
	if m.To <= m.From {
		return fmt.Errorf("%w: from=%d to=%d", ErrInvalidVersionRange, m.From, m.To)
	}
	if strings.TrimSpace(m.Description) == "" {
		return ErrMissingDescription
	}
	return nil
}

// Migration defines a reversible notebook migration step.
type Migration interface {
	Metadata() Metadata
	Up(ctx context.Context, req Context) error
	Down(ctx context.Context, req Context) error
}

// Registry defines migration registration and lookup contract.
type Registry interface {
	Register(Migration) error
	Get(id string) (Migration, bool)
	List() []Migration
	Validate() error
}

// MigrationRegistry is the default in-memory registry implementation.
type MigrationRegistry struct {
	byID    map[string]Migration
	ordered []Migration
}

// NewMigrationRegistry creates an empty migration registry.
func NewMigrationRegistry() *MigrationRegistry {
	return &MigrationRegistry{byID: map[string]Migration{}, ordered: []Migration{}}
}

// Register adds one migration to the registry.
func (r *MigrationRegistry) Register(m Migration) error {
	if m == nil {
		return fmt.Errorf("nil migration")
	}
	meta := m.Metadata()
	if err := meta.Validate(); err != nil {
		return err
	}
	if _, exists := r.byID[meta.ID]; exists {
		return fmt.Errorf("duplicate migration id: %s", meta.ID)
	}
	r.byID[meta.ID] = m
	r.ordered = append(r.ordered, m)
	return nil
}

// Get returns migration by ID.
func (r *MigrationRegistry) Get(id string) (Migration, bool) {
	m, ok := r.byID[id]
	return m, ok
}

// List returns all registered migrations sorted by source version.
func (r *MigrationRegistry) List() []Migration {
	out := make([]Migration, len(r.ordered))
	copy(out, r.ordered)
	sort.Slice(out, func(i, j int) bool {
		mi := out[i].Metadata()
		mj := out[j].Metadata()
		if mi.From == mj.From {
			return mi.ID < mj.ID
		}
		return mi.From < mj.From
	})
	return out
}

// Validate validates registry migration chain.
func (r *MigrationRegistry) Validate() error {
	return ValidateChain(r.List())
}

// Direction indicates whether a migration step should run Up or Down.
type Direction string

const (
	DirectionUp   Direction = "up"
	DirectionDown Direction = "down"
)

// PlannedStep is one migration action in an execution plan.
type PlannedStep struct {
	Migration Migration
	Direction Direction
}

// Plan contains ordered migration steps between current and target versions.
type Plan struct {
	Current Version
	Target  Version
	Steps   []PlannedStep
}

// BuildPlan creates an ordered migration plan from current to target.
func BuildPlan(reg Registry, current, target Version) (Plan, error) {
	if reg == nil {
		return Plan{}, fmt.Errorf("%w: nil registry", ErrInvalidMigrationChain)
	}

	if err := validateRegistryChain(reg); err != nil {
		return Plan{}, err
	}

	return buildPlanFromMigrations(reg.List(), current, target)
}

// Execute validates migration chain, plans path, and executes steps in order.
func Execute(ctx context.Context, reg Registry, req Context, current, target Version) (Plan, error) {
	plan, err := BuildPlan(reg, current, target)
	if err != nil {
		return Plan{}, err
	}

	for _, step := range plan.Steps {
		var runErr error
		switch step.Direction {
		case DirectionUp:
			runErr = step.Migration.Up(ctx, req)
		case DirectionDown:
			runErr = step.Migration.Down(ctx, req)
		default:
			runErr = fmt.Errorf("unknown migration direction %q", step.Direction)
		}

		if runErr != nil {
			meta := step.Migration.Metadata()
			return Plan{}, fmt.Errorf("execute migration %s (%s): %w", meta.ID, step.Direction, runErr)
		}
	}

	return plan, nil
}

func validateRegistryChain(reg Registry) error {
	if err := reg.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidMigrationChain, err)
	}

	if err := ValidateChain(reg.List()); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidMigrationChain, err)
	}

	return nil
}

// ValidateChain enforces a strict linear migration chain.
func ValidateChain(migrations []Migration) error {
	if len(migrations) == 0 {
		return nil
	}

	metas := make([]Metadata, 0, len(migrations))
	for _, migration := range migrations {
		meta := migration.Metadata()
		if err := meta.Validate(); err != nil {
			return err
		}
		if meta.To != meta.From+1 {
			return fmt.Errorf("migration %s must advance exactly one version: from=%d to=%d", meta.ID, meta.From, meta.To)
		}
		metas = append(metas, meta)
	}

	sort.Slice(metas, func(i, j int) bool {
		if metas[i].From == metas[j].From {
			return metas[i].ID < metas[j].ID
		}
		return metas[i].From < metas[j].From
	})

	seenFrom := make(map[Version]string, len(metas))
	seenTo := make(map[Version]string, len(metas))
	for _, meta := range metas {
		if prevID, exists := seenFrom[meta.From]; exists {
			return fmt.Errorf("duplicate migration source version %d: %s and %s", meta.From, prevID, meta.ID)
		}
		if prevID, exists := seenTo[meta.To]; exists {
			return fmt.Errorf("duplicate migration target version %d: %s and %s", meta.To, prevID, meta.ID)
		}
		seenFrom[meta.From] = meta.ID
		seenTo[meta.To] = meta.ID
	}

	for i := 1; i < len(metas); i++ {
		if metas[i].From != metas[i-1].To {
			return fmt.Errorf("migration gap between %s (%d->%d) and %s (%d->%d)",
				metas[i-1].ID, metas[i-1].From, metas[i-1].To,
				metas[i].ID, metas[i].From, metas[i].To,
			)
		}
	}

	return nil
}

func buildPlanFromMigrations(migrations []Migration, current, target Version) (Plan, error) {
	plan := Plan{Current: current, Target: target, Steps: []PlannedStep{}}
	if current == target {
		return plan, nil
	}

	byFrom := make(map[Version]Migration, len(migrations))
	byTo := make(map[Version]Migration, len(migrations))
	for _, migration := range migrations {
		meta := migration.Metadata()
		byFrom[meta.From] = migration
		byTo[meta.To] = migration
	}

	if target > current {
		version := current
		for version < target {
			migration, ok := byFrom[version]
			if !ok {
				return Plan{}, fmt.Errorf("%w: cannot migrate up from version %d", ErrMigrationPathNotFound, version)
			}
			plan.Steps = append(plan.Steps, PlannedStep{Migration: migration, Direction: DirectionUp})
			version = migration.Metadata().To
		}
		return plan, nil
	}

	version := current
	for version > target {
		migration, ok := byTo[version]
		if !ok {
			return Plan{}, fmt.Errorf("%w: cannot migrate down from version %d", ErrMigrationPathNotFound, version)
		}
		plan.Steps = append(plan.Steps, PlannedStep{Migration: migration, Direction: DirectionDown})
		version = migration.Metadata().From
	}

	return plan, nil
}
