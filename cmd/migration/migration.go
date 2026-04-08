package migration

import (
	"context"
	"fmt"

	"github.com/miljimo/easymigration/internal/config"
	"github.com/miljimo/easymigration/internal/reader"
	"github.com/spf13/cobra"
)

func runMigration(filename string) error {

	err := config.Migrate(context.Background(), filename)
	if err != nil {
		return err
	}
	return nil
}

// The command to start
var migrationStart = &cobra.Command{
	Use:   "start",
	Short: "migrates sql files to mysql database",
	Long:  "A command line tool that make it easy to migrate sql and data to database",
	RunE: func(cmd *cobra.Command, args []string) error {

		filename, err := cmd.Flags().GetString("file")
		if err != nil {
			cmd.Help()
			return err
		}
		if reader.Exist(filename) != true {
			return fmt.Errorf("Migration configuration file does not exist ,  filename = %s", filename)
		}
		return runMigration(filename)
	},
}

func init() {
	migrationStart.Flags().StringP("file", "f", "", "The path to the configuration file to load")
	rootCmd.AddCommand(migrationStart)
}
