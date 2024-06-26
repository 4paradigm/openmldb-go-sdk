package openmldb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
)

func init() {
	sql.Register("openmldb", &openmldbDriver{})
}

// compile time validation that our types implements the expected interfaces
var (
	_ driver.Driver        = openmldbDriver{}
	_ driver.DriverContext = openmldbDriver{}

	_ driver.Connector = connecter{}
)

type openmldbDriver struct{}

func parseDsn(dsn string) (host string, db string, mode queryMode, err error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", "", "", fmt.Errorf("invlaid URL: %w", err)
	}

	if u.Scheme != "openmldb" && u.Scheme != "" {
		return "", "", "", fmt.Errorf("invalid URL: unknown schema '%s'", u.Scheme)
	}

	p := strings.Split(strings.TrimLeft(u.Path, "/"), "/")

	mode = ModeOnline
	if u.Query().Has("mode") {
		m := u.Query().Get("mode")
		if _, ok := allQueryMode[m]; !ok {
			return "", "", "", fmt.Errorf("invalid mode: %s", m)
		}
		mode = allQueryMode[m]
	}

	if len(p) == 0 {
		return "", "", "", fmt.Errorf("invalid URL: DB name not found")
	}

	return u.Host, p[0], mode, nil
}

// Open implements driver.Driver.
func (openmldbDriver) Open(name string) (driver.Conn, error) {
	// name should be the URL of the api server, e.g. openmldb://localhost:6543/db
	host, db, mode, err := parseDsn(name)
	if err != nil {
		return nil, err
	}

	return &conn{host: host, db: db, mode: mode, closed: false}, nil
}

// OpenConnector implements driver.DriverContext.
func (openmldbDriver) OpenConnector(name string) (driver.Connector, error) {
	host, db, mode, err := parseDsn(name)
	if err != nil {
		return nil, err
	}

	return &connecter{host, db, mode}, nil
}

type connecter struct {
	host string
	db   string
	mode queryMode
}

// Connect implements driver.Connector.
func (c connecter) Connect(ctx context.Context) (driver.Conn, error) {
	conn := &conn{host: c.host, db: c.db, mode: c.mode, closed: false}
	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}
	return conn, nil
}

// Driver implements driver.Connector.
func (connecter) Driver() driver.Driver {
	return &openmldbDriver{}
}
