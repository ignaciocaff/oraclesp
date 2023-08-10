package core

import (
	"context"

	"github.com/jmoiron/sqlx"
)

var (
	db         *sqlx.DB
	appContext context.Context
)

func Configure(dbConn *sqlx.DB, ctx context.Context) {
	if dbConn == nil {
		panic("Configuración de base de datos y/o contexto no proporcionados")
	}
	db = dbConn
	appContext = ctx
}

func GetDB() *sqlx.DB {
	return db
}

func GetContext() context.Context {
	return appContext
}
