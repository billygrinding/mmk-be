package dbresolver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
)

// DB is a logical database with multiple underlying physical databases
// forming a single ReadWrite with multiple ReadOnly database.
// Reads and writes are automatically directed to the correct physical db.
type DB struct {
	rwdb            *sql.DB
	rodb            *sql.DB
	totalConnection int
}

// Open concurrently opens each underlying physical db.
// dataSourceNames must be a semi-comma separated list of DSNs with the first
// one being used as the RW-database and the rest as RO-database.
func Open(driverName, dataSourceNames string) (db *DB, err error) {
	db = &DB{}
	conns := strings.Split(dataSourceNames, ";")
	db.totalConnection = len(conns)
	if len(conns) > 2 {
		db.totalConnection = 2
	}

	err = doParallely(db.totalConnection, func(i int) (err error) {
		if i == 0 {
			db.rwdb, err = sql.Open(driverName, conns[i])
			return err
		}
		db.rodb, err = sql.Open(driverName, conns[i])
		return err
	})

	return db, err
}

// Close closes all physical databases concurrently, releasing any open resources.
func (db *DB) Close() error {
	return doParallely(db.totalConnection, func(i int) (err error) {
		if i == 0 {
			return db.rwdb.Close()
		}
		return db.rodb.Close()
	})
}

// Driver returns the physical database's underlying driver.
func (db *DB) Driver() driver.Driver {
	return db.ReadWrite().Driver()
}

// Begin starts a transaction on the RW-database. The isolation level is dependent on the driver.
func (db *DB) Begin() (*sql.Tx, error) {
	return db.ReadWrite().Begin()
}

// BeginTx starts a transaction with the provided context on the RW-database.
//
// The provided TxOptions is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support,
// an error will be returned.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.ReadWrite().BeginTx(ctx, opts)
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// Exec uses the RW-database as the underlying physical db.
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.ReadWrite().Exec(query, args...)
}

// ExecContext executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// Exec uses the RW-database as the underlying physical db.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.ReadWrite().ExecContext(ctx, query, args...)
}

// Ping verifies if a connection to each physical database is still alive,
// establishing a connection if necessary.
func (db *DB) Ping() error {
	err := db.rwdb.Ping()
	if err != nil {
		return err
	}

	if db.rodb != nil {
		return db.rodb.Ping()
	}

	return nil
}

// PingContext verifies if a connection to each physical database is still
// alive, establishing a connection if necessary.
func (db *DB) PingContext(ctx context.Context) error {
	var errRODB, errRWDB error

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		errRWDB = db.rwdb.PingContext(ctx)
		wg.Done()
	}()

	if db.rodb != nil {
		wg.Add(1)
		go func() {
			errRODB = db.rodb.PingContext(ctx)
			wg.Done()
		}()
	}

	wg.Wait()

	if errRWDB != nil && errRODB != nil {
		return errRWDB
	}

	return nil
}

// Prepare creates a prepared statement for later queries or executions
// on each physical database, concurrently.
func (db *DB) Prepare(query string) (Stmt, error) {
	stmt := &stmt{
		db: db,
	}
	err := doParallely(db.totalConnection, func(i int) (err error) {
		if i == 0 {
			stmt.rwstmt, err = db.rwdb.Prepare(query)
			return err
		}

		if db.rodb != nil {
			stmt.rostmt, err = db.rodb.Prepare(query)
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return stmt, nil
}

// PrepareContext creates a prepared statement for later queries or executions
// on each physical database, concurrently.
//
// The provided context is used for the preparation of the statement, not for
// the execution of the statement.
func (db *DB) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	stmt := &stmt{
		db: db,
	}
	err := doParallely(db.totalConnection, func(i int) (err error) {
		if i == 0 {
			stmt.rwstmt, err = db.rwdb.PrepareContext(ctx, query)
			return err
		}

		if db.rodb != nil {
			stmt.rostmt, err = db.rodb.PrepareContext(ctx, query)
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return stmt, nil
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// Query uses a RO database as the physical db.
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	ret, err := db.ReadOnly().Query(query, args...)
	if db.isConnectionError(err) {
		return db.ReadWrite().Query(query, args...)
	}
	return ret, err
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// QueryContext uses a RO database as the physical db.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ret, err := db.ReadOnly().QueryContext(ctx, query, args...)
	if db.isConnectionError(err) {
		return db.ReadWrite().QueryContext(ctx, query, args...)
	}
	return ret, err
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always return a non-nil value.
// Errors are deferred until Row's Scan method is called.
// QueryRow uses a RO database as the physical db.
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	row := db.ReadOnly().QueryRow(query, args...)
	if db.isConnectionError(row.Err()) {
		return db.ReadWrite().QueryRow(query, args...)
	}
	return row
}

// QueryRowContext executes a query that is expected to return at most one row.
// QueryRowContext always return a non-nil value.
// Errors are deferred until Row's Scan method is called.
// QueryRowContext uses a RO database as the physical db.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	row := db.ReadOnly().QueryRowContext(ctx, query, args...)
	if db.isConnectionError(row.Err()) {
		return db.ReadWrite().QueryRowContext(ctx, query, args...)
	}
	return row
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool for each underlying physical db.
// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns then the
// new MaxIdleConns will be reduced to match the MaxOpenConns limit
// If n <= 0, no idle connections are retained.
func (db *DB) SetMaxIdleConns(n int) {
	db.rwdb.SetMaxIdleConns(n)
	if db.rodb != nil {
		db.rodb.SetMaxIdleConns(n)
	}
}

// SetMaxOpenConns sets the maximum number of open connections
// to each physical database.
// If MaxIdleConns is greater than 0 and the new MaxOpenConns
// is less than MaxIdleConns, then MaxIdleConns will be reduced to match
// the new MaxOpenConns limit. If n <= 0, then there is no limit on the number
// of open connections. The default is 0 (unlimited).
func (db *DB) SetMaxOpenConns(n int) {
	db.rwdb.SetMaxOpenConns(n)
	if db.rodb != nil {
		db.rodb.SetMaxOpenConns(n)
	}
}

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
// Expired connections may be closed lazily before reuse.
// If d <= 0, connections are reused forever.
func (db *DB) SetConnMaxLifetime(d time.Duration) {
	db.rwdb.SetConnMaxLifetime(d)
	if db.rodb != nil {
		db.rodb.SetConnMaxLifetime(d)
	}
}

// ReadOnly returns the ReadOnly database
func (db *DB) ReadOnly() *sql.DB {
	if db.rodb == nil {
		return db.rwdb
	}
	return db.rodb
}

// ReadWrite returns the main writer physical database
func (db *DB) ReadWrite() *sql.DB {
	return db.rwdb
}

func (db *DB) isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// the db in stop status will return this error, and it's not *pg.Error
	if strings.Contains(err.Error(), "connection reset by peer") ||
		strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "timed out") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "starting up") ||
		strings.Contains(err.Error(), "EOF") {
		return true
	}

	errPG, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	/*
		copy from https://www.postgresql.org/docs/9.3/errcodes-appendix.html
			"08000": "connection_exception",
			"08003": "connection_does_not_exist",
			"08006": "connection_failure",
			"08001": "sqlclient_unable_to_establish_sqlconnection",
			"08004": "sqlserver_rejected_establishment_of_sqlconnection",
			"08007": "transaction_resolution_unknown",
			"08P01": "protocol_violation",
			57P01	admin_shutdown
			57P02	crash_shutdown
			57P03	cannot_connect_now //shutting down, restart up

			53000	insufficient_resources
			53100	disk_full
			53200	out_of_memory
			53300	too_many_connections
			53400	configuration_limit_exceeded
	*/
	if ArrayContainsStr([]string{"08000", "08003", "08006", "08001", "08004", "08007", "08P01",
		"57P01", "57P02", "57P03",
		"53000", "53100", "53200", "53300", "53400"}, string(errPG.Code)) {
		return true
	}
	return false
}

func ArrayContainsStr(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
