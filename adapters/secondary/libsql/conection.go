package libsql

import (
	"database/sql"

	"github.com/go-on-bike/bike/assert"
)

func (l *Operator) Connect() {
	assert.Nil(l.db, "libsql cannot create new conection to db, there's one already")

	db, err := sql.Open("libsql", *l.options.url)
	assert.ErrNil(err, "libsql Open failed")

	err = db.Ping()
	assert.ErrNil(err, "libsql Ping failed")

	l.db = db
}

func (l *Operator) Close() error {
	assert.NotNil(l.db, "libsql operator cannot run close on nil")

	return l.db.Close()
}
