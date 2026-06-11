package cli

import (
	"os"

	"github.com/gleanerio/gleaner2/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create release graphs (provider RDF rolled into one file)",
	Long:  `Generate releases for the indexes sources and also a master release`,
	Run: func(cmd *cobra.Command, args []string) {
		requireConfig()
		err := pkg.Release(viperVal, mc)
		if err != nil {
			log.Println(err) // was log.Fatal which seems odd
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
