package main

import (
	"syscall"

	"github.com/miljimo/easymigration/cmd/migration"
)

func Execute() error {
	err := migration.RootCommand().Execute()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := Execute()
	if err != nil {
		syscall.Exit(-1)
	}
}
