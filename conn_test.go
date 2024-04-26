package openmldb

import (
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseReqToJson(t *testing.T) {
	for _, tc := range []struct {
		mode   string
		sql    string
		input  []driver.Value
		expect string
	}{
		{
			"offline",
			"SELECT 1;",
			nil,
			`{
				"mode": "offline",
				"sql": "SELECT 1;"
			}`,
		},
		{
			"online",
			"SELECT c1, c2 FROM demo WHERE c1 = ? AND c2 = ?;",
			[]driver.Value{
				int16(2), // int16
				int32(1), // int32
				"bb",     // string
				Null[string]{Null: sql.Null[string]{V: "foo", Valid: true}}, // string
				time.UnixMilli(8000), // timestamp
				Null[time.Time]{Null: sql.Null[time.Time]{V: time.UnixMilli(4000), Valid: true}},                              // timestamp
				Null[time.Time]{Null: sql.Null[time.Time]{V: time.UnixMilli(4000), Valid: false}},                             // timestamp
				NullDate{Null: sql.Null[time.Time]{V: time.Date(2022, time.October, 10, 0, 0, 0, 0, time.UTC), Valid: true}}}, // date
			`{
				"mode": "online",
				"sql": "SELECT c1, c2 FROM demo WHERE c1 = ? AND c2 = ?;",
				"input": {
					"schema": ["int16", "int32", "string", "string", "timestamp", "timestamp", "timestamp", "date"],
					"data": [2, 1, "bb", "foo", 8000, 4000, null, "2022-10-10"]
				}
			}`,
		},
	} {
		actual, err := marshalQueryRequest(tc.mode, tc.sql, tc.input...)
		assert.NoError(t, err)
		assert.JSONEq(t, tc.expect, string(actual))
	}
}

func TestParseRespFromJson(t *testing.T) {
	for _, tc := range []struct {
		resp   string
		expect queryResp
	}{
		{
			`{
				"code": 0,
				"msg": "ok"
			}`,
			queryResp{
				Code: 0,
				Msg:  "ok",
				Data: nil,
			},
		},
		{
			`{
				"code": 0,
				"msg": "ok",
				"data": {
					"schema": ["date", "string"],
					"data": []
				}
			}`,
			queryResp{
				Code: 0,
				Msg:  "ok",
				Data: &respData{
					Schema: []string{"date", "string"},
					Data:   [][]driver.Value{},
				},
			},
		},
		{
			`{
				"code": 0,
				"msg": "ok",
				"data": {
					"schema": ["Int32", "String"],
					"data": [[1, "bb"], [2, "bb"]]
				}
			}`,
			queryResp{
				Code: 0,
				Msg:  "ok",
				Data: &respData{
					Schema: []string{"Int32", "String"},
					Data: [][]driver.Value{
						{int32(1), "bb"},
						{int32(2), "bb"},
					},
				},
			},
		},
		{
			`{
				"code": 0,
				"msg": "ok",
				"data": {
					"schema": ["Bool", "Int16", "Int32", "Int64", "Float", "Double", "String"],
					"data": [[true, 1, 1, 1, 1, 1, "bb"]]
				}
			}`,
			queryResp{
				Code: 0,
				Msg:  "ok",
				Data: &respData{
					Schema: []string{"Bool", "Int16", "Int32", "Int64", "Float", "Double", "String"},
					Data: [][]driver.Value{
						{true, int16(1), int32(1), int64(1), float32(1), float64(1), "bb"},
					},
				},
			},
		},
	} {
		actual, err := unmarshalQueryResponse(strings.NewReader(tc.resp))
		assert.NoError(t, err)
		assert.Equal(t, &tc.expect, actual)
	}
}
