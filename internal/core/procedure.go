package core

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

func ExecuteStoreProcedure(db *sqlx.DB, context context.Context, spName string, results interface{}, args ...interface{}) error {
	conn, err := db.Conn(context)
	resultsVal := reflect.ValueOf(results)

	if err != nil {
		return err
	}

	var cursor driver.Rows

	cmdText := buildCmdText(spName, args...)

	execArgs := buildExecutionArguments(&cursor, args...)

	fmt.Printf("Before execution %s time %s", spName, time.Now().String())
	if _, err := conn.ExecContext(context, cmdText, execArgs...); err != nil {
		return err
	}
	fmt.Printf("After execution %s time %s", spName, time.Now().String())

	cols := cursor.(driver.RowsColumnTypeScanType).Columns()
	rows := make([]driver.Value, len(cols))

	if resultsVal.Kind() == reflect.Ptr && resultsVal.Elem().Kind() == reflect.Slice {
		fmt.Printf("Before populate multiple %s time %s", spName, time.Now().String())
		allRows, err := populateRows(cursor, cols, rows)
		if err != nil {
			return err
		}
		fmt.Printf("After populate multiple %s time %s", spName, time.Now().String())

		fmt.Printf("Before mapToSlice %s time %s", spName, time.Now().String())
		mapToSlice(results, cols, allRows)
		fmt.Printf("After mapToSlice %s time %s", spName, time.Now().String())
	} else {
		fmt.Printf("Before populate one %s time %s", spName, time.Now().String())
		populateOne(cursor, cols, rows)
		fmt.Printf("After populate one %s time %s", spName, time.Now().String())

		fmt.Printf("Before mapTo %s time %s", spName, time.Now().String())
		mapTo(results, cols, rows)
		fmt.Printf("After mapTo %s time %s", spName, time.Now().String())

	}

	cursor.Close()
	return nil
}

func populateRows(cursor driver.Rows, cols []string, rows []driver.Value) ([][]driver.Value, error) {
	var allRows [][]driver.Value
	for {
		if err := cursor.Next(rows); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		newRow := make([]driver.Value, len(rows))
		copy(newRow, rows)
		allRows = append(allRows, newRow)
	}
	return allRows, nil
}

func mapToSlice(slicePtr interface{}, cols []string, allRows [][]driver.Value) error {
	slicePtrValue := reflect.ValueOf(slicePtr)
	sliceType := slicePtrValue.Elem().Type()
	elemType := sliceType.Elem()

	for _, val := range allRows {
		if val != nil {
			newElem := reflect.New(elemType).Elem()
			mapTo(newElem.Addr().Interface(), cols, val)
			slicePtrValue.Elem().Set(reflect.Append(slicePtrValue.Elem(), newElem))
		}
	}
	return nil
}

func mapTo(obj interface{}, cols []string, dests []driver.Value) {
	type CustomMap struct {
		string
		bool
	}
	v := reflect.ValueOf(obj).Elem()
	t := reflect.TypeOf(obj).Elem()
	tags := make(map[string]CustomMap)

	if v.Kind() != reflect.Struct {
		fmt.Println("it is not a struct")
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		arrayTags := field.Tag.Get("oracle")
		parts := strings.Split(arrayTags, ",")
		tagValue := parts[0]
		convertible := len(parts) > 1 && parts[1] == "convert"
		if tagValue != "" {
			tags[tagValue] = CustomMap{fieldName, convertible}
		}
	}
	for i, col := range cols {
		fieldName := tags[col].string
		field := v.FieldByName(fieldName)
		if field.IsValid() && field.CanSet() {
			fieldType := field.Type()
			val := dests[i]
			if val != nil {
				if tags[col].bool && fieldType.Kind() == reflect.Bool {
					val = val == "S"
				}
				if fieldType.Kind() == reflect.String {
					val = trimTrailingWhitespace(val.(string))
				}
				destType := reflect.TypeOf(val)
				if destType.ConvertibleTo(fieldType) {
					field.Set(reflect.ValueOf(val).Convert(fieldType))
				} else {
					fmt.Printf("can not convert %v to %v\n", destType, fieldType)
				}
			} else {
				field.Set(reflect.Zero(fieldType))
			}
		}
	}
}

func buildExecutionArguments(cursor *driver.Rows, args ...interface{}) []interface{} {
	execArgs := make([]interface{}, len(args)+1)
	execArgs[0] = sql.Out{Dest: cursor}
	copy(execArgs[1:], args)
	return execArgs
}

func buildCmdText(spName string, args ...interface{}) string {
	cmdText := fmt.Sprintf("BEGIN %s(:1", spName)
	for i := 0; i < len(args); i++ {
		cmdText += fmt.Sprintf(", :%d", i+2)
	}
	cmdText += "); END;"
	return cmdText
}

func trimTrailingWhitespace(input string) string {
	if len(input) == 0 {
		return input
	}
	input = strings.TrimRight(input, " ")
	return input
}

func populateOne(cursor driver.Rows, cols []string, rows []driver.Value) error {
	for {
		if err := cursor.Next(rows); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}
