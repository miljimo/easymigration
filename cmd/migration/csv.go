package migration

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var csvCommand = &cobra.Command{
	Use:   "ingestion",
	Short: "Ingest csv files into database table",
	Long:  "The command allow the ability to ingest",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := cmd.Flags().GetString("path")
		if err != nil {
			return err
		}

		if strings.Trim(path, " ") == "" {
			return fmt.Errorf("path must be provided to the data source")
		}

		fmt.Println("Command: to ingest data into the database")

		return nil
	},
}

func init() {
	csvCommand.Flags().String("path",
		"",
		"path to the file or directory of csv files to ingest into the database")

	RootCommand().AddCommand(csvCommand)

}
