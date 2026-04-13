package exporter

import (
	"context"
	"testing"

	"github.com/miljimo/easymigration/internal/config"
	"github.com/miljimo/easymigration/internal/data"
	"github.com/miljimo/easymigration/internal/environment"
)

func TestExportCSVFileIntoMYSQLDatabase(t *testing.T) {
	environ := environment.New()
	environ.Set("DB_ROOT_USER", "root")
	environ.Set("DB_PASSWORD", "password")
	environ.Set("DB_HOST", "localhost")
	environ.Set("DB_USER", "webmaster")
	environ.Set("DB_PORT", "3306")
	environ.Set("DB_NAME", "email_histories_db")

	cxt := context.Background()
	conf, err := config.OpenConfigFile("../../fixtures/config.json")
	if err != nil {
		t.Errorf("unable to load configuration")
		return
	}
	connStr, err := conf.Credential().String()
	if err != nil {
		t.Errorf("Unable to process connection string")
		return
	}
	dataCxt, err := data.WithCredential(cxt, connStr)
	if err != nil {
		t.Errorf("open database connection failed")
		return
	}
	defer dataCxt.Close()
	headers := data.NewRowHeader([]string{"user", "password", "status"})
	df, err := data.NewTable("user", headers)
	df.AddRowItems([]string{"obaro", "password_100", "status"})
	if err != nil {
		t.Errorf("Test failed %s", err)
		return
	}
	err = Export(context.Background(), dataCxt, df)
	if err != nil {
		t.Errorf("Failed, export error =  %s", err)
		return
	}

}
