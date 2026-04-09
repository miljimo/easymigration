package reader

import (
	"fmt"
	"os"
)

// Open a new os file object
func openFile(filename string) (*os.File, error) {
	if !Exist(filename) {
		return nil, fmt.Errorf("filename %s does not exist.", filename)
	}
	stream, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read input file = %s , error = %v", filename, err)
	}
	return stream, nil
}
