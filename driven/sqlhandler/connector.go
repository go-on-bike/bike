package sqlhandler

import (
	"database/sql"
	"fmt"
)

type Connector struct {
	db      *sql.DB
	options connOpts
}

// NewConnector crea un nuevo conector de base de datos.
// Panics si se proporciona una URL vacía, ya que esto representa un error de programación.
func NewConnector(opts ...ConnOption) *Connector {
	c := &Connector{}
	for _, opt := range opts {
		opt(&c.options)
	}
	if c.options.url == nil {
		panic("database URL is required")
	}
	return c
}

// Connect establece una conexión con la base de datos usando el driver especificado.
// Retorna un error si ya existe una conexión o si hay problemas al conectar.
func (c *Connector) Connect(driver string) error {
	if c.db != nil {
		return fmt.Errorf("cannot create new connection: database connection already exists")
	}

	db, err := sql.Open(driver, *c.options.url)
	if err != nil {
		return fmt.Errorf("failed to open connection with driver %s: %w", driver, err)
	}

	if err = db.Ping(); err != nil {
		db.Close() // Cerramos la conexión si el ping falla
		return fmt.Errorf("failed to ping database with driver %s: %w", driver, err)
	}

	c.db = db
	return nil
}

// Close cierra la conexión a la base de datos.
// Panic si se intenta cerrar una conexión nil, ya que esto representa un error de programación.
func (c *Connector) Close() error {
	if c.db == nil {
		panic("cannot close: database connection is nil")
	}
	err := c.db.Close()
	if err != nil {
		return err
	}

	c.db = nil
	return nil
}
