package sqlhandler

import (
	"database/sql"
	"io"
)

type SQLHandler struct {
	*Connector
	*Migrator
	stderr io.Writer
}

const SigSQLHandler string = "sqlhandler"

func NewDataHandler(stderr io.Writer, connOpts []ConnOption, migrOpts []MigrOption) *SQLHandler {
	c := NewConnector(stderr, connOpts...)
	m := NewMigrator(stderr, nil, migrOpts...)

	return &SQLHandler{stderr: stderr, Connector: c, Migrator: m}
}

func (h *SQLHandler) Connect(driver string) error {
	if err := h.Connector.Connect(driver); err != nil {
		return err
	}
	h.Migrator.db = h.Connector.db
	return nil
}

func (h *SQLHandler) Close() error {
	if err := h.Connector.Close(); err != nil {
		return err
	}
	h.Migrator.db = nil
	return nil
}

func (h *SQLHandler) SetDB(db *sql.DB) {
	h.Connector.SetDB(db)
	h.Migrator.SetDB(db)
}
