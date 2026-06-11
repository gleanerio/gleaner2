package cli

import (
	"github.com/gleanerio/gleaner2/internal/sparql"
	"github.com/gleanerio/gleaner2/pkg"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Graph management commands (clear, drop)",
	Long:  `Commands for managing named graphs in the triplestore.`,
}

var graphClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all graphs from the triplestore",
	Long:  `Clear removes ALL graphs from the triplestore. Requires --dangerous flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		requireConfig()
		err := pkg.Clear(viperVal, mc)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	},
}

var graphDropCmd = &cobra.Command{
	Use:   "drop [graph-uri]",
	Short: "Drop a specific named graph",
	Long:  `Drop removes a specific named graph from the triplestore.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		requireConfig()
		_, err := sparql.Drop(viperVal, args[0])
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.AddCommand(graphClearCmd)
	graphCmd.AddCommand(graphDropCmd)
}
