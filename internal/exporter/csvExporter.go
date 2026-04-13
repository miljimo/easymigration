package exporter

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/miljimo/easymigration/internal/data"
	"github.com/miljimo/easymigration/internal/reader"
)

type Exporter interface {
	Export(table data.Table) error
}

type imptCSVExporter struct {
	db data.DataContext
}

func (ex *imptCSVExporter) tableExist(tableName string) bool {
	query := fmt.Sprintf(
		"SELECT * FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
		ex.db.Name(), tableName,
	)

	records, err := ex.db.Query(context.Background(), query)
	if err != nil {
		log.Println("Error checking if table exists:", err)
		return false
	}
	return len(records) > 0
}

func (ex *imptCSVExporter) dropTable(tableName string) error {
	query := fmt.Sprintf("DROP TABLE `%s`", tableName)
	_, err := ex.db.Query(context.Background(), query)
	return err
}

func (ex *imptCSVExporter) createTable(tableName string, headers []data.Column) error {
	definitions := make([]string, 0, len(headers))
	for _, col := range headers {
		dbType := "VARCHAR(100)"

		if col.Type() == data.DBBool {
			dbType = "BOOL"
		}
		if col.Type() == data.DBDatetime {
			dbType = "TIMESTAMP"
		}

		if col.Type() == data.DBFloat {
			dbType = "FLOAT"
		}

		if col.Type() == data.DBInteger {
			dbType = "INT"
		}

		definitions = append(definitions, fmt.Sprintf("`%s` %s", col.Name(), dbType))
	}
	columns := strings.Join(definitions, ", ")
	query := fmt.Sprintf("CREATE TABLE `%s` (%s)", tableName, columns)

	_, err := ex.db.Query(context.Background(), query)
	return err
}

func (ex *imptCSVExporter) createRecords(tableName string, df data.Table) error {
	headers := make([]string, 0, len(df.Headers().ToStringList()))
	for _, header := range df.Headers().Columns() {
		headers = append(headers, fmt.Sprintf("`%s`", header.Name()))
	}

	var (
		valueStrings []string
		valueArgs    []interface{}
	)

	for _, row := range df.Rows() {
		placeholders := make([]string, row.Size())
		items := row.ToList()
		for i, value := range items {
			placeholders[i] = "?"
			valueArgs = append(valueArgs, value)
		}
		valueStrings = append(valueStrings, fmt.Sprintf("(%s)", strings.Join(placeholders, ",")))
	}

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s",
		tableName,
		strings.Join(headers, ", "),
		strings.Join(valueStrings, ", "),
	)
	fmt.Println("SQL = ")
	fmt.Println(query)
	fmt.Println(" =")

	_, err := ex.db.Execute(context.Background(), query, valueArgs...)
	return err
}

// The function will export the table data into the database
func Export(cxt context.Context, dataContext data.DataContext, df data.Table) error {
	if df == nil || df.RowCounts() == 0 {
		return fmt.Errorf("no data to export")
	}

	exporter := &imptCSVExporter{db: dataContext}

	if !exporter.tableExist(df.Name()) {
		exporter.createTable(df.Name(), df.Headers().Columns())
	}
	return exporter.createRecords(df.Name(), df)

}

func ExportAll(cxt context.Context, dataContext data.DataContext, folderPath string, callback func(data.Table) (data.Table, error)) error {

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".csv") {
			df, err := reader.ReadCSV(path)
			if err != nil {
				return err
			}
			if callback != nil {
				df, err = callback(df)
				if err != nil {
					return fmt.Errorf("failed:  %s: %w", info.Name(), err)
				}
			}
			return Export(cxt, dataContext, df)
		}
		return nil
	})
	return err
}
