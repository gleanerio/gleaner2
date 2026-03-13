package cli

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Commands for initializing and managing configuration files.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a configuration directory",
	Long: `Initialize creates a new configuration directory structure with
template configuration files for services and sources.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgDir := "configs"
		if len(args) > 0 {
			cfgDir = args[0]
		}

		// Create config directory structure
		dirs := []string{
			cfgDir,
		}
		for _, d := range dirs {
			if err := os.MkdirAll(d, os.ModePerm); err != nil {
				log.Fatal("Error creating directory:", err)
			}
		}

		fmt.Printf("Configuration directory initialized at: %s\n", cfgDir)
		fmt.Println("Create your configuration files:")
		fmt.Printf("  %s/nabu.yaml     - Nabu configuration (graph loading)\n", cfgDir)
		fmt.Printf("  %s/gleaner.yaml  - Gleaner configuration (data harvesting)\n", cfgDir)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
}
