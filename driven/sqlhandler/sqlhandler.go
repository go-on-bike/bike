package sqlhandler

type SQLHandler struct {
    Migrator
    Connector
}

func New(options ...ConnOption) *SQLHandler {
	c := NewConnector(options...)
	m := NewMigrator(c.DB)
    return &SQLHandler{Connector: *c, Migrator: *m}
}
