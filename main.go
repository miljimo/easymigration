package main

import (
	"context"
	"fmt"
	"syscall"

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

var rootCmd = &cobra.Command{
	Use:  "dm",
	Long: "migration tools for databases ",
	Run: func(cmd *cobra.Command, args []string) {

		cmd.Help()
	},
}

// The command to start
var migrationStart = &cobra.Command{
	Use:   "start",
	Short: "migrate the sql files into databases objects",
	Long:  "starts the migration process",
	RunE: func(cmd *cobra.Command, args []string) error {

		filename, err := cmd.Flags().GetString("file")
		if err != nil {
			cmd.Help()
			return err
		}
		if reader.Exist(filename) != true {
			return fmt.Errorf("migration configuration doesn't exists ,  filename = %s", filename)
		}
		return runMigration(filename)
	},
}

func init() {
	migrationStart.Flags().StringP("file", "f", "", "The configuration file to load")
	rootCmd.AddCommand(migrationStart)
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}

//entry point of the command

func main() {
	err := Execute()
	if err != nil {
		syscall.Exit(-1)
	}
}
