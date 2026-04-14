package config

import (
	"testing"
)

func Test_GetFilesForIngestionLoadedSuccessfully(t *testing.T) {
	ingest := NewIngestion("../../fixtures/csv/*user_*.csv")

	files, err := ingest.Files()
	if err != nil {
		t.Errorf("Error = %s", err)
	}
	if len(files) == 0 {
		t.Error("no files found")
	}

}
