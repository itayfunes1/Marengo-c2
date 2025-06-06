package utils

import (
	"bytes"
	"fmt"

	"github.com/go-sqlite/sqlite3"
)

type TableRow struct {
	columns map[string]int
	record  *sqlite3.Record
}

func VisitTableRows(db *sqlite3.DbFile, tableName string, columnNameMappings map[string]string, f func(rowID *int64, row TableRow) error) error {
	columns := make(map[string]int)
	if table, ok := findTable(db, tableName); ok {
		for index, column := range table.Columns() {
			columnName := column.Name()
			if mappedColumnName, ok := columnNameMappings[columnName]; ok {
				columnName = mappedColumnName
			}
			if _, ok := columns[columnName]; !ok {
				columns[columnName] = index
			}
		}
	} else {
		return fmt.Errorf("Unable to find table named [%s] in %v", tableName, db)
	}
	return db.VisitTableRecords(tableName, func(rowID *int64, record sqlite3.Record) error {
		return f(rowID, TableRow{columns, &record})
	})
}

func findTable(db *sqlite3.DbFile, tableName string) (sqlite3.Table, bool) {
	for _, table := range db.Tables() {
		if table.Name() == tableName {
			return table, true
		}
	}
	return sqlite3.Table{}, false
}

func (row TableRow) BytesOrFallback(columnName string, fallback []byte) ([]byte, error) {
	rawValue := row.ValueOrFallback(columnName, nil)
	if value, ok := rawValue.([]byte); ok {
		return value, nil
	}
	return nil, fmt.Errorf("expected column [%s] to be []byte; got %T with value %[2]v", columnName, rawValue)
}
func (row TableRow) BytesStringOrFallback(columnName string, fallback []byte) ([]byte, error) {
	rawValue := row.ValueOrFallback(columnName, nil)
	if value, ok := rawValue.([]byte); ok {
		return value, nil
	}
	switch value := rawValue.(type) {
	case []byte:
		return value, nil
	case string:
		return []byte(value), nil
	}
	return nil, fmt.Errorf("expected column [%s] to be []byte or string; got %T with value %[2]v", columnName, rawValue)
}

func (row TableRow) Bool(columnName string) (bool, error) {
	rawValue, err := row.Value(columnName)
	if err != nil {
		return false, err
	}
	return rawValue != 0, nil
}

func (row TableRow) Int64(columnName string) (int64, error) {
	rawValue, err := row.Value(columnName)
	if err != nil {
		return 0, err
	}
	switch value := rawValue.(type) {
	case int64:
		return value, nil
	case uint64:
		if int64(value) < 0 {
			return 0, fmt.Errorf("expected column [%s] to be int64; got uint64 value that can't fit in int64: %d", columnName, rawValue)
		}
		return int64(value), nil
	case int32:
		return int64(value), nil
	case int:
		return int64(value), nil
	default:
		return 0, fmt.Errorf("expected column [%s] to be int64; got %T with value %[2]v", columnName, rawValue)
	}
}

func (row TableRow) String(columnName string) (string, error) {
	rawValue, err := row.Value(columnName)
	if err != nil {
		return "", err
	}
	if value, ok := rawValue.(string); ok {
		return value, nil
	}
	return "", fmt.Errorf("expected column [%s] to be string; got %T with value %[2]v", columnName, rawValue)
}

func (row TableRow) Value(columnName string) (interface{}, error) {
	if index, ok := row.columns[columnName]; !ok {
		return nil, fmt.Errorf("table doesn't have a column named [%s]", columnName)
	} else if count := len(row.columns); count <= index {
		return nil, fmt.Errorf("column named [%s] has index %d but row only has %d values", columnName, index, count)
	} else {
		return row.record.Values[index], nil
	}
}

func (row TableRow) ValueOrFallback(columnName string, fallback interface{}) interface{} {
	if index, ok := row.columns[columnName]; ok && index < len(row.columns) {
		return row.record.Values[index]
	}
	return fallback
}

func DecryptCipherText(encrypted []byte, secretedKey []byte) ([]byte, error) {
	var decrypt func(encrypted, password []byte) ([]byte, error)
	switch {
	case bytes.HasPrefix(encrypted, prefixDPAPI[:]):
		// present before Chrome v80 on Windows
		decrypt = func(encrypted, _ []byte) ([]byte, error) {
			return DecryptWinApi(encrypted)
		}
	case bytes.HasPrefix(encrypted, []byte(`v10`)):
		fallthrough
	default:
		decrypt = decryptAES256GCM
	}

	decrypted, err := decrypt(encrypted, secretedKey)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}
