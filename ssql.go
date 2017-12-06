package ssql

import (
	"database/sql"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	zglob "github.com/mattn/go-zglob"
	"github.com/nleof/goyesql"
)

type DB interface {
	Query(name string, args ...interface{}) (*sql.Rows, error)
	QueryRow(name string, args ...interface{}) *sql.Row
	Get(dest interface{}, name string, args ...interface{}) error
	Select(dest interface{}, name string, args ...interface{}) error
	Exec(name string, args ...interface{}) (sql.Result, error)
	NamedExec(name string, arg interface{}) (sql.Result, error)
	Beginx() (Tx, error)
}

type Tx interface {
	Query(name string, args ...interface{}) (*sql.Rows, error)
	QueryRow(name string, args ...interface{}) *sql.Row
	Get(dest interface{}, name string, args ...interface{}) error
	Select(dest interface{}, name string, args ...interface{}) error
	Exec(name string, args ...interface{}) (sql.Result, error)
	NamedExec(name string, arg interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
}

type Querier interface {
	Query(name string, args ...interface{}) (*sql.Rows, error)
	QueryRow(name string, args ...interface{}) *sql.Row
	Get(dest interface{}, name string, args ...interface{}) error
	Select(dest interface{}, name string, args ...interface{}) error
	Exec(name string, args ...interface{}) (sql.Result, error)
	NamedExec(name string, arg interface{}) (sql.Result, error)
}

func Open(driverName, dataSourceName, sqlPath string) (DB, error) {
	matches, err := zglob.Glob(sqlPath)
	if err != nil {
		return nil, err
	}

	stmts := map[string]goyesql.Queries{}

	for _, path := range matches {
		filename := filepath.Base(path)
		filenameWithoutExt := strings.Split(filename, ".")[0]

		stmts[filenameWithoutExt] = goyesql.MustParseFile(path)
	}

	db, err := sqlx.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return SqlxDB{db: db, stmts: stmts}, nil
}

type SqlxDB struct {
	db    *sqlx.DB
	stmts map[string]goyesql.Queries
}

func (db SqlxDB) Query(name string, args ...interface{}) (*sql.Rows, error) {
	query := db.lookupSqlStmt(name)
	return db.db.Query(query, args...)
}

func (db SqlxDB) QueryRow(name string, args ...interface{}) *sql.Row {
	query := db.lookupSqlStmt(name)
	return db.db.QueryRow(query, args...)
}

func (db SqlxDB) Get(dest interface{}, name string, args ...interface{}) error {
	query := db.lookupSqlStmt(name)
	return db.db.Get(dest, query, args...)
}

func (db SqlxDB) Select(dest interface{}, name string, args ...interface{}) error {
	query := db.lookupSqlStmt(name)
	return db.db.Select(dest, query, args...)
}

func (db SqlxDB) Exec(name string, args ...interface{}) (sql.Result, error) {
	query := db.lookupSqlStmt(name)
	return db.db.Exec(query, args...)
}

func (db SqlxDB) NamedExec(name string, arg interface{}) (sql.Result, error) {
	query := db.lookupSqlStmt(name)
	return db.db.NamedExec(query, arg)
}

func (db SqlxDB) Beginx() (Tx, error) {
	tx, err := db.db.Beginx()
	if err != nil {
		return nil, err
	}

	return &SqlxTx{db, tx}, nil
}

func (db SqlxDB) lookupSqlStmt(name string) string {
	parts := strings.Split(name, ".")
	return db.stmts[parts[0]][goyesql.Tag(parts[1])]
}

type SqlxTx struct {
	db SqlxDB
	tx *sqlx.Tx
}

func (tx SqlxTx) Query(name string, args ...interface{}) (*sql.Rows, error) {
	query := tx.db.lookupSqlStmt(name)
	return tx.tx.Query(query, args...)
}

func (tx SqlxTx) QueryRow(name string, args ...interface{}) *sql.Row {
	query := tx.db.lookupSqlStmt(name)
	return tx.tx.QueryRow(query, args...)
}

func (tx SqlxTx) Get(dest interface{}, name string, args ...interface{}) error {
	query := tx.db.lookupSqlStmt(name)
	return tx.tx.Get(dest, query, args...)
}

func (tx SqlxTx) Select(dest interface{}, name string, args ...interface{}) error {
	query := tx.db.lookupSqlStmt(name)
	return tx.tx.Select(dest, query, args...)
}

func (tx SqlxTx) Exec(name string, args ...interface{}) (sql.Result, error) {
	query := tx.db.lookupSqlStmt(name)
	return tx.tx.Exec(query, args...)
}

func (tx SqlxTx) NamedExec(name string, arg interface{}) (sql.Result, error) {
	query := tx.db.lookupSqlStmt(name)
	return tx.tx.NamedExec(query, arg)
}
func (tx SqlxTx) Commit() error {
	return tx.tx.Commit()
}

func (tx SqlxTx) Rollback() error {
	return tx.tx.Rollback()
}
