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
	first := time.Now()
	fmt.Printf("Starting procedure %s time %s", spName, first)

	conn, err := db.Conn(context)
	resultsVal := reflect.ValueOf(results)

	if err != nil {
		return err
	}

	var cursor driver.Rows

	cmdText := buildCmdText(spName, args...)

	execArgs := buildExecutionArguments(&cursor, args...)

	if _, err := conn.ExecContext(context, cmdText, execArgs...); err != nil {
		return err
	}

	cols := cursor.(driver.RowsColumnTypeScanType).Columns()
	rows := make([]driver.Value, len(cols))

	if resultsVal.Kind() == reflect.Ptr && resultsVal.Elem().Kind() == reflect.Slice {
		allRows, err := populateRows(cursor, cols, rows)
		if err != nil {
			return err
		}
		mapToSlice(results, cols, allRows)
	} else {
		populateOne(cursor, cols, rows)
		mapTo(results, cols, rows)
	}
	cursor.Close()
	calcEnd(first, spName)
	return nil
}

func calcEnd(now time.Time, spName string) {
	second := time.Now()
	fmt.Printf("Ending procedure %s time %s", spName, now)
	duration := second.Sub(now)
	fmt.Printf("Duration %d seconds %d milliseconds", int64(duration.Seconds()), duration.Nanoseconds()/int64(time.Millisecond))
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
