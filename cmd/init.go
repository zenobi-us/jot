package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/services"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize jot configuration",
	Long: `Creates the jot configuration directory and default config file.

The config file is created at ~/.config/jot/config.json (or the
path specified by JOT_CONFIG environment variable).

Examples:
  # Initialize configuration
  jot init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfgService.Write(cfgService.Store); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}

		fmt.Printf("Jot initialized at %s\n", services.GlobalConfigFile())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
