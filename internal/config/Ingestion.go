package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/miljimo/easymigration/internal/data"
	"github.com/miljimo/easymigration/internal/reader"
)

type Ingestion interface {
	Files() ([]string, error)
	Tables() ([]data.Table, error)
}

type IngestionData struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Parameters []string `json:"parameters"`
	Procedure  bool     `json:"isProcedure"`
}

func (ingest *IngestionData) Files() ([]string, error) {
	if ingest == nil {
		return nil, errors.New("IngestionData,  must be initialised before it can be use")
	}
	return GetFilesWithBasePattern(ingest.Path)
}

func (ingest *IngestionData) Tables() ([]data.Table, error) {
	files, err := ingest.Files()
	if err != nil {
		return nil, err
	}
	var tables []data.Table

	for _, file := range files {

		if !reader.Exist(file) {
			return nil, fmt.Errorf("file ='%s' does not exists ", file)
		}
		table, err := reader.ReadCSV(file)
		if err != nil {
			return nil, fmt.Errorf("ReadCSV error = %s", err)
		}
		ingest.Name = strings.Trim(ingest.Name, " ")

		if ingest.Name != "" {
			table.SetName(ingest.Name)
		}
		// We want to make sure that the table name is
		if ingest.Name == "" {
			name := filepath.Base(table.Name())
			splits := strings.Split(name, ".")
			table.SetName(splits[len(splits)-1])
		}

		if len(ingest.Parameters) != 0 {
			ftable, err := table.Filters(ingest.Parameters)
			if err != nil {
				fmt.Println("Table error, invalid parameters")
				break
			}
			table = ftable
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func NewIngestion(path string) Ingestion {
	return &IngestionData{
		Path: path,
		Name: "",
	}
}

type MigrationData struct {
	DatabaseName string          `json:"name"`
	Reset        bool            `json:"reset"`
	Ingestions   []IngestionData `json:"ingestions"`
}
