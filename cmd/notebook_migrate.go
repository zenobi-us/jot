package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/migrations"
	"github.com/zenobi-us/jot/internal/services"
)

var notebookMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Inspect and apply notebook/config migrations",
	Long: `Inspect and apply migration steps for Jot configuration data.

This command supports dry-run preview by default, and can apply detected
migrations with --apply.

Use 'jot notebook migrate list' to see migration status and available steps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apply, _ := cmd.Flags().GetBool("apply")

		registry, err := buildMigrationRegistry()
		if err != nil {
			return err
		}

		svc := services.NewMigrationService()
		report, err := svc.MigrateOpenNotesToJot(services.MigrationOptions{Apply: apply, Registry: registry})
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		printMigrationReport(report)
		return nil
	},
}

var notebookMigrateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List migration status or available migration steps",
	RunE: func(cmd *cobra.Command, args []string) error {
		registry, err := buildMigrationRegistry()
		if err != nil {
			return err
		}

		explicitNotebook, _ := cmd.Flags().GetString("notebook")
		report, err := buildMigrationListStatusReport(registry, explicitNotebook)
		if err != nil {
			return err
		}

		fmt.Print(renderMigrationStatusReport(report, "all", strings.TrimSpace(explicitNotebook) != ""))
		return nil
	},
}

func init() {
	notebookMigrateCmd.Flags().Bool("apply", false, "Apply changes (default is dry-run)")

	notebookMigrateCmd.AddCommand(notebookMigrateListCmd)
	notebookCmd.AddCommand(notebookMigrateCmd)
}

func buildMigrationRegistry() (*services.MigrationRegistry, error) {
	registry := services.NewMigrationRegistry()
	if err := registry.Register(migrations.Migration00001RenameOpenNotesConfig{}); err != nil {
		return nil, fmt.Errorf("failed to register built-in migrations: %w", err)
	}
	return registry, nil
}

func printMigrationReport(report *services.MigrationReport) {
	_ = printMigrationReportTo(os.Stdout, report)
}

func printMigrationReportTo(w io.Writer, report *services.MigrationReport) error {
	mode := "DRY RUN"
	if report.Applied {
		mode = "APPLY"
	}

	writef := func(format string, args ...any) error {
		_, err := fmt.Fprintf(w, format, args...)
		return err
	}
	writeln := func(s string) error {
		_, err := fmt.Fprintln(w, s)
		return err
	}

	if err := writef("Notebook migration (%s)\n", mode); err != nil {
		return err
	}
	if err := writef("Global config version: v%d -> v%d\n", report.GlobalConfig.CurrentVersion, report.GlobalConfig.TargetVersion); err != nil {
		return err
	}
	if err := writef("Global config: %s\n", renderFileMigration(report.GlobalConfig)); err != nil {
		return err
	}
	if !report.Applied {
		if err := writef("Notebook target version: v%d\n", notebookTargetVersion(report)); err != nil {
			return err
		}
	}

	if err := writeln("Notebook config migrations:"); err != nil {
		return err
	}
	if len(report.NotebookConfigs) == 0 {
		if err := writeln("  - no notebook paths found in config"); err != nil {
			return err
		}
	} else {
		for _, item := range report.NotebookConfigs {
			if err := writef("  - %s\n", renderFileMigration(item)); err != nil {
				return err
			}
		}
	}

	if len(report.LegacyEnvVars) > 0 {
		if err := writeln("Detected legacy env vars in current shell:"); err != nil {
			return err
		}
		for _, name := range report.LegacyEnvVars {
			if err := writef("  - %s\n", name); err != nil {
				return err
			}
		}
	}

	if len(report.ProfileReferences) > 0 {
		if err := writeln("Detected legacy references in shell profiles:"); err != nil {
			return err
		}
		for _, ref := range report.ProfileReferences {
			if err := writef("  - %s\n", ref.File); err != nil {
				return err
			}
			for _, match := range ref.Matches {
				if err := writef("    * %s\n", match); err != nil {
					return err
				}
			}
		}
	}

	if len(report.Warnings) > 0 {
		if err := writeln("Warnings:"); err != nil {
			return err
		}
		for _, warning := range report.Warnings {
			if err := writef("  - %s\n", warning); err != nil {
				return err
			}
		}
	}

	if !report.Applied {
		if err := writeln("Tip: run with --apply to execute the migration."); err != nil {
			return err
		}
	}

	return nil
}

func notebookTargetVersion(report *services.MigrationReport) services.Version {
	target := report.GlobalConfig.TargetVersion
	for _, item := range report.NotebookConfigs {
		if item.TargetVersion > target {
			target = item.TargetVersion
		}
	}
	return target
}

func renderMigrationCatalog(registry services.Registry, scope string) string {
	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "" {
		scope = "all"
	}

	notebookScope := scope == "all" || scope == "registered" || scope == "current" || scope == "notebook"
	globalScope := scope == "all" || scope == "global"

	var b strings.Builder
	if globalScope {
		b.WriteString("Global migrations\n")
		b.WriteString("  - global-config-copy: 0 -> 1 (Copy ~/.config/opennotes/config.json to ~/.config/jot/config.json)\n")
	}

	if notebookScope {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("Notebook migrations\n")
		for _, migration := range registry.List() {
			meta := migration.Metadata()
			b.WriteString(fmt.Sprintf("  - %s: %d -> %d (%s)\n", meta.ID, meta.From, meta.To, meta.Description))
		}
	}

	if b.Len() == 0 {
		b.WriteString("No migrations for scope: ")
		b.WriteString(scope)
		b.WriteString("\n")
	}

	return b.String()
}

type migrationGlobalStatusRow struct {
	Path           string
	CurrentVersion services.Version
	TargetVersion  services.Version
	Status         string
	PlannedSteps   int
}

type migrationNotebookStatusRow struct {
	Scope          string
	NotebookPath   string
	CurrentVersion services.Version
	TargetVersion  services.Version
	Status         string
	PlannedSteps   int
	Reason         string
}

type migrationListStatusReport struct {
	Global              *migrationGlobalStatusRow
	CurrentNotebook     *migrationNotebookStatusRow
	RegisteredNotebooks []migrationNotebookStatusRow
	Notices             []string
}

func buildMigrationListStatusReport(registry services.Registry, explicitNotebook string) (migrationListStatusReport, error) {
	report := migrationListStatusReport{}
	explicitNotebook = strings.TrimSpace(explicitNotebook)
	if explicitNotebook != "" {
		path, err := filepath.Abs(explicitNotebook)
		if err != nil {
			path = explicitNotebook
		}
		row := assessNotebookMigration(path, "current", registry)
		report.CurrentNotebook = &row
		return report, nil
	}

	globalRow, err := assessGlobalMigration()
	if err != nil {
		return report, err
	}
	report.Global = &globalRow

	if notebookService != nil {
		nb, err := notebookService.Infer("")
		if err != nil {
			report.Notices = append(report.Notices, fmt.Sprintf("Could not resolve current notebook: %v", err))
		} else if nb != nil {
			row := assessNotebookMigration(nb.Config.Root, "current", registry)
			report.CurrentNotebook = &row
		}
	}

	seen := map[string]struct{}{}
	currentPath := ""
	if report.CurrentNotebook != nil {
		currentPath = report.CurrentNotebook.NotebookPath
	}

	if cfgService != nil {
		for _, path := range cfgService.Store.Notebooks {
			trimmed := strings.TrimSpace(path)
			if trimmed == "" {
				continue
			}
			absPath, err := filepath.Abs(trimmed)
			if err != nil {
				absPath = trimmed
			}
			if absPath == currentPath {
				continue
			}
			if _, ok := seen[absPath]; ok {
				continue
			}
			seen[absPath] = struct{}{}
			report.RegisteredNotebooks = append(report.RegisteredNotebooks, assessNotebookMigration(absPath, "registered", registry))
		}
	}

	return report, nil
}

func assessGlobalMigration() (migrationGlobalStatusRow, error) {
	legacyConfigPath, jotConfigPath, err := migrationConfigPathsForList()
	if err != nil {
		return migrationGlobalStatusRow{}, err
	}

	legacyExists := pathExists(legacyConfigPath)
	jotExists := pathExists(jotConfigPath)

	row := migrationGlobalStatusRow{
		Path:          jotConfigPath,
		TargetVersion: 1,
	}
	if jotExists {
		row.CurrentVersion = 1
	}

	switch {
	case !legacyExists && !jotExists:
		row.Status = "missing-source"
	case !legacyExists && jotExists:
		row.Status = "already-current"
	case legacyExists && jotExists:
		row.Status = "already-current"
	case legacyExists && !jotExists:
		row.Status = "would-migrate"
		row.PlannedSteps = 1
	}

	return row, nil
}

func assessNotebookMigration(notebookPath, scope string, registry services.Registry) migrationNotebookStatusRow {
	from := filepath.Join(notebookPath, ".opennotes.json")
	to := filepath.Join(notebookPath, services.NotebookConfigFile)
	legacyExists := pathExists(from)
	jotExists := pathExists(to)
	target := latestNotebookTargetVersion(registry)
	current := services.Version(0)
	if jotExists {
		current = 1
	}

	row := migrationNotebookStatusRow{
		Scope:          scope,
		NotebookPath:   notebookPath,
		CurrentVersion: current,
		TargetVersion:  target,
	}

	if !legacyExists && current == 0 {
		row.Status = "missing-source"
		return row
	}

	plan, err := services.BuildPlan(registry, current, target)
	if err != nil {
		row.Status = "blocked"
		row.Reason = err.Error()
		return row
	}

	row.PlannedSteps = len(plan.Steps)
	if row.PlannedSteps == 0 {
		row.Status = "already-current"
	} else {
		row.Status = "would-migrate"
	}
	return row
}

func latestNotebookTargetVersion(reg services.Registry) services.Version {
	if reg == nil {
		return 0
	}
	latest := services.Version(0)
	for _, m := range reg.List() {
		meta := m.Metadata()
		if meta.To > latest {
			latest = meta.To
		}
	}
	return latest
}

func migrationConfigPathsForList() (string, string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve user config dir: %w", err)
	}
	legacy := filepath.Join(configDir, "opennotes", "config.json")
	jot := filepath.Join(configDir, "jot", "config.json")
	return legacy, jot, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func renderMigrationStatusReport(report migrationListStatusReport, scope string, explicitNotebook bool) string {
	var b strings.Builder

	includeGlobal := !explicitNotebook && (scope == "all" || scope == "global")
	includeCurrent := scope == "all" || scope == "current"
	includeRegistered := !explicitNotebook && (scope == "all" || scope == "registered")

	if includeGlobal {
		b.WriteString("Global migration status\n")
		if report.Global == nil {
			b.WriteString("- unavailable\n")
		} else {
			b.WriteString(fmt.Sprintf("- path: %s\n", report.Global.Path))
			b.WriteString(fmt.Sprintf("  version: v%d -> v%d\n", report.Global.CurrentVersion, report.Global.TargetVersion))
			b.WriteString(fmt.Sprintf("  status: %s\n", report.Global.Status))
			b.WriteString(fmt.Sprintf("  steps: %d\n", report.Global.PlannedSteps))
		}
	}

	if includeCurrent || includeRegistered {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("Notebook migration status\n")
		if includeCurrent {
			if report.CurrentNotebook != nil {
				b.WriteString(renderNotebookStatusBlock(*report.CurrentNotebook))
			} else {
				b.WriteString("- scope: current\n")
				b.WriteString("  status: No current notebook detected\n")
			}
		}
		if includeRegistered {
			if len(report.RegisteredNotebooks) == 0 {
				b.WriteString("- scope: registered\n")
				b.WriteString("  status: No registered notebooks found\n")
			} else {
				for _, row := range report.RegisteredNotebooks {
					b.WriteString(renderNotebookStatusBlock(row))
				}
			}
		}
	}

	if len(report.Notices) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("Notices\n")
		for _, notice := range report.Notices {
			b.WriteString("- ")
			b.WriteString(notice)
			b.WriteString("\n")
		}
	}

	if b.Len() == 0 {
		b.WriteString("No migrations for scope: ")
		b.WriteString(scope)
		b.WriteString("\n")
	}

	return b.String()
}

func renderNotebookStatusBlock(row migrationNotebookStatusRow) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("- scope: %s\n", row.Scope))
	b.WriteString(fmt.Sprintf("  path: %s\n", row.NotebookPath))
	b.WriteString(fmt.Sprintf("  version: v%d -> v%d\n", row.CurrentVersion, row.TargetVersion))
	b.WriteString(fmt.Sprintf("  status: %s\n", row.Status))
	b.WriteString(fmt.Sprintf("  steps: %d\n", row.PlannedSteps))
	if row.Reason != "" {
		b.WriteString(fmt.Sprintf("  reason: %s\n", row.Reason))
	}
	return b.String()
}

func renderFileMigration(item services.FileMigration) string {
	return fmt.Sprintf("[%s] %s -> %s (v%d -> v%d)", item.Status, item.From, item.To, item.CurrentVersion, item.TargetVersion)
}
