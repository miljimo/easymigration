package migration

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "obz",
	Long: "A simple migration tools",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func RootCommand() *cobra.Command {
	return rootCmd
}
