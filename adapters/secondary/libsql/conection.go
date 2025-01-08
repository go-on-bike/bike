package libsql

import (
	"database/sql"

	"github.com/go-on-bike/bike/assert"
)

func (l *Operator) Connect() {
	assert.Nil(l.DB, "libsql cannot create new conection to db, there's one already")

	db, err := sql.Open("libsql", *l.options.url)
	assert.ErrNil(err, "libsql Open failed")

	err = db.Ping()
	assert.ErrNil(err, "libsql Ping failed")

	l.DB = db
}

func (l *Operator) Close() error {
	assert.NotNil(l.DB, "libsql operator cannot run close on nil")

	return l.DB.Close()
}
