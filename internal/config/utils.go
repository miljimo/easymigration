package config

import (
	"os"
	"path/filepath"
)

// The function will get the base and match any patterns that matches the base filename
func GetFilesWithBasePattern(filePath string) ([]string, error) {
	var files []string

	parentFolder := filepath.Dir(filePath)
	pattern := filepath.Base(filePath)

	err := filepath.Walk(parentFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// only work with files.
		if !info.IsDir() {
			name := info.Name()
			match, err := filepath.Match(pattern, name)
			if err != nil {
				return err
			}
			if match {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}
