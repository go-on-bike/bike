package libsql

import (
	"database/sql"
)

type Operator struct {
	db      *sql.DB
	options options
}

func NewOperator(opts ...Option) *Operator {
	l := Operator{}
	for _, opt := range opts {
		opt(&l.options)
	}

	return &l
}
