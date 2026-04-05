package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type RowsAffected int64
type DBResults []map[string]interface{}

const (
	DEFAULT_CONNECTION_TIMEOUT = 1 * time.Second
	DB_PROCEDURE_TIMEOUT       = 15 * time.Second
	MYSQL_DRIVER               = "mysql"
	MAX_DB_IDLE_CONNECTIONS    = 3
	MAX_DB_OPEN_CONNECTION     = 3
)

type DataCredential interface {
	GetConnectionString() (string, error)
}

/*
	    DataContext
		 this provides the generic function that can be use to run queries and manipulate databases
*/

type DataContext interface {
	CallProc(ctx context.Context, procName string, params ...interface{}) (DBResults, error)
	Query(ctx context.Context, query string, params ...interface{}) (DBResults, error)
	Execute(ctx context.Context, query string, params ...interface{}) (RowsAffected, error)
	Close() error
}

type dataContextImpl struct {
	db *sql.DB
}

func (db *dataContextImpl) Close() error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

func (db *dataContextImpl) fetch(rows *sql.Rows) (DBResults, error) {
	resultSets := DBResults{}
	columns, err := rows.Columns()
	if err != nil {
		return resultSets, err
	}

	columnCounts := len(columns)

	for rows.Next() {
		columnValues := make([]interface{}, columnCounts)
		columnPointers := make([]interface{}, columnCounts)
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return resultSets, err
		}
		// Create a map and fill it with column names and values
		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := columnValues[i]
			// Handle []byte -> string conversion (common with TEXT/VARCHAR)
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			rowMap[strings.ToLower(colName)] = val
		}
		resultSets = append(resultSets, rowMap)
	}
	return resultSets, nil
}

func (db *dataContextImpl) createParameterPlaceholder(params ...interface{}) string {
	if len(params) == 0 {
		return "()"
	}
	placeholders := make([]string, len(params))
	for i := range params {
		placeholders[i] = "?"
	}
	return fmt.Sprintf("(%s)", strings.Join(placeholders, ","))
}

func (db *dataContextImpl) CallProc(ctx context.Context, procName string, params ...interface{}) (DBResults, error) {
	if procName == "" {
		return nil, errors.New("procedure name is required")
	}
	placeholders := db.createParameterPlaceholder(params...)
	query := fmt.Sprintf("CALL %s%s", procName, placeholders)
	return db.Query(ctx, query, params...)
}

func (db *dataContextImpl) Query(ctx context.Context, query string, params ...interface{}) (DBResults, error) {
	if db == nil {
		return nil, fmt.Errorf("mysql client object must not be nil")
	}
	ctx, cancel := context.WithTimeout(ctx, DB_PROCEDURE_TIMEOUT)
	defer cancel()

	rows, err := db.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("DB: error =%w", err)
	}
	defer rows.Close()
	return db.fetch(rows)
}

func (db *dataContextImpl) Execute(ctx context.Context, query string, params ...interface{}) (RowsAffected, error) {
	if db == nil {
		return 0, fmt.Errorf("mysql client object must not be nil")
	}
	ctx, cancel := context.WithTimeout(ctx, DB_PROCEDURE_TIMEOUT)
	defer cancel()

	rowsAffected, err := db.db.ExecContext(ctx, query, params...)
	if err != nil {
		return 0, fmt.Errorf("DB: error =%w", err)
	}
	rows, err := rowsAffected.RowsAffected()
	if err != nil {
		return 0, err
	}
	return RowsAffected(rows), nil
}

/*
   Provides a means to connect to the databases.
*/

type dataConnectionImpl struct {
}

func (connector *dataConnectionImpl) initial(cxt context.Context, db *sql.DB) error {
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(MAX_DB_IDLE_CONNECTIONS)
	db.SetMaxOpenConns(MAX_DB_OPEN_CONNECTION)
	ctx, cancel := context.WithTimeout(cxt, DEFAULT_CONNECTION_TIMEOUT)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	return nil
}

// The function open a connection to the database and return a DataContext that can be use to run
// SQL queries
func (connector *dataConnectionImpl) createDataContext(cxt context.Context, credential DataCredential, createFunc func(cxt context.Context, db *sql.DB) (DataContext, error)) (DataContext, error) {

	dsn, err := credential.GetConnectionString()
	if err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open(MYSQL_DRIVER, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to open connection: %w", err)
	}

	if err := connector.initial(cxt, sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	// if the creator function is provided then we should use it
	if createFunc != nil {
		dataCxt, err := createFunc(cxt, sqlDB)
		if err != nil {
			return nil, err
		}

		return dataCxt, nil
	}
	// otherwise use the default mysql context
	return &dataContextImpl{db: sqlDB}, nil
}

func WithCredential(cxt context.Context, credential DataCredential) (DataContext, error) {
	conn := dataConnectionImpl{}
	return conn.createDataContext(cxt, credential, nil)
}
