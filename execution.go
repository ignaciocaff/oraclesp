package oraclesp

import (
	"context"

	"github.com/ignaciocaff/oraclesp/internal/core"
	"github.com/jmoiron/sqlx"
)

// ExecuteStore executes a stored procedure and maps the results to the provided results parameter.
// It takes a database connection, context, stored procedure name, results interface{}, and optional arguments.
func Execute(db *sqlx.DB, procedureName string, context context.Context, result interface{}, args ...interface{}) error {
	core.ExecuteStoreProcedure(db, context, procedureName, result, args...)
	return nil
}
