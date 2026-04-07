package reader

import (
	"fmt"
	"os"
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

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}
