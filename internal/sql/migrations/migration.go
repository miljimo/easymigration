package migrations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"example.com/datamigration/internal/sql/data"
)

/*
The database context implementations
*/

/**Implementation of FileReader Context Reader**/
type FileReader interface {
	Exist(filename string) bool
	ReadAll(filename string) (string, error)
	ReadAllBytes(filename string) ([]byte, error)
}

type implFileReader struct {
}

func (file *implFileReader) Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func (reader *implFileReader) ReadAllBytes(filename string) ([]byte, error) {
	if !reader.Exist(filename) {
		return nil, fmt.Errorf("%s does not exists", filename)
	}
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read input file = %s , error = %v", filename, err)
	}

	return bytes, nil
}

func (reader *implFileReader) ReadAll(filename string) (string, error) {
	bytes, err := reader.ReadAllBytes(filename)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func NewFile() (FileReader, error) {
	return &implFileReader{}, nil
}

/*Implementation of the Database JSOn configuration and its Functionality*/
const (
	CURRENT_CONFIGURATION_VERSION = "1.0.0"
)

type DBCredential interface {
	GetConnectionString() (string, error)
}
type CredentialConfig struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`

	// private collections
}

func (c *CredentialConfig) GetConnectionString() (string, error) {
	if c.Host == "" || c.Username == "" {
		return "", errors.New("invalid database configuration")
	}
	const (
		allowMultiStatments = true
	)
	if strings.Trim(c.Name, " ") == "" {
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true&loc=Local&multiStatements=%t", c.Username, c.Password, c.Host, c.Port, allowMultiStatments), nil
	}
	// connect to specific database
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&multiStatements=%t", c.Username, c.Password, c.Host, c.Port, c.Name, allowMultiStatments), nil
}

type ScriptConfig struct {
	Path  string `json:"path"`
	Order int    `json:"Order"`

	// use internal only
	Files []string `json:"_"`
}

func (script *ScriptConfig) WithPatterns() bool {
	specialChars := []string{"*", "?", "[", "]", "{", "}"}
	for _, c := range specialChars {
		if strings.Contains(script.Path, c) {
			return true
		}
	}
	return false
}

func (script *ScriptConfig) IsDirectory() bool {
	info, err := os.Stat(script.Path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (script *ScriptConfig) getDirectoryFiles(filePath string) ([]string, error) {
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

func (script *ScriptConfig) Prepare(rootPath string) error {
	script.Files = make([]string, 0)
	if script.WithPatterns() || script.IsDirectory() {
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

func (script *ScriptConfig) Execute(cxt context.Context, dataContext data.DataContext, processContent func(content string) (string, error)) error {
	file, _ := NewFile()

	for _, filename := range script.Files {
		fmt.Println("Processing file = " + filename)
		content, err := file.ReadAll(filename)
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

type DatabaseConfig struct {
	Name    string         `json:"name"`
	Scripts []ScriptConfig `json:"scripts"`
	Skip    bool           `json:"skip"`

	// Privates fields
	environ        *EnvironConfig
	configRootPath string
}

func (dbConfig *DatabaseConfig) Prepare() {
	sort.Slice(dbConfig.Scripts, func(a, b int) bool {
		return dbConfig.Scripts[a].Order < dbConfig.Scripts[b].Order
	})

}

func (dbConfig *DatabaseConfig) Migrate(cxt context.Context, dbContext data.DataContext) error {
	dbConfig.Prepare()

	for _, script := range dbConfig.Scripts {

		script.Prepare(dbConfig.configRootPath)
		err := script.Execute(cxt, dbContext, func(content string) (string, error) {
			// Additional processing on the content before
			// executing it.
			bytes, err := dbConfig.environ.ReplaceAll(content)
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		})

		if err != nil {
			return fmt.Errorf("Error = %s", err)
		}
	}
	return nil
}

type EnvironConfig struct {
	keyEntries []string
	values     map[string]string
	re         *regexp.Regexp
}

func (environ *EnvironConfig) Keys() []string {
	if environ.keyEntries == nil {
		environ.keyEntries = make([]string, 0)
	}
	return environ.keyEntries
}
func (environ *EnvironConfig) ContainsKey(key string) bool {
	for _, v := range environ.keyEntries {
		if v == key {
			return true
		}
	}
	return false
}
func (environ *EnvironConfig) Get(key string) string {
	if environ.ContainsKey(key) {
		return environ.values[key]
	}
	return os.Getenv(key)
}
func (environ *EnvironConfig) Set(key string, value string) error {
	if environ.ContainsKey(key) {
		environ.values[key] = value
	}
	return os.Setenv(key, value)
}

func (environ *EnvironConfig) Unset(key string, value string) error {
	if environ.ContainsKey(key) {
		delete(environ.values, key)
	}
	return os.Unsetenv(key)
}

func (environ *EnvironConfig) ReplaceAll(content string) ([]byte, error) {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		key := strings.TrimSpace(re.FindStringSubmatch(match)[1])
		val, ok := os.LookupEnv(key)
		if !ok {
			fmt.Println("Missing environment variables = ", key)
		}
		return val
	})
	return []byte(content), nil
}

func (environ *EnvironConfig) Extract(content string) {
	if (strings.Trim(content, " ")) == "" {
		return
	}
	// extract all the environment variables use in this files content
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	keys := func(content string) []string {
		matches := re.FindAllStringSubmatch(content, -1)
		var vars []string
		for _, m := range matches {
			vars = append(vars, m[1])
		}
		return vars
	}(content)

	for _, key := range keys {
		if environ.ContainsKey(key) {
			continue
		}
		val, ok := os.LookupEnv(key)
		if ok {
			environ.values[key] = val
		}
		environ.keyEntries = append(environ.keyEntries, key)
	}
}

func NewConfigEnviron() *EnvironConfig {
	return &EnvironConfig{
		re:         regexp.MustCompile(`\$\{([^}]+)\}`),
		values:     make(map[string]string, 0),
		keyEntries: make([]string, 0)}
}

/*
Implementing the configuration
  The object houses the json objects and the functions
  that will be use to migrate it to the databases.
*/

type Configuration struct {
	Version     string           `json:"version"`
	Credentials CredentialConfig `json:"credentials"`
	Databases   []DatabaseConfig `json:"databases"`

	// private variables
	environ *EnvironConfig
	path    string
}

func (config *Configuration) Validate() (bool, error) {
	if CURRENT_CONFIGURATION_VERSION != config.Version {
		return false, fmt.Errorf("Invalid configuration version read = %s", config.Version)
	}
	return true, nil
}

func (config *Configuration) Migrate(cxt context.Context, dataContext data.DataContext) error {
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

func (config *Configuration) Environ() EnvironConfig {
	if config.environ == nil {
		config.environ = &EnvironConfig{}
	}
	return *config.environ
}

// Implementations of Configuration Loader

type configReaderImpl struct {
}

func (reader *configReaderImpl) jsonDecode(filename string, environ *EnvironConfig) (Configuration, error) {
	var result Configuration
	file, _ := NewFile()
	bytes, err := file.ReadAllBytes(filename)
	if err != nil {
		return result, err
	}
	bytes, err = environ.ReplaceAll(string(bytes))
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(bytes, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return result, nil
}

func (reader *configReaderImpl) decode(filename string) (*Configuration, error) {

	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	fstream, err := NewFile()
	if err != nil {
		return nil, err
	}
	environ := NewConfigEnviron()

	if !fstream.Exist(filename) {
		return nil, fmt.Errorf("File %s does not exists", filename)
	}
	config, err := reader.jsonDecode(filename, environ)
	config.environ = environ
	config.path = filepath.Dir(filename)

	if err != nil {
		return nil, err
	}
	_, err = config.Validate()
	if err != nil {
		return nil, err
	}
	// Get the port address

	return &config, err
}

// public function for this library, everything should be privates
// not accessible.

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func Decode(path string) (*Configuration, error) {
	decoder := configReaderImpl{}
	return decoder.decode(path)
}

func Migrate(cxt context.Context, config Configuration) error {
	dataCxt, err := data.WithCredential(cxt, &config.Credentials)
	if err != nil {
		return err
	}

	defer dataCxt.Close()

	return config.Migrate(cxt, dataCxt)
}
