package core

import (
	"github.com/jmoiron/sqlx"
)

var (
	db *sqlx.DB
)

func Configure(dbConn *sqlx.DB) {
	if dbConn == nil {
		panic("Configuración de base de datos y/o contexto no proporcionados")
	}
	db = dbConn
}

func GetDB() *sqlx.DB {
	return db
}
