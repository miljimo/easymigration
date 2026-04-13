package config

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/miljimo/easymigration/internal/data"
	"github.com/miljimo/easymigration/internal/environment"
	"github.com/miljimo/easymigration/internal/reader"
)

const (
	CURRENT_CONFIGURATION_VERSION = "1.0.0"
)

type Configuration interface {
	Validate() (bool, error)
	Migrate(cxt context.Context, dataContext data.DataContext) error
	Environ() environment.Environment
	Credential() Stringer
	SetEnvironment(environ environment.Environment)
	Path() string
}

type ConfigurationData struct {
	Version     string         `json:"version"`
	Credentials CredentialData `json:"credentials"`
	Databases   []DatabaseData `json:"databases"`

	// private variables
	environ environment.Environment
	path    string
}

func (config *ConfigurationData) Validate() (bool, error) {
	if CURRENT_CONFIGURATION_VERSION != config.Version {
		return false, fmt.Errorf("Invalid configuration version read = %s", config.Version)
	}
	return true, nil
}

func (config *ConfigurationData) Migrate(cxt context.Context, dataContext data.DataContext) error {
	if cxt.Err() != nil {
		return cxt.Err()
	}
	for _, dbConfig := range config.Databases {
		if dbConfig.Skip {
			fmt.Println("Skipping database  = ", dbConfig.Name)
			continue
		}
		if strings.Trim(dbConfig.Name, " ") != "" {
			dataContext.Execute(cxt, fmt.Sprintf("USE %s;", dbConfig.Name))
		}
		dbConfig.environ = config.environ
		dbConfig.configRootPath = config.path
		err := dbConfig.Migrate(cxt, dataContext)
		if err != nil {
			return err
		}
	}
	return nil
}

func (config *ConfigurationData) Environ() environment.Environment {
	if config.environ == nil {
		config.environ = environment.New()
	}
	return config.environ
}

func (config *ConfigurationData) Credential() Stringer {
	return &config.Credentials
}

func (config *ConfigurationData) SetEnvironment(environ environment.Environment) {
	if config.environ != nil {
		return
	}
	config.environ = environ
}

func (config *ConfigurationData) Path() string {
	return config.path
}

func OpenConfigFile(filename string) (Configuration, error) {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	var config ConfigurationData
	var environ = environment.New()
	file, _ := reader.Open(filename)
	bytes, err := file.ReadAllBytes()

	if err != nil {
		return nil, err
	}
	bytes, err = environ.ReplaceAll(string(bytes))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	config.SetEnvironment(environ)
	config.path = filepath.Dir(filename)
	_, err = config.Validate()
	if err != nil {
		return &config, err
	}
	return &config, nil
}

func Migrate(cxt context.Context, filename string) error {
	config, err := OpenConfigFile(filename)
	if err != nil {
		return err
	}

	connStr, err := config.Credential().String()
	if err != nil {
		return err
	}
	dataCxt, err := data.WithCredential(cxt, connStr)
	if err != nil {
		return err
	}
	defer dataCxt.Close()
	return config.Migrate(cxt, dataCxt)
}
