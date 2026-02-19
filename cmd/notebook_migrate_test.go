package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zenobi-us/jot/internal/migrations"
	"github.com/zenobi-us/jot/internal/services"
)

func TestPrintMigrationReport_IncludesVersionTargetsInDryRun(t *testing.T) {
	report := &services.MigrationReport{
		Applied: false,
		GlobalConfig: services.FileMigration{
			From:           "/tmp/opennotes/config.json",
			To:             "/tmp/jot/config.json",
			Status:         "would-migrate",
			CurrentVersion: 0,
			TargetVersion:  1,
		},
		NotebookConfigs: []services.FileMigration{
			{
				From:           "/tmp/book/.opennotes.json",
				To:             "/tmp/book/.jot.json",
				Status:         "would-migrate",
				CurrentVersion: 0,
				TargetVersion:  1,
			},
		},
	}

	var out bytes.Buffer
	assert.NoError(t, printMigrationReportTo(&out, report))

	content := out.String()
	assert.Contains(t, content, "Global config version: v0 -> v1")
	assert.Contains(t, content, "Notebook target version: v1")
	assert.Contains(t, content, "[would-migrate] /tmp/book/.opennotes.json -> /tmp/book/.jot.json (v0 -> v1)")
}

func TestRenderMigrationCatalog_DefaultScopeShowsGlobalAndNotebook(t *testing.T) {
	registry := services.NewMigrationRegistry()
	assert.NoError(t, registry.Register(migrations.Migration00001RenameOpenNotesConfig{}))

	output := renderMigrationCatalog(registry, "all")

	assert.Contains(t, output, "Global migrations")
	assert.Contains(t, output, "global-config-copy")
	assert.Contains(t, output, "Notebook migrations")
	assert.Contains(t, output, "00001_rename_opennotes_config")
	assert.Contains(t, output, "0 -> 1")
}

func TestRenderMigrationCatalog_GlobalScope(t *testing.T) {
	registry := services.NewMigrationRegistry()
	output := renderMigrationCatalog(registry, "global")

	assert.Contains(t, output, "Global migrations")
	assert.NotContains(t, output, "Notebook migrations")
}

func TestRenderMigrationStatusReport_IncludesRequiredNotebookFields(t *testing.T) {
	report := migrationListStatusReport{
		Global: &migrationGlobalStatusRow{
			Path:           "/home/user/.config/jot/config.json",
			CurrentVersion: 0,
			TargetVersion:  1,
			Status:         "would-migrate",
			PlannedSteps:   1,
		},
		CurrentNotebook: &migrationNotebookStatusRow{
			Scope:          "current",
			NotebookPath:   "/work/current",
			CurrentVersion: 0,
			TargetVersion:  1,
			Status:         "would-migrate",
			PlannedSteps:   1,
		},
		RegisteredNotebooks: []migrationNotebookStatusRow{
			{
				Scope:          "registered",
				NotebookPath:   "/work/registered",
				CurrentVersion: 1,
				TargetVersion:  1,
				Status:         "already-current",
				PlannedSteps:   0,
			},
		},
	}

	output := renderMigrationStatusReport(report, "all", false)

	assert.Contains(t, output, "Global migration status")
	assert.Contains(t, output, "Notebook migration status")
	assert.Contains(t, output, "path: /work/current")
	assert.Contains(t, output, "version: v0 -> v1")
	assert.Contains(t, output, "status: would-migrate")
	assert.Contains(t, output, "status: already-current")
}

func TestRenderMigrationStatusReport_NoCurrentNotebookMessage(t *testing.T) {
	report := migrationListStatusReport{}

	output := renderMigrationStatusReport(report, "current", false)

	assert.Contains(t, output, "No current notebook detected")
}

func TestNotebookMigrateListCmd_DoesNotExposeViewFlag(t *testing.T) {
	assert.Nil(t, notebookMigrateListCmd.Flags().Lookup("view"))
}

func TestNotebookMigrateListCmd_DoesNotExposeScopeFlag(t *testing.T) {
	assert.Nil(t, notebookMigrateListCmd.Flags().Lookup("scope"))
}
