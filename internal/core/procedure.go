package core

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	godror "github.com/godror/godror"

	"github.com/jmoiron/sqlx"
)

// Store procedure with Godror
func ExecuteStoreProcedure(db *sqlx.DB, context context.Context, spName string, results interface{}, args ...interface{}) error {
	resultsVal := reflect.ValueOf(results)
	var driverRows driver.Rows
	cmdText := buildCmdText(spName, args...)
	execArgs := buildExecutionArguments(&driverRows, args...)
	conn, err := db.Conn(context)
	if err != nil {
		return err
	}

	stmt, err := conn.PrepareContext(context, cmdText)
	if err != nil {
		return err
	}
	if _, err := stmt.ExecContext(context, execArgs...); err != nil {
		return err
	}

	if err != nil {
		return err
	}
	cols := driverRows.(driver.RowsColumnTypeScanType).Columns()
	dests := make([]driver.Value, len(cols))

	if resultsVal.Kind() == reflect.Ptr && resultsVal.Elem().Kind() == reflect.Slice {
		allRows, err := populateRows(driverRows, dests)
		if err != nil {
			return err
		}
		mapToSlice(results, cols, allRows)
	} else {
		populateOne(driverRows, dests)
		mapTo(results, cols, dests)
	}
	defer driverRows.Close()
	defer stmt.Close()
	defer conn.Close()
	return nil
}

func populateRows(cursor driver.Rows, rows []driver.Value) ([][]driver.Value, error) {
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
	v := reflect.ValueOf(obj).Elem()
	t := reflect.TypeOf(obj).Elem()

	if v.Kind() != reflect.Struct {
		fmt.Println("it is not a struct")
		return
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// tags behavior
		arrayTags := field.Tag.Get("oracle")
		parts := strings.Split(arrayTags, ",")
		tagValue := parts[0]
		//convertible := len(parts) > 1 && parts[1] == "convert"
		// end tags behavior

		fieldType := field.Type
		structField := v.Field(i)
		var posInCol int
		for j, elem := range cols {
			if elem == tagValue {
				posInCol = j
				break
			}
		}
		if structField.IsValid() && structField.CanSet() {
			value := dests[posInCol]
			if value != nil {
				valueType := reflect.TypeOf(value)
				destValue := reflect.New(fieldType).Elem()
				fieldStrategyByType(fieldType, valueType, value, destValue)
				structField.Set(destValue)
			} else {
				structField.Set(reflect.Zero(fieldType))
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

func populateOne(cursor driver.Rows, rows []driver.Value) error {
	for {
		if err := cursor.Next(rows); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error cursor.Next: %+v", err)
			return err
		}
	}
	return nil
}

func fieldStrategyByType(fieldType reflect.Type, valueType reflect.Type, value driver.Value, destValue reflect.Value) (reflect.Value, error) {
	switch value := value.(type) {
	case string:
		if valueType.Kind() == reflect.String && fieldType.Kind() == reflect.Int {
			desInt, _ := strconv.Atoi(value)
			destValue.SetInt(int64(desInt))
		} else if fieldType.Kind() == reflect.String {
			destValue.SetString(trimTrailingWhitespace(value))
		} else if fieldType.Kind() == reflect.Bool && valueType.Kind() == reflect.String {
			if len(value) == 1 {
				if value == "S" || value == "N" {
					destValue.SetBool(value == "S")
				}
			}
		}
	case int64:
		if fieldType.Kind() == reflect.Int {
			destValue.SetInt(value)
		} else if fieldType.Kind() == reflect.Int64 {
			destValue.SetInt(value)
		} else if fieldType.Kind() == reflect.String {
			destValue.SetString(strconv.FormatInt(value, 10))
		}
	case []uint8:
		if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Uint8 {
			byteArray := make([]byte, len(value))
			copy(byteArray, value)
			destValue.SetBytes(byteArray)
		}
	case float64:
		if fieldType.Kind() == reflect.Float32 {
			destValue.SetFloat(value)
		} else if fieldType.Kind() == reflect.Float64 {
			destValue.SetFloat(value)
		} else if fieldType.Kind() == reflect.String {
			destValue.SetString(strconv.FormatFloat(value, 'f', -1, 64))
		}
	case bool:
		if fieldType.Kind() == reflect.Bool {
			destValue.SetBool(value)
		} else if fieldType.Kind() == reflect.String {
			destValue.SetString(strconv.FormatBool(value))
		}
	case time.Time:
		if fieldType == reflect.TypeOf(time.Time{}) {
			destValue.Set(reflect.ValueOf(value))
		} else if fieldType.Kind() == reflect.String {
			destValue.SetString(value.Format(time.RFC3339))
		}
	case godror.Number:
		val, err := strconv.Atoi(value.String())
		if err != nil {
			fmt.Printf("Error strconv.Atoi: %+v", err)
		} else {
			if fieldType.Kind() == reflect.Int {
				destValue.SetInt(int64(val))
			} else if fieldType.Kind() == reflect.Int64 {
				destValue.SetInt(int64(val))
			} else if fieldType.Kind() == reflect.String {
				destValue.SetString(value.String())
			}
		}
	default:
		fmt.Printf("Unhandled type: %T\n", value)
	}
	return reflect.Value{}, nil
}
