package cli

import (
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [command]",
	Short: "Handles starting the server or workers",
	Long: `
This command allows starting the server or workers. The server is the main
component of the application and handles all the requests. The workers are
responsible for performing background tasks such as sending emails.`,
}

func init() {
	rootCmd.AddCommand(startCmd)
}
