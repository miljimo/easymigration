package migration

import (
	"context"
	"errors"
	"strings"

	"github.com/miljimo/easymigration/internal/config"
	"github.com/miljimo/easymigration/internal/exporter"
	"github.com/miljimo/easymigration/internal/reader"
	"github.com/spf13/cobra"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "ingest csv files into the databases",
	Long:  "A command line tool that enable ingestion of data into the database",

	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := cmd.Flags().GetString("path")
		if err != nil {
			return err
		}
		if strings.Trim(path, " ") == " " {
			return errors.New("path must be provided")
		}

		if !reader.Exist(path) {
			return errors.New("configuration path does not exists")
		}

		conf, err := config.OpenConfigFile(path)
		if err != nil {
			return err
		}

		exporter.ExportAll(context.Background(), conf)

		return nil

	},
}

func init() {

	ingestCmd.Flags().StringP("path", "p", "", "The path to the configuration file where the csv files are located")
	RootCommand().AddCommand(ingestCmd)
}
