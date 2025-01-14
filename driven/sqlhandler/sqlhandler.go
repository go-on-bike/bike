package sqlhandler

type SQLHandler struct {
    Migrator
    Connector
}

func NewHandler(connOpts []ConnOption, migrOpts []MigrOption) *SQLHandler {
	c := NewConnector(connOpts...)
	m := NewMigrator(c.DB, migrOpts...)
    return &SQLHandler{Connector: *c, Migrator: *m}
}
