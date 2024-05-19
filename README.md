OpenMLDB Go SDK
------

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/4paradigm/openmldb-go-sdk)
[![GitHub Release](https://img.shields.io/github/v/release/4paradigm/openmldb-go-sdk)](https://github.com/4paradigm/openmldb-go-sdk/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/4paradigm/openmldb-go-sdk)](https://goreportcard.com/report/github.com/4paradigm/openmldb-go-sdk)
[![Go Reference](https://pkg.go.dev/badge/github.com/4paradigm/openmldb-go-sdk.svg)](https://pkg.go.dev/github.com/4paradigm/openmldb-go-sdk)
[![Codecov](https://img.shields.io/codecov/c/github/4paradigm/openmldb-go-sdk)](https://codecov.io/gh/4paradigm/openmldb-go-sdk)

Pure Go [OpenMLDB](https://github.com/4paradigm/OpenMLDB) driver for database/sql, connect via HTTP.

## Features

- Lightweight
- Pure Go implementation, No C-bindings
- Connection over HTTP
- Full OpenMLDB SQL support, work with online and offline mode
- Numeric, bool, string, date, timestamp data type support

## Requirements

- OpenMLDB with all components version >= 0.6.2
- OpenMLDB API Server setted up

## Installation

```sh
go get github.com/4paradigm/openmldb-go-sdk
```

## Data Source Name (DSN)

```
openmldb://<API_SERVER_HOST>:<API_SERVER_PORT>/<DB_NAME>?mode=<MODE_NAME>
```

For example, to open a database to `test_db` by api server at `127.0.0.1:8080`:
```go
db, err := sql.Open("openmldb", "openmldb://127.0.0.1:8080/test_db")
```

`<DB_NAME>` is mandatory in DSN, and at this time (version 0.2.0), you must ensure the database `<DB_NAME>` created before open go connection.
DSN parameters (the `?mode=<MODE_NAME>` part) are optional.


### Query Mode (Optional)

The execution mode for OpenMLDB, defined as `mode=<MODE_NAME>`, default to `online`, available values are:
- `online`: online preview mode
- `offsync`: offline mode with system variable `sync_job = true`
- `offasync`: offline mode with system variable `sync_job = false`


## Data type support

int16, int32, int64, float, double, bool, date, timestamp and string types in OpenMLDB SQL are supported.
Since Go types are flexible by design, you may choose any type in Go by your favor, as long as that type implements
[sql#Scanner](https://pkg.go.dev/database/sql#Scanner) interface.

For example, a SQL string type, can be represented in Go with `string`, `sql.NullString`, `sql.Null[string]`, `string` is able
to represent SQL string when it is not NULL, error reported if you want to save a NULL value into string, while the later two types
are able to save all SQL string, regardless nullable:

```go
import (
  "database/sql"
)

// ...

{
  var s string
  err := db.QueryRow("SELECT name FROM foo WHERE id=?", id).Scan(&s)
  // err returned from Scan if NULL value returned from query
}


{
  var s sql.NullString
  err := db.QueryRow("SELECT name FROM foo WHERE id=?", id).Scan(&s)
  // NullString is safe for query returns NULL
  // ...

  if s.Valid {
    // use s.String
  } else {
    // NULL value
  }
}
```

### Timestamp and date support

We use `time.Time` internally represents SQL timestamp and date type, so you can choose whatever type that is
scannable from `time.Time`, like `sql.NullTime`, or simply `time.Time` itself.


## Getting Start

```go
package main

import (
  "context"
  "database/sql"

  _ "github.com/4paradigm/openmldb-go-sdk"
)

func main() {
  db, err := sql.Open("openmldb", "openmldb://127.0.0.1:8080/test_db")
  if err != nil {
    panic(err)
  }

  defer db.Close()

  ctx := context.Background()

  // execute DDL
  if _, err := db.ExecContext(ctx, `CREATE TABLE demo (c1 int, c2 string);`); err != nil {
    panic(err)
  }

  // execute DML
  if _, err := db.ExecContext(ctx, `INSERT INTO demo VALUES (1, "bb"), (2, "bb");`); err != nil {
    panic(err)
  }

  // execute DQL
  rows, err := db.QueryContext(ctx, `SELECT c1, c2 FROM demo;`)
  if err != nil{
    panic(err)
  }

  var col1 int
  var col2 string

  // iterating query result
  for rows.Next() {
    if err := rows.Scan(&col1, &col2); err != nil {
      panic(err)
    }
    println(col1, col2)
  }
}
```
