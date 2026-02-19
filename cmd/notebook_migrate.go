package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/migrations"
	"github.com/zenobi-us/jot/internal/services"
)

var notebookMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate legacy notebook/config setup to the current format",
	Long: `Migrates OpenNotes-era configuration to Jot.

What this command does:
  - Copies global config: ~/.config/opennotes/config.json -> ~/.config/jot/config.json
  - Renames notebook config files: .opennotes.json -> .jot.json
  - Detects and reports legacy OPENNOTES_* environment/profile references

By default this command runs in dry-run mode and only reports changes.
Use --apply to perform migrations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apply, _ := cmd.Flags().GetBool("apply")

		registry := services.NewMigrationRegistry()
		if err := registry.Register(migrations.Migration00001RenameOpenNotesConfig{}); err != nil {
			return fmt.Errorf("failed to register built-in migrations: %w", err)
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

func init() {
	notebookMigrateCmd.Flags().Bool("apply", false, "Apply changes (default is dry-run)")
	notebookCmd.AddCommand(notebookMigrateCmd)
}

func printMigrationReport(report *services.MigrationReport) {
	mode := "DRY RUN"
	if report.Applied {
		mode = "APPLY"
	}

	fmt.Printf("Notebook migration (%s)\n", mode)
	fmt.Printf("Global config: %s\n", renderFileMigration(report.GlobalConfig))

	fmt.Println("Notebook config migrations:")
	if len(report.NotebookConfigs) == 0 {
		fmt.Println("  - no notebook paths found in config")
	} else {
		for _, item := range report.NotebookConfigs {
			fmt.Printf("  - %s\n", renderFileMigration(item))
		}
	}

	if len(report.LegacyEnvVars) > 0 {
		fmt.Println("Detected legacy env vars in current shell:")
		for _, name := range report.LegacyEnvVars {
			fmt.Printf("  - %s\n", name)
		}
	}

	if len(report.ProfileReferences) > 0 {
		fmt.Println("Detected legacy references in shell profiles:")
		for _, ref := range report.ProfileReferences {
			fmt.Printf("  - %s\n", ref.File)
			for _, match := range ref.Matches {
				fmt.Printf("    * %s\n", match)
			}
		}
	}

	if len(report.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range report.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	if !report.Applied {
		fmt.Println("Tip: run with --apply to execute the migration.")
	}
}

func renderFileMigration(item services.FileMigration) string {
	return fmt.Sprintf("[%s] %s -> %s", item.Status, item.From, item.To)
}
