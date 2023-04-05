package cli

import (
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth [command]",
	Short: "Interact with authentication resources",
	Long: `This command allows interacting with authentication resources to manage auth
clients, users, and more.

The usage of this command assumes that the person using it has the necessary
permissions to perform the actions. No authorization or authentication is
performed by this command.`,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
