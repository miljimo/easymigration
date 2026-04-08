package data

import "testing"

func Test_RenameDataFrameColumnName(t *testing.T) {
	headers := NewRowHeader([]string{"Email", "Message", "Time", "Date"})
	df, _ := NewFrame("test", headers)
	df.Rename("Email", "email")

	if df.Contains("Email") {
		t.Errorf("Failed: expect Email column to be missing now since its changed")
	}
	if !df.Contains("email") {
		t.Errorf("Failed : expect 'email' column to there")
	}

}
