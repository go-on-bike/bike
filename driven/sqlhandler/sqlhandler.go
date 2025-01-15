package sqlhandler

type SQLHandler struct {
	*Connector
	*Migrator
}

func NewDataHandler(connOpts []ConnOption, migrOpts []MigrOption) *SQLHandler {
	c := NewConnector(connOpts...)
	m := NewMigrator(nil, migrOpts...)
	return &SQLHandler{Connector: c, Migrator: m}
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
