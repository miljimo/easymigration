package reader

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/miljimo/easymigration/internal/data"
)

/*
The database context implementations
*/

type FileReader interface {
	ReadAll() (string, error)
	ReadAllBytes() ([]byte, error)
}

type implFileReader struct {
	filename string
}

func (fs *implFileReader) ReadAllBytes() ([]byte, error) {
	bytes, err := os.ReadFile(fs.filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read input file = %s , error = %v", fs.filename, err)
	}
	return bytes, nil
}

func (fs *implFileReader) ReadAll() (string, error) {
	bytes, err := fs.ReadAllBytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

/*
Open a function and create the instance of a reader
*/
func Open(filename string) (FileReader, error) {
	fs := &implFileReader{filename: filename}
	if !Exist(fs.filename) {
		return nil, fmt.Errorf("%s doesnt not exist", filename)
	}
	return fs, nil
}

func ReadCSV(filename string) (data.Table, error) {
	fs, err := openFile(filename)
	if err != nil {
		return nil, err
	}
	defer fs.Close()
	reader := csv.NewReader(fs)
	record, err := reader.Read()
	if err != nil {
		return nil, err
	}

	baseName := filepath.Base(filename)
	filenameWithExtension := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	frame, err := data.NewFrame(filenameWithExtension, data.NewRowHeader(record))
	if err != nil {
		return nil, err
	}

	//  Read all the records until there is no record to
	// read from the reader.
	for {
		record, err = reader.Read()
		if err != nil {
			break
		}
		frame.AddRowItems(record)
	}
	return frame, nil
}

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}
