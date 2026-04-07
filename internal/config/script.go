package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/miljimo/easymigration/internal/reader"
	"github.com/miljimo/easymigration/internal/sql/data"
)

type Script interface {
	Prepare(rootPath string) error
	Execute(cxt context.Context, dataContext data.DataContext, processContent func(content string) (string, error)) error
}

type ScriptData struct {
	Path  string `json:"path"`
	Order int    `json:"Order"`
	// use internal only
	Files []string `json:"_"`
}

func (script *ScriptData) withPatterns() bool {
	specialChars := []string{"*", "?", "[", "]", "{", "}"}
	for _, c := range specialChars {
		if strings.Contains(script.Path, c) {
			return true
		}
	}
	return false
}

func (script *ScriptData) isDirectory() bool {
	info, err := os.Stat(script.Path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (script *ScriptData) getDirectoryFiles(filePath string) ([]string, error) {
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

func (script *ScriptData) Prepare(rootPath string) error {
	script.Files = make([]string, 0)
	if script.withPatterns() || script.isDirectory() {
		absPath := filepath.Join(rootPath, script.Path)
		// load the directory with the files .sql
		files, err := script.getDirectoryFiles(absPath)
		if err != nil {
			return err
		}
		script.Files = append(script.Files, files...)
		return nil
	}
	script.Files = append(script.Files, filepath.Join(rootPath, script.Path))
	return nil
}

func (script *ScriptData) Execute(cxt context.Context, dataContext data.DataContext, processContent func(content string) (string, error)) error {

	for _, filename := range script.Files {
		fmt.Println("Processing file = " + filename)
		fs, err := reader.Open(filename)
		if err != nil {
			return err
		}
		content, err := fs.ReadAll()
		if err != nil {
			return err
		}
		content = strings.ReplaceAll(content, "DELIMITER $$", "")
		content = strings.ReplaceAll(content, "DELIMITER ;", "")
		content = strings.ReplaceAll(content, "$$", "")

		if processContent != nil {
			content, err = processContent(content)
			if err != nil {
				return err
			}
		}
		// Add some level of security to prevent SQL injections.
		_, err = dataContext.Execute(cxt, content)
		if err != nil {
			return err
		}

	}
	return nil
}
