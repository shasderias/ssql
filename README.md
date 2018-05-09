# SSQL

SSQL is a thin layer over [sqlx](https://github.com/jmoiron/sqlx) and [goyesql](https://github.com/nleof/goyesql).

When establishing a database connection, in addition to the arguments accepted by sql.Open, SSQL accepts a pattern. The pattern is passed to [zglob](https://github.com/mattn/go-zglob) to obtain a list of files, which is then parsed using [goyesql](https://github.com/nleof/goyesql). SSQL then allows you refer to the parsed SQL statements by name - in the form "[base filename].[statement name]" when calling sql/sqlx functions.

# Installation

```
$ go get -u gitlab.com/shasderias/ssql
```

# Usage

Create one or more SQL files.

```sql
-- /pkg/foo/sql/foo.sql

-- name: All
SELECT *
FROM foo;

-- name: Get
SELECT *
FROM foo
WHERE id = $1;
```

```sql
-- /pkg/bar/sql/bar.sql

-- name: All
SELECT *
FROM bar;

-- name: Get
SELECT *
FROM bar
WHERE id = $1;
```

Call ssql.Open.

```go
// ssql.Open's first 2 arguments are the same as sql.Open (driver, database URL)
// ssql.Open's last argument is a pattern that is passed to zglob for matching
db, _ := ssql.Open("postgres", "postgres://mydb:mydb@localhost/mydb", "./pkg/**/.sql")

foo := Foo{}
// Get has the same semantics as sqlx's Get, except that a name should be passed
// as the 2nd argument
_ = db.Get(&foo, "foo.Get", 1)

bars := []Bar{}
// Select has the same semantics as sqlx's Select, except that a name should be
// passed as the 2nd argument
_ = db.Select(&bars, "bar.All")

thing := Thing{}
// The underlying sqlx.DB can be accessed via DB
_ = db.DB().Select(&thing, "SELECT * FROM things WHERE id = $1", 1)
```

# Implemented Methods

The following methods from sql/sqlx have been implemented.

```go
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
```

# Querier Interface

For code that may or may not be executed in a transaction, the ssql.Querier interface defines the subset of functions common to DB and Tx.

```go
type Querier interface {
	Query(name string, args ...interface{}) (*sql.Rows, error)
	QueryRow(name string, args ...interface{}) *sql.Row
	Get(dest interface{}, name string, args ...interface{}) error
	Select(dest interface{}, name string, args ...interface{}) error
	Exec(name string, args ...interface{}) (sql.Result, error)
	NamedExec(name string, arg interface{}) (sql.Result, error)
}
```