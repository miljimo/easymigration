package environment

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Environment interface {
	Keys() []string
	ContainsKey(key string) bool
	Get(key string) string
	Set(key string, value string) error
	Unset(key string, value string) error
	ReplaceAll(content string) ([]byte, error)
	Extract(content string)
}

type impEnvironment struct {
	keyEntries []string
	values     map[string]string
	re         *regexp.Regexp
}

func (environ *impEnvironment) Keys() []string {
	if environ.keyEntries == nil {
		environ.keyEntries = make([]string, 0)
	}
	return environ.keyEntries
}
func (environ *impEnvironment) ContainsKey(key string) bool {
	for _, v := range environ.keyEntries {
		if v == key {
			return true
		}
	}
	return false
}
func (environ *impEnvironment) Get(key string) string {
	if environ.ContainsKey(key) {
		return environ.values[key]
	}
	return os.Getenv(key)
}
func (environ *impEnvironment) Set(key string, value string) error {
	if environ.ContainsKey(key) {
		environ.values[key] = value
	}
	return os.Setenv(key, value)
}

func (environ *impEnvironment) Unset(key string, value string) error {
	if environ.ContainsKey(key) {
		delete(environ.values, key)
	}
	return os.Unsetenv(key)
}

func (environ *impEnvironment) ReplaceAll(content string) ([]byte, error) {
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

func (environ *impEnvironment) Extract(content string) {
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

func New() Environment {
	return &impEnvironment{
		re:         regexp.MustCompile(`\$\{([^}]+)\}`),
		values:     make(map[string]string, 0),
		keyEntries: make([]string, 0)}
}
