package openmldb_test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	// register openmldb driver
	_ "github.com/4paradigm/openmldb-go-sdk"
	"github.com/stretchr/testify/assert"
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
		createTableStmt := "CREATE TABLE demo(c1 int, c2 string);"
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
		insertValueStmt := `INSERT INTO demo VALUES (1, "bb");`
		// insertValueStmt := `INSERT INTO demo VALUES (1, "bb"), (2, "bb");`
		_, err := db.ExecContext(ctx, insertValueStmt)
		assert.NoError(t, err, "fail to exec %s", insertValueStmt)
	}

	t.Run("query", func(t *testing.T) {
		queryStmt := `SELECT c1, c2 FROM demo`
		rows, err := db.QueryContext(ctx, queryStmt)
		assert.NoError(t, err, "fail to query %s", queryStmt)

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
