package libsql

import (
	"database/sql"
	"io"

	"github.com/matisin/bike/assert"
)

type LibSqlOperator struct {
	db      *sql.DB
	options options
}

func NewLibSqlOperator(stdout io.Writer, outformat string, opts ...Option) *LibSqlOperator {
	l := LibSqlOperator{}
	for _, opt := range opts {
		opt(&l.options)
	}

	return &l
}

func (l *LibSqlOperator) Connect() {
	assert.DBNil(l.db)

	db, err := sql.Open("libsql", *l.options.url)
	assert.ErrNil(err, "DB connection is not nil")

	err = db.Ping()
	assert.ErrNil(err, "DB ping failed")

	l.db = db
}

func (l *LibSqlOperator) Close() error {
	return l.db.Close()
}
