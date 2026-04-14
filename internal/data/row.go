package data

import (
	"fmt"
	"strings"
)

type Row interface {
	Columns() []Column
	Add(column string, value string)
	Remove(columnName string)
	AddColumn(column Column)
	GetColumn(name string) Column
	Exists(column Column) bool
	Size() int
	Filters(filters []string) (Row, error)
	Contains(name string) bool
	ToList() []interface{}
	ToStringList() []string
	GetType(string) DBType
}

type rowImpl struct {
	columns []Column
	size    int
	header  bool
}

func CreateRow(header Row, row []string) (Row, error) {
	if header.Size() != len(row) {
		return nil, fmt.Errorf("headers and rows size must be the same")
	}
	s := rowImpl{columns: make([]Column, 0), size: 0}
	for index, item := range header.Columns() {
		s.Add(item.Name(), row[index])
	}
	return &s, nil
}

func NewRowHeader(row []string) Row {
	s := rowImpl{columns: make([]Column, 0), size: 0}
	for _, item := range row {
		s.Add(item, item)
	}
	return &s
}

func (s *rowImpl) contains(name string, values []string) bool {
	for _, value := range values {
		if value == name {
			return true
		}
	}
	return false
}

func (s *rowImpl) Columns() []Column {
	return s.columns
}

func (s *rowImpl) Add(columnName string, value string) {
	s.AddColumn(NewColumn(columnName, value,
		"string",
		len(s.columns)))

}

func (s *rowImpl) Size() int {
	return s.size
}

func (s *rowImpl) Remove(columnName string) {
	for i, column := range s.columns {
		if column.Name() == columnName {
			s.columns = append(s.columns[:i], s.columns[i+1:]...)
			s.size -= 1
			return
		}
	}
}
func (s *rowImpl) GetColumn(name string) Column {
	for _, col := range s.columns {
		if strings.ToLower(col.Name()) == strings.ToLower(name) {
			return col
		}
	}
	return nil
}

func (s *rowImpl) Exists(c Column) bool {
	for _, col := range s.columns {
		if col.Name() == c.Name() {
			return true
		}
	}
	return false
}

func (s *rowImpl) Contains(name string) bool {
	for _, col := range s.columns {
		if col.Name() == name {
			return true
		}
	}
	return false
}

func (s *rowImpl) AddColumn(c Column) {
	s.columns = append(s.columns, c)
	s.size += 1
}

func (s *rowImpl) Filters(filters []string) (Row, error) {
	newS := NewRowHeader([]string{})
	for _, filter := range filters {
		column := s.GetColumn(filter)
		if column == nil {
			continue
		}
		newS.AddColumn(column)
	}
	return newS, nil
}

func (r *rowImpl) ToList() []interface{} {
	values := []interface{}{}
	for _, col := range r.columns {
		values = append(values, col.Value())
	}
	return values
}

func (r *rowImpl) ToStringList() []string {
	values := []string{}
	for _, col := range r.columns {
		values = append(values, col.Text())
	}
	return values
}
func (r *rowImpl) GetType(columnName string) DBType {
	for _, col := range r.columns {
		if col.Name() == columnName {
			return col.Type()
		}
	}
	return DBNull
}
