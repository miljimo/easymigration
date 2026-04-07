/*
The configuration object that houses the credential of the database to connect.
*/
package config

import (
	"errors"
	"fmt"
	"strings"
)

type Credential interface {
	String() (string, error)
}

type CredentialData struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}

// Get the connection string of the credentials
func (c *CredentialData) String() (string, error) {
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
