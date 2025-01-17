package sqlhandler

import (
	"database/sql"
	"fmt"
	"io"
)

type connOpts struct {
	url *string
}

type Connector struct {
	stderr  io.Writer
	db      *sql.DB
	options connOpts
}

const SigConn string = "sqlhandler connector"

// NewConnector crea un nuevo conector de base de datos.
// Panics si se proporciona una URL vacía, ya que esto representa un error de programación.
func NewConnector(stderr io.Writer, opts ...ConnOption) *Connector {
	if stderr == nil {
		panic(fmt.Sprintf("%s: stderr cannot be nil", SigConn))
	}

	c := &Connector{stderr: stderr}
	for _, opt := range opts {
		opt(&c.options)
	}
	return c
}

// Connect establece una conexión con la base de datos usando el driver especificado.
// Retorna un error si ya existe una conexión o si hay problemas al conectar.
func (c *Connector) Connect(driver string) error {
	if c.db != nil {
		return fmt.Errorf("%s: cannot create new connection: database connection already exists", SigConn)
	}

	fmt.Fprintf(c.stderr, "%s: connecting to url", SigConn)

	db, err := sql.Open(driver, *c.options.url)
	if err != nil {
		return fmt.Errorf("%s: failed to open connection with driver %s: %v", SigConn, driver, err)
	}

	fmt.Fprintf(c.stderr, "%s: ping to db connection", SigConn)
	if err = db.Ping(); err != nil {
		db.Close() // Cerramos la conexión si el ping falla
		return fmt.Errorf("%s: failed to ping database with driver %s: %v", SigConn, driver, err)
	}
	fmt.Fprintf(c.stderr, "%s: connected succesfully", SigConn)

	c.db = db
	return nil
}

// Close cierra la conexión a la base de datos.(
// Panic si se intenta cerrar una conexión nil, ya que esto representa un error de programación.
func (c *Connector) Close() error {
	if c.db == nil {
		panic(fmt.Sprintf("%s: database connection is nil", SigConn))
	}
	fmt.Fprintf(c.stderr, "%s: closing connection to current connection", SigConn)
	err := c.db.Close()
	if err != nil {
		return fmt.Errorf("%s: close on connection failed %v", SigConn, err)
	}

	c.db = nil
	return nil
}

func (c *Connector) SetDB(db *sql.DB) {
	fmt.Fprintf(c.stderr, "%s: setting new db", SigConn)
	if isConnected(c.db) {
		panic(fmt.Sprintf("%s: cannot change connected connection", SigConn))
	}

	if !isConnected(db) {
		panic(fmt.Sprintf("%s: cannot change connection to a closed one", SigConn))
	}

	c.db = db
}

func (c *Connector) IsConnected() bool {
	fmt.Fprintf(c.stderr, "%s: checking if db is still conected", SigConn)
	return isConnected(c.db)
}

func (c *Connector) DB() (db *sql.DB) {
	// TODO: deprecate this function
	fmt.Fprintf(c.stderr, "%s: returning DB of connector", SigConn)

	if !isConnected(c.db) {
		panic("DB is nil")
	}
	return c.db
}
