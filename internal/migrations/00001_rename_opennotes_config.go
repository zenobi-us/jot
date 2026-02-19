package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zenobi-us/jot/internal/services"
)

const (
	legacyNotebookConfigFile = ".opennotes.json"
)

// Migration00001RenameOpenNotesConfig renames legacy notebook config to .jot.json.
type Migration00001RenameOpenNotesConfig struct{}

func (m Migration00001RenameOpenNotesConfig) Metadata() services.Metadata {
	return services.Metadata{
		ID:          "00001_rename_opennotes_config",
		From:        0,
		To:          1,
		Description: "Rename notebook config from .opennotes.json to .jot.json",
	}
}

func (m Migration00001RenameOpenNotesConfig) Up(ctx context.Context, req services.Context) error {
	_ = ctx
	from := filepath.Join(req.NotebookPath, legacyNotebookConfigFile)
	to := filepath.Join(req.NotebookPath, services.NotebookConfigFile)
	return renameConfigIfNeeded(from, to, req.DryRun)
}

func (m Migration00001RenameOpenNotesConfig) Down(ctx context.Context, req services.Context) error {
	_ = ctx
	from := filepath.Join(req.NotebookPath, services.NotebookConfigFile)
	to := filepath.Join(req.NotebookPath, legacyNotebookConfigFile)
	return renameConfigIfNeeded(from, to, req.DryRun)
}

func renameConfigIfNeeded(from, to string, dryRun bool) error {
	fromExists := fileExists(from)
	toExists := fileExists(to)

	if !fromExists {
		if toExists {
			return nil
		}
		return nil
	}

	if toExists {
		return fmt.Errorf("target config already exists: %s", to)
	}

	if dryRun {
		return nil
	}

	if err := os.Rename(from, to); err != nil {
		return fmt.Errorf("rename config %s -> %s: %w", from, to, err)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
