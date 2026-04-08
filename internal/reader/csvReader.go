package reader

// import (
// 	"encoding/csv"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"strings"

// 	"github.com/miljimo/easymigration/internal/data"
// )

// type csvFileReader struct {
// }

// func (reader *csvFileReader) exist(filename string) bool {
// 	_, err := os.Stat(filename)
// 	return err == nil || !os.IsNotExist(err)
// }

// func (reader *csvFileReader) open(filename string) (*os.File, error) {
// 	if !reader.exist(filename) {
// 		return nil, fmt.Errorf("%s does not exists", filename)
// 	}
// 	stream, err := os.Open(filename)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to read input file = %s , error = %v", filename, err)
// 	}
// 	return stream, nil
// }

// func (fr *csvFileReader) read(filename string) (data.Table, error) {
// 	fs, err := fr.open(filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer fs.Close()
// 	reader := csv.NewReader(fs)
// 	record, err := reader.Read()
// 	if err != nil {
// 		return nil, err
// 	}

// 	baseName := filepath.Base(filename)
// 	filenameWithExtension := strings.TrimSuffix(baseName, filepath.Ext(baseName))
// 	frame, err := data.NewFrame(filenameWithExtension, data.NewRowHeader(record))
// 	if err != nil {
// 		return nil, err
// 	}
// 	// read all the rows until there is eof file or error
// 	for {
// 		record, err = reader.Read()
// 		if err != nil {
// 			break
// 		}
// 		frame.AddRowItems(record)
// 	}

// 	// try and detected the frame data types
// 	err = infers.NewDataInference(100).Detect(frame)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return frame, nil
// }

// // Interface implementations
// func (fr *csvFileReader) Read(filename string) (data.Table, error) {
// 	return fr.read(filename)
// }

// func NewCSVFileReader() CSVFileReader {
// 	return &csvFileReader{}
// }
