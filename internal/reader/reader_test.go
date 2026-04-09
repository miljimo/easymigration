package reader

import (
	"testing"
)

func TestReadCSVFileTable(t *testing.T) {

	filename := "./fixtures/csv/test_users.csv"
	table, err := ReadCSV(filename)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	if table.RowCounts() != 3 {
		t.Errorf("Failed : expecting 3 record , but %d recieved", table.RowCounts())
	}

	if !table.Contains("Name") {
		t.Errorf("Does not contain header 'Name'")
	}
}
