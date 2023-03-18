package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and exit",
	Long:  `Prints the version of the application and exits.`,
	Run: func(cmd *cobra.Command, args []string) {
		b, err := json.Marshal(versionInfo)
		if err != nil {
			logger.Panic("failed to marshal version info", zap.Error(err))
		}

		fmt.Println(string(b))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
