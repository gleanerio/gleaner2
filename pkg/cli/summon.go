package cli

import (
	"github.com/gleanerio/gleaner2/internal/summoner"

	"github.com/spf13/cobra"
)

var summonCmd = &cobra.Command{
	Use:   "summon",
	Short: "Run the summoner to harvest JSON-LD from data sources",
	Long: `Summon reads the configured data sources and harvests JSON-LD
structured data from websites via their sitemaps, APIs, or other methods.
The harvested data is stored in the configured S3/MinIO object store.`,
	Run: func(cmd *cobra.Command, args []string) {
		requireConfig()
		summoner.Summoner(mc, viperVal)
	},
}

func init() {
	rootCmd.AddCommand(summonCmd)
}
