package cli

import (
	"github.com/gleanerio/nabu/internal/millers"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
)

var millCmd = &cobra.Command{
	Use:   "mill",
	Short: "Run the miller to process harvested JSON-LD data",
	Long: `Mill processes the harvested JSON-LD data through the configured
milling pipeline. This converts JSON-LD to RDF (N-Quads) using the shared
graph conversion code, and optionally runs SHACL validation.`,
	Run: func(cmd *cobra.Command, args []string) {
		if viperVal == nil {
			log.Fatal("Configuration not loaded")
			os.Exit(1)
		}
		millers.Millers(mc, viperVal)
	},
}

func init() {
	rootCmd.AddCommand(millCmd)
}
