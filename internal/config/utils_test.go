package config

import "testing"

func TestGetFileWithBasePattern_Work(t *testing.T) {
	files, err := GetFilesWithBasePattern("../../fixtures/csv/*user_*.csv")
	if err != nil {
		t.Errorf("errors : %s", err)
	}
	if len(files) == 0 {
		t.Errorf("Expecting a files *.go but non found")
	}
}
