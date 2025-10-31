package cli

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"

	"log/slog"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and exit",
	Long:  `Prints the version of the application and exits.`,
	Run: func(_ *cobra.Command, _ []string) {
		b, err := json.Marshal(versionInfo)
		if err != nil {
			logger.Panic(context.Background(), "failed to marshal version info", slog.Any("error", err))
		}

		fmt.Println(string(b))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
