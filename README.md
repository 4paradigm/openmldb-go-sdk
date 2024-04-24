OpenMLDB Go SDK
------

Pure Go [OpenMLDB](https://github.com/4paradigm/OpenMLDB) driver for database/sql, connect via HTTP.

## Features

## Requirements

- OpenMLDB with all components version >= 0.6.2
- OpenMLDB API Server setted up

## Installation

```sh
go get github.com/4paradigm/openmldb-go-sdk
```

## Data Source Name (DSN)

```
openmldb://<API_SERVER_HOST>:<API_SERVER_PORT>/<DB_NAME>
```

For example, to open a database to `test_db` by api server at `127.0.0.1:8080`:
```go
db, err := sql.Open("openmldb", "openmldb://127.0.0.1:8080/test_db")
```

`<DB_NAME>` is mandatory in DSN, and at this time (version 0.2.0), you must ensure the database `<DB_NAME>` created before open go connection.

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
