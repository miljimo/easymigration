package data

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
)

// Database types declarations
type DBType string

const (
	DBNull     DBType = "None"
	DBString   DBType = "string"
	DBInteger  DBType = "int"
	DBBool     DBType = "bool"
	DBFloat    DBType = "float"
	DBDatetime DBType = "datetime"
)

var DateFormats = []string{
	time.DateOnly, // "2006-01-02"
	time.RFC3339,  // "2006-01-02T15:04:05Z07:00"
	time.DateTime, // "2006-01-02 15:04:05"
	time.UnixDate, // "Mon Jan _2 15:04:05 MST 2006"
	"01/02/2006",  // MM/DD/YYYY
	"02/01/2006",  // DD/MM/YYYY
}

type Column interface {
	Name() string
	SetName(name string)
	Type() DBType
	SetType(DBType)
	Value() interface{}
	Text() string
	Index() int
}

type columnImpl struct {
	name  string
	text  string
	typed DBType
	index int
	value interface{}
}

func NewColumn(name string, text string, dType DBType, index int) Column {
	return &columnImpl{name: name,
		text:  text,
		typed: dType,
		index: index,
		value: nil}
}

func (c *columnImpl) Name() string {
	return c.name
}

func (c *columnImpl) SetName(name string) {
	c.name = name
}

func (c *columnImpl) Type() DBType {
	return c.typed
}

func (c *columnImpl) SetType(dbType DBType) {
	c.typed = dbType
}

func (c *columnImpl) Index() int {
	return c.index
}

func (c *columnImpl) Value() interface{} {
	val, err := c.convertTextToGoType(c.text, c.typed)
	c.value = val
	if err != nil {
		log.Println(err)
		return nil
	}
	return c.value
}

func (c *columnImpl) Text() string {
	return c.text
}
func (c *columnImpl) convertTextToGoType(value string, dtype DBType) (interface{}, error) {
	val := strings.TrimSpace(value)
	if val == "" {
		return nil, nil
	}
	switch dtype {
	case DBNull:
		return nil, nil
	case DBInteger:
		return strconv.Atoi(val)
	case DBFloat:
		return strconv.ParseFloat(val, 64)
	case DBBool:
		return strconv.ParseBool(val)
	case DBDatetime:
		for _, layout := range DateFormats {
			t, err := time.Parse(layout, val)
			if err == nil {
				return t.Format("2006-01-02"), nil
			}
		}
		return nil, errors.New("unable to parse datetime")
	case DBString:
		return val, nil
	default:
		return val, nil
	}
}
