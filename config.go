package oraclesp

import (
	"context"
	"github.com/ignaciocaff/oraclesp/internal/core"
	"github.com/jmoiron/sqlx"
)

// Configure sets up the global database connection and application context.
// It takes a database connection (dbConn) and a context (ctx) as parameters.
// The function checks if both parameters are provided and not nil, and then assigns them to the global variables db and appContext.
// If any of the parameters is missing, it panics with an error message indicating the missing configuration.

func Configure(dbConn *sqlx.DB, ctx context.Context) {
	core.Configure(dbConn, ctx)
}
