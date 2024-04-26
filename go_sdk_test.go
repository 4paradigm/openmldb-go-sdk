package openmldb_test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	// register openmldb driver
	"github.com/stretchr/testify/assert"

	openmldb "github.com/4paradigm/openmldb-go-sdk"
)

var apiServer string

var db *sql.DB
var ctx context.Context

// user may use sql.NullXXX types to represent SQL values that may be null

type demoStruct1 struct {
	c1 int32
	c2 string
	ts time.Time
	dt time.Time
}
type demoStruct2 struct {
	c1 sql.NullInt32
	c2 sql.NullString
	ts sql.NullTime
	dt sql.NullTime
}
type demoStruct3 struct {
	c1 openmldb.Null[int32]
	c2 openmldb.Null[string]
	ts openmldb.Null[time.Time]
	dt openmldb.NullDate
}

func TestPingCtx(t *testing.T) {
	assert.NoError(t, db.PingContext(ctx), "fail to ping connect")
}

func TestQuery1(t *testing.T) {
	// use time.Time to represent both timestamp and date
	queryStmt := `SELECT * FROM demo`
	rows, err := db.QueryContext(ctx, queryStmt)
	assert.NoError(t, err, "fail to query %s", queryStmt)

	var demo demoStruct1
	{
		assert.True(t, rows.Next())
		assert.NoError(t, rows.Scan(&demo.c1, &demo.c2, &demo.ts, &demo.dt))
		assert.Equal(t, demoStruct1{1, "bb", time.UnixMilli(3000), time.Date(2022, time.December, 12, 0, 0, 0, 0, time.UTC)}, demo)
	}
	// {
	// 	assert.True(t, rows.Next())
	// 	assert.NoError(t, rows.Scan(&demo.c1, &demo.c2))
	// 	assert.Equal(t, struct {
	// 		c1 int32
	// 		c2 string
	// 	}{2, "bb"}, demo)
	// }
}

func TestQuery2(t *testing.T) {
	// use sql.NullTime to represent both timestamp and date
	queryStmt := `SELECT * FROM demo`
	rows, err := db.QueryContext(ctx, queryStmt)
	assert.NoError(t, err, "fail to query %s", queryStmt)

	var demo demoStruct2
	assert.True(t, rows.Next())
	assert.NoError(t, rows.Scan(&demo.c1, &demo.c2, &demo.ts, &demo.dt))
	assert.Equal(t, sql.NullInt32{Int32: 1, Valid: true}, demo.c1)
	assert.Equal(t, sql.NullString{String: "bb", Valid: true}, demo.c2)
	assert.Equal(t, sql.NullTime{Time: time.UnixMilli(3000), Valid: true}, demo.ts)
	assert.Equal(t, sql.NullTime{Time: time.Date(2022, time.December, 12, 0, 0, 0, 0, time.UTC), Valid: true}, demo.dt)
}

func TestQuery3(t *testing.T) {
	// use openmldb.Null[T] and openmldb.NullDate to represent timestamp and date
	queryStmt := `SELECT * FROM demo`
	rows, err := db.QueryContext(ctx, queryStmt)
	assert.NoError(t, err, "fail to query %s", queryStmt)

	var demo demoStruct3
	assert.True(t, rows.Next())
	assert.NoError(t, rows.Scan(&demo.c1, &demo.c2, &demo.ts, &demo.dt))
	assert.Equal(t, openmldb.Null[int32]{Null: sql.Null[int32]{V: 1, Valid: true}}, demo.c1)
	assert.Equal(t, openmldb.Null[string]{Null: sql.Null[string]{V: "bb", Valid: true}}, demo.c2)
	assert.Equal(t, openmldb.Null[time.Time]{Null: sql.Null[time.Time]{V: time.UnixMilli(3000), Valid: true}}, demo.ts)
	assert.Equal(t, openmldb.NullDate{Null: sql.Null[time.Time]{V: time.Date(2022, time.December, 12, 0, 0, 0, 0, time.UTC), Valid: true}}, demo.dt)
}

func TestQueryWithParams(t *testing.T) {
	parameterQueryStmt := `SELECT * FROM demo WHERE c2 = ? AND c1 = ? AND ts = ?;`
	rows, err := db.QueryContext(ctx, parameterQueryStmt, "bb", 1, time.UnixMilli(3000))
	assert.NoError(t, err, "fail to query %s", parameterQueryStmt)

	var demo demoStruct1
	{
		assert.True(t, rows.Next())
		assert.NoError(t, rows.Scan(&demo.c1, &demo.c2, &demo.ts, &demo.dt))
		assert.Equal(t, demoStruct1{1, "bb", time.UnixMilli(3000), time.Date(2022, time.December, 12, 0, 0, 0, 0, time.UTC)}, demo)
	}
}

func TestQueryWithParamsExpectsNull(t *testing.T) {
	_, err := db.ExecContext(ctx, "create table test2 (id int16, val int64, dt date)")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.ExecContext(ctx, "drop table test2")
		assert.NoError(t, err)
	})

	{
		_, err := db.ExecContext(ctx, "insert into test2 values (1, NULL, NULL)")
		assert.NoError(t, err)
	}

	rows, err := db.QueryContext(ctx, "select * from test2 where id = ?", 1)
	assert.NoError(t, err)
	var demo struct {
		id  sql.NullInt16
		val sql.NullInt64
		dt  sql.NullTime
	}
	{
		assert.True(t, rows.Next())
		assert.NoError(t, rows.Scan(&demo.id, &demo.val, &demo.dt))
		assert.Equal(t, sql.NullInt16{Int16: 1, Valid: true}, demo.id)
		assert.Equal(t, sql.NullInt64{Int64: 0, Valid: false}, demo.val)
		assert.Equal(t, sql.NullTime{Time: time.Time{}, Valid: false}, demo.dt)
	}
}

func TestQueryWithParamsResultsEmpty(t *testing.T) {
	_, err := db.ExecContext(ctx, "create table test3 (id int16, val int64, dt date)")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.ExecContext(ctx, "drop table test3")
		assert.NoError(t, err)
	})

	{
		_, err := db.ExecContext(ctx, "insert into test3 values (1, NULL, NULL)")
		assert.NoError(t, err)
	}

	{
		rows, err := db.QueryContext(ctx, "select * from test3 where id = ?", int16(10))
		assert.NoError(t, err)
		assert.False(t, rows.Next())
	}

	{
		// disabled since https://github.com/4paradigm/OpenMLDB/issues/3902
		// _, err := db.QueryContext(ctx, "select * from test3 where id = ?",
		// 	openmldb.Null[int16]{Null: sql.Null[int16]{V: 0, Valid: false}})
		// assert.NoError(t, err)
		// assert.False(t, rows.Next())
	}
}

func PrepareAndRun(m *testing.M) int {
	var err error
	db, err = sql.Open("openmldb", fmt.Sprintf("openmldb://%s/test_db", apiServer))
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to open connect: %s", err)
		os.Exit(1)
	}

	ctx = context.Background()

	{
		createTableStmt := "CREATE TABLE demo(c1 int, c2 string, ts timestamp, dt date);"
		_, err := db.ExecContext(ctx, createTableStmt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to exec %s", createTableStmt)
			os.Exit(1)
		}
	}

	defer func() {
		dropTableStmt := "DROP TABLE demo;"
		_, err := db.ExecContext(ctx, dropTableStmt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to drop table: %s", err)
			os.Exit(1)
		}
	}()
	{
		// FIXME: ordering issue
		insertValueStmt := `INSERT INTO demo VALUES (1, "bb", 3000, "2022-12-12");`
		// insertValueStmt := `INSERT INTO demo VALUES (1, "bb"), (2, "bb");`
		_, err := db.ExecContext(ctx, insertValueStmt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to exec: %s", insertValueStmt)
			os.Exit(1)
		}
	}

	return m.Run()

}

func TestMain(m *testing.M) {
	flag.StringVar(&apiServer, "apiserver", "127.0.0.1:9527", "endpoint to apiserver")
	flag.Parse()

	if len(apiServer) == 0 {
		log.Fatalf("non-empty api server address required")
	}

	os.Exit(PrepareAndRun(m))
}
