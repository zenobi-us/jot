package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/services"
)

var notebookListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all notebooks",
	Long: `Lists all registered notebooks and notebooks found in ancestor directories.

Shows notebooks from:
  - Global config (~/.config/jot/config.json)
  - Ancestor directories containing .jot.json

Examples:
  # List all known notebooks
  jot notebook list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if jot has been initialized
		if !cfgService.Exists() {
			return fmt.Errorf("jot not initialized. Run 'jot init' to create config file")
		}

		notebooks, err := notebookService.List("")
		if err != nil {
			return err
		}

		if len(notebooks) == 0 {
			fmt.Println("No notebooks found.")
			fmt.Println("")
			fmt.Println("Create one with:")
			fmt.Println("  jot notebook create --name \"My Notebook\"")
			return nil
		}

		// Get current notebook for marking
		currentNotebook, _ := notebookService.Infer("")

		return displayNotebookList(notebooks, currentNotebook)
	},
}

func init() {
	notebookCmd.AddCommand(notebookListCmd)
}

func displayNotebookList(notebooks []*services.Notebook, currentNotebook *services.Notebook) error {
	output, err := services.TuiRender("notebook-list", map[string]any{
		"Notebooks":       notebooks,
		"CurrentNotebook": currentNotebook,
	})
	if err != nil {
		// Fallback to simple output
		fmt.Printf("Found %d notebook(s):\n\n", len(notebooks))
		for _, nb := range notebooks {
			fmt.Printf("  %s\n", nb.Config.Name)
			fmt.Printf("    Path: %s\n", nb.Config.Path)
			fmt.Printf("    Root: %s\n", nb.Config.Root)
			if len(nb.Config.Contexts) > 0 {
				fmt.Printf("    Contexts: %v\n", nb.Config.Contexts)
			}
			fmt.Println()
		}
		return nil
	}

	fmt.Print(output)
	return nil
}
