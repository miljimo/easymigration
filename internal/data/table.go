package data

import "fmt"

type Table interface {
	Name() string
	SetName(name string)
	Rename(columnName string, to string) error
	Headers() Row
	Rows() []Row
	AddRowItems(items []string) error
	AddRow(row Row) error
	Filters(filters []string) (Table, error)
	Contains(columnName string) bool
	ContainsAll(names []string) bool
	RowCounts() int
}

type tableImpl struct {
	name      string
	headers   Row
	data      []Row
	rowCounts int
}

func (df *tableImpl) Headers() Row {
	return df.headers
}

func (df *tableImpl) AddRowItems(row []string) error {
	rowItems, err := CreateRow(df.headers, row)
	if err != nil {
		return err
	}
	return df.AddRow(rowItems)
}

func (df *tableImpl) Filters(filters []string) (Table, error) {
	headers, err := df.headers.Filters(filters)
	if err != nil {
		return nil, err
	}
	frame, err := NewTable(df.name, headers)
	if err != nil {
		return nil, err
	}

	for _, series := range df.data {
		s, err := series.Filters(filters)
		if err != nil {
			return nil, err
		}
		frame.AddRow(s)
	}
	return frame, nil
}

func (df *tableImpl) AddRow(s Row) error {
	if s == nil {
		return fmt.Errorf("cannot add a nil series object")
	}
	if s.Size() != df.headers.Size() {
		return fmt.Errorf("item size must be the same with the header column size")
	}
	df.data = append(df.data, s)
	df.rowCounts += 1
	return nil
}

func (df *tableImpl) Contains(columnName string) bool {
	return df.headers.Contains(columnName)
}

func (df *tableImpl) ContainsAll(columnNames []string) bool {

	if len(columnNames) == 0 {
		return false
	}

	for _, col := range columnNames {
		if !df.headers.Contains(col) {
			return false
		}
	}
	return true
}

func (df *tableImpl) RowCounts() int {
	return df.rowCounts
}
func (df *tableImpl) Rows() []Row {
	copies := make([]Row, df.rowCounts)
	copy(copies, df.data)
	return df.data
}

func (df *tableImpl) Name() string {
	return df.name
}

func (df *tableImpl) SetName(name string) {
	df.name = name
}

func (df *tableImpl) rowRename(row Row, name string, to string) {
	for _, column := range row.Columns() {
		if column.Name() == name {
			column.SetName(to)
		}
	}
}

func (df *tableImpl) Rename(columnName string, to string) error {
	if to == "" {
		return fmt.Errorf("rename to valid name and must not be empty")
	}
	df.rowRename(df.headers, columnName, to)
	for _, row := range df.data {
		df.rowRename(row, columnName, to)
	}
	return nil
}

func (df *tableImpl) Split(n int) ([]Table, error) {
	tables := make([]Table, 0)
	if n < df.RowCounts() {
		tables = append(tables, df)
		return tables, nil
	}
	// int chunks := df.RowCounts()/ n;
	// for range(chunks){

	// }

	return tables, nil
}

func NewTable(title string, headers Row) (Table, error) {
	visited := NewRowHeader([]string{})
	for _, column := range headers.Columns() {
		if visited.Exists(column) {
			return nil, fmt.Errorf("DataFrame: duplicated column header : '%s' found", column.Name())
		}
		column.SetName(column.Name())
		visited.AddColumn(column)
	}
	return &tableImpl{headers: visited, data: []Row{}, rowCounts: 0, name: title}, nil
}
