package config

import (
	"context"
	"fmt"
	"sort"

	"github.com/miljimo/easymigration/internal/environment"
	"github.com/miljimo/easymigration/internal/sql/data"
)

type Database interface {
	Prepare()
	Migrate(cxt context.Context, dbContext data.DataContext)
}

type DatabaseData struct {
	Name    string       `json:"name"`
	Scripts []ScriptData `json:"scripts"`
	Skip    bool         `json:"skip"`

	// Privates fields
	environ        environment.Environment
	configRootPath string
}

func (dbConfig *DatabaseData) Prepare() {
	sort.Slice(dbConfig.Scripts, func(a, b int) bool {
		return dbConfig.Scripts[a].Order < dbConfig.Scripts[b].Order
	})

	// Prepare it.
	for _, script := range dbConfig.Scripts {
		script.Prepare(dbConfig.configRootPath)
	}

}

func (dbConfig *DatabaseData) Migrate(cxt context.Context, dbContext data.DataContext) error {
	dbConfig.Prepare()
	for _, script := range dbConfig.Scripts {
		err := script.Execute(cxt, dbContext, func(content string) (string, error) {
			// Additional processing on the content before
			// executing it.
			bytes, err := dbConfig.environ.ReplaceAll(content)
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		})

		if err != nil {
			return fmt.Errorf("Error = %s", err)
		}
	}
	return nil
}
