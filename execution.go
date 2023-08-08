package oraclesp

import (
	"context"

	"github.com/ignaciocaff/oraclesp/internal/core"
)

// ExecuteStore executes a stored procedure and maps the results to the provided results parameter.
// It takes a database connection, context, stored procedure name, results interface{}, and optional arguments.
func Execute(procedureName string, context context.Context, result interface{}, args ...interface{}) error {
	core.ExecuteStoreProcedure(core.GetDB(), context, procedureName, result, args...)
	return nil
}
