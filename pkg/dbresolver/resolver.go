package dbresolver

import "database/sql"

// WrapDatabaseConnection will wrap to DB connection between RW and RO database
func WrapDatabaseConnection(rwDB, roDB *sql.DB) *DB {
	if rwDB == nil {
		panic("RW Database is required")
	}
	totalConnection := 1
	if roDB != nil {
		totalConnection = 2
	}

	return &DB{
		rwdb:            rwDB,
		rodb:            roDB,
		totalConnection: totalConnection,
	}
}
