package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/services"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize opennotes configuration",
	Long:  `Creates the opennotes configuration directory and default config file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfgService.Write(cfgService.Store); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}

		fmt.Printf("OpenNotes initialized at %s\n", services.GlobalConfigFile())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
