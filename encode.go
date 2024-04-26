package openmldb

import (
	"time"
)

func parseDateStr(src string) (time.Time, error) {
	// api server returns date type as string formatted 'yyyy-mm-dd'
	dval, err := time.Parse(time.DateOnly, src)
	if err != nil {
		return time.Time{}, err
	}

	return dval, nil
}
