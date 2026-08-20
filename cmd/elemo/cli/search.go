package cli

import (
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [command]",
	Short: "Manage the search index",
	Long: `This command administers the searchable projection stored in Meilisearch.

The usage of this command assumes that the person using it has the necessary
permissions to perform the actions. No authorization or authentication is
performed by this command.`,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
