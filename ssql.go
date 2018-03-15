package ssql

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	zglob "github.com/mattn/go-zglob"
	"github.com/nleof/goyesql"
)

type ErrStmtNotFound string

func (err ErrStmtNotFound) Error() string {
	return fmt.Sprintf("statement '%s' not found", string(err))
}

type DB interface {
	Query(name string, args ...interface{}) (*sql.Rows, error)
	QueryRow(name string, args ...interface{}) *sql.Row
	Get(dest interface{}, name string, args ...interface{}) error
	Select(dest interface{}, name string, args ...interface{}) error
	Exec(name string, args ...interface{}) (sql.Result, error)
	NamedExec(name string, arg interface{}) (sql.Result, error)
	Beginx() (Tx, error)
	DB() *sqlx.DB
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
	query, err := db.lookupSqlStmt(name)
	if err != nil {
		return nil, err
	}
	return db.db.Query(query, args...)
}

func (db SqlxDB) QueryRow(name string, args ...interface{}) *sql.Row {
	query, err := db.lookupSqlStmt(name)
	if err != nil {
		return nil
	}
	return db.db.QueryRow(query, args...)
}

func (db SqlxDB) Get(dest interface{}, name string, args ...interface{}) error {
	query, err := db.lookupSqlStmt(name)
	if err != nil {
		return err
	}
	return db.db.Get(dest, query, args...)
}

func (db SqlxDB) Select(dest interface{}, name string, args ...interface{}) error {
	query, err := db.lookupSqlStmt(name)
	if err != nil {
		return err
	}
	return db.db.Select(dest, query, args...)
}

func (db SqlxDB) Exec(name string, args ...interface{}) (sql.Result, error) {
	query, err := db.lookupSqlStmt(name)
	if err != nil {
		return nil, err
	}
	return db.db.Exec(query, args...)
}

func (db SqlxDB) NamedExec(name string, arg interface{}) (sql.Result, error) {
	query, err := db.lookupSqlStmt(name)
	if err != nil {
		return nil, err
	}
	return db.db.NamedExec(query, arg)
}

func (db SqlxDB) Beginx() (Tx, error) {
	tx, err := db.db.Beginx()
	if err != nil {
		return nil, err
	}

	return &SqlxTx{db, tx}, nil
}

func (db SqlxDB) DB() *sqlx.DB {
	return db.db
}

func (db SqlxDB) lookupSqlStmt(name string) (string, error) {
	parts := strings.Split(name, ".")
	stmt := db.stmts[parts[0]][goyesql.Tag(parts[1])]
	if stmt == "" {
		return "", ErrStmtNotFound(name)
	}
	return stmt, nil
}

type SqlxTx struct {
	db SqlxDB
	tx *sqlx.Tx
}

func (tx SqlxTx) Query(name string, args ...interface{}) (*sql.Rows, error) {
	query, err := tx.db.lookupSqlStmt(name)
	if err != nil {
		return nil, err
	}
	return tx.tx.Query(query, args...)
}

func (tx SqlxTx) QueryRow(name string, args ...interface{}) *sql.Row {
	query, err := tx.db.lookupSqlStmt(name)
	if err != nil {
		return nil
	}
	return tx.tx.QueryRow(query, args...)
}

func (tx SqlxTx) Get(dest interface{}, name string, args ...interface{}) error {
	query, err := tx.db.lookupSqlStmt(name)
	if err != nil {
		return err
	}
	return tx.tx.Get(dest, query, args...)
}

func (tx SqlxTx) Select(dest interface{}, name string, args ...interface{}) error {
	query, err := tx.db.lookupSqlStmt(name)
	if err != nil {
		return err
	}
	return tx.tx.Select(dest, query, args...)
}

func (tx SqlxTx) Exec(name string, args ...interface{}) (sql.Result, error) {
	query, err := tx.db.lookupSqlStmt(name)
	if err != nil {
		return nil, err
	}
	return tx.tx.Exec(query, args...)
}

func (tx SqlxTx) NamedExec(name string, arg interface{}) (sql.Result, error) {
	query, err := tx.db.lookupSqlStmt(name)
	if err != nil {
		return nil, err
	}
	return tx.tx.NamedExec(query, arg)
}

func (tx SqlxTx) Commit() error {
	return tx.tx.Commit()
}

func (tx SqlxTx) Rollback() error {
	return tx.tx.Rollback()
}
