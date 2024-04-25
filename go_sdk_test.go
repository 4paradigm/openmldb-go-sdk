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

func Test_driver(t *testing.T) {
	db, err := sql.Open("openmldb", fmt.Sprintf("openmldb://%s/test_db", apiServer))
	if err != nil {
		t.Errorf("fail to open connect: %s", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("fail to close connection: %s", err)
		}
	}()

	ctx := context.Background()
	assert.NoError(t, db.PingContext(ctx), "fail to ping connect")

	{
		createTableStmt := "CREATE TABLE demo(c1 int, c2 string, ts timestamp, dt date);"
		_, err := db.ExecContext(ctx, createTableStmt)
		assert.NoError(t, err, "fail to exec %s", createTableStmt)
	}

	defer func() {
		dropTableStmt := "DROP TABLE demo;"
		_, err := db.ExecContext(ctx, dropTableStmt)
		if err != nil {
			t.Errorf("fail to drop table: %s", err)
		}
	}()

	{
		// FIXME: ordering issue
		insertValueStmt := `INSERT INTO demo VALUES (1, "bb", 3000, "2022-12-12");`
		// insertValueStmt := `INSERT INTO demo VALUES (1, "bb"), (2, "bb");`
		_, err := db.ExecContext(ctx, insertValueStmt)
		assert.NoError(t, err, "fail to exec %s", insertValueStmt)
	}

	t.Run("query", func(t *testing.T) {
		queryStmt := `SELECT * FROM demo`
		rows, err := db.QueryContext(ctx, queryStmt)
		assert.NoError(t, err, "fail to query %s", queryStmt)

		var demo struct {
			c1 int32
			c2 string
			ts time.Time
			dt openmldb.NullDate
		}
		{
			assert.True(t, rows.Next())
			assert.NoError(t, rows.Scan(&demo.c1, &demo.c2, &demo.ts, &demo.dt))
			assert.Equal(t, struct {
				c1 int32
				c2 string
				ts time.Time
				dt openmldb.NullDate
			}{1, "bb", time.UnixMilli(3000), openmldb.NullDate{Time: time.Date(2022, time.December, 12, 0, 0, 0, 0, time.UTC), Valid: true}}, demo)
		}
		// {
		// 	assert.True(t, rows.Next())
		// 	assert.NoError(t, rows.Scan(&demo.c1, &demo.c2))
		// 	assert.Equal(t, struct {
		// 		c1 int32
		// 		c2 string
		// 	}{2, "bb"}, demo)
		// }
	})

	t.Run("query with parameter", func(t *testing.T) {
		parameterQueryStmt := `SELECT c1, c2 FROM demo WHERE c2 = ? AND c1 = ?;`
		rows, err := db.QueryContext(ctx, parameterQueryStmt, "bb", 1)
		assert.NoError(t, err, "fail to query %s", parameterQueryStmt)

		var demo struct {
			c1 int32
			c2 string
		}
		{
			assert.True(t, rows.Next())
			assert.NoError(t, rows.Scan(&demo.c1, &demo.c2))
			assert.Equal(t, struct {
				c1 int32
				c2 string
			}{1, "bb"}, demo)
		}
	})
}

func TestMain(m *testing.M) {
	flag.StringVar(&apiServer, "apiserver", "127.0.0.1:9527", "endpoint to apiserver")
	flag.Parse()

	if len(apiServer) == 0 {
		log.Fatalf("non-empty api server address required")
	}

	os.Exit(m.Run())
}
