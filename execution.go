package oraclesp

import (
	"github.com/ignaciocaff/oraclesp/internal/core"
)

// ExecuteStore executes a stored procedure and maps the results to the provided results parameter.
// It takes a database connection, context, stored procedure name, results interface{}, and optional arguments.
func Execute(procedureName string, result interface{}, args ...interface{}) error {
	core.ExecuteStoreProcedure(core.GetDB(), procedureName, result, args...)
	return nil
}
