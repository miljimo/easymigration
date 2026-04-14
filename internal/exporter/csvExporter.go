package exporter

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/miljimo/easymigration/internal/config"
	"github.com/miljimo/easymigration/internal/data"
)

type Exporter interface {
	Export(table data.Table) error
}

type imptCSVExporter struct {
	db data.DataContext
}

func (ex *imptCSVExporter) tableExist(cxt context.Context, tableName string) bool {
	query := fmt.Sprintf(
		"SELECT * FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'",
		ex.db.Name(), tableName,
	)

	records, err := ex.db.Query(cxt, query)
	if err != nil {
		log.Println("Error checking if table exists:", err)
		return false
	}
	return len(records) > 0
}

func (ex *imptCSVExporter) dropTable(cxt context.Context, tableName string) error {
	query := fmt.Sprintf("DROP TABLE `%s`", tableName)
	_, err := ex.db.Query(cxt, query)
	return err
}

func (ex *imptCSVExporter) createTable(cxt context.Context, tableName string, headers []data.Column) error {
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

	_, err := ex.db.Query(cxt, query)
	return err
}

func (ex *imptCSVExporter) createRecords(cxt context.Context, tableName string, df data.Table) error {
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

	_, err := ex.db.Execute(cxt, query, valueArgs...)
	return err
}

// The function will export the table data into the database
func Export(cxt context.Context, dataContext data.DataContext, df data.Table) error {
	if df == nil || df.RowCounts() == 0 {
		return fmt.Errorf("no data to export")
	}
	exporter := &imptCSVExporter{db: dataContext}
	if !exporter.tableExist(cxt, df.Name()) {
		return fmt.Errorf("SQL table %s does not exists", df.Name())
	}
	return exporter.createRecords(cxt, df.Name(), df)

}

func ExportAll(cxt context.Context, config config.Configuration) error {
	connStr, _ := config.Credential().String()
	dataCxt, err := data.WithCredential(cxt, connStr)

	if err != nil {
		return err
	}

	defer dataCxt.Close()

	for _, m := range config.Data() {
		for _, i := range m.Ingestions {
			tables, err := i.Tables()
			if err != nil {
				return err
			}

			for _, table := range tables {
				if i.Procedure != true {
					err = Export(cxt, dataCxt, table)
					if err != nil {
						return err
					}
					continue
				}
				for _, row := range table.Rows() {
					_, err = dataCxt.CallProc(cxt, table.Name(), row.ToList()...)
					if err != nil {
						return err
					}

				}

			}
		}
	}

	return nil
}
