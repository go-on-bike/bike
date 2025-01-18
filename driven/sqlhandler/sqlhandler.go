package sqlhandler

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"
)

type sqlHandler struct {
	db      *sql.DB
	options options
	stderr  io.Writer
}

type QueryDB interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Stats() sql.DBStats
	PingContext(ctx context.Context) error
}

const Sig = "sqlhandler.go"

func NewDataHandler(stderr io.Writer, opts ...Option) *sqlHandler {
	defaultDriver := "libsql"
	defaultIdleConns := 25
	defaultOpenConns := 25

	if stderr == nil {
		panic(fmt.Sprintf("%s new: stderr cannot be nil", Sig))
	}

	h := &sqlHandler{stderr: stderr}
	for _, opt := range opts {
		opt(&h.options)
	}
	if h.options.driver == nil {
		h.options.driver = &defaultDriver
	}

	if h.options.maxIdleConns == nil {
		h.options.maxIdleConns = &defaultIdleConns
	}

	if h.options.maxOpenConns == nil {
		h.options.maxOpenConns = &defaultOpenConns
	}

	return h
}

func (h *sqlHandler) connect() error {
	if h.db != nil {
		return fmt.Errorf("%s start: cannot create new connection: database connection already exists", Sig)
	}

	if h.options.driver == nil {
		return fmt.Errorf("%s start: cannot create new conection without a driver", Sig)
	}

	if h.options.url == nil {
		return fmt.Errorf("%s start: cannot create new connection without a url", Sig)
	}

	fmt.Fprintf(h.stderr, "%s start: connecting to url", Sig)
	db, err := sql.Open(*h.options.driver, *h.options.url)
	if err != nil {
		return fmt.Errorf("%s start: failed to open connection with driver %s: %v", Sig, *h.options.driver, err)
	}

	fmt.Fprintf(h.stderr, "%s start: ping to db connection", Sig)
	if err = db.Ping(); err != nil {
		db.Close() // Cerramos la conexión si el ping falla
		return fmt.Errorf("%s start: failed to ping database with driver %s: %v", Sig, *h.options.driver, err)
	}

	fmt.Fprintf(h.stderr, "%s start: connected succesfully", Sig)

	h.db = db
	return nil
}

func (h *sqlHandler) Start(ctx context.Context) error {
	err := h.connect()
	go func() {
		<-ctx.Done()
		h.Shutdown(ctx)
	}()
	return err
}

func (h *sqlHandler) close() error {
	if h.db == nil {
		return fmt.Errorf("%s shutdown: database connection is nil ", Sig)
	}
	err := h.db.Close()
	h.db = nil
	if err != nil {
		return fmt.Errorf("%s shutdown: close on connection failed %v", Sig, err)
	}
	return nil
}

func (h *sqlHandler) Shutdown(ctx context.Context) error {
	fmt.Fprintf(h.stderr, "%s shutdown: closing connection to current connection", Sig)
	if h.db == nil {
		return fmt.Errorf("%s shutdown: database connection is nil ", Sig)
	}

	h.db.SetMaxOpenConns(0)

	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	for {
		stats := h.db.Stats()
		if stats.InUse == 0 {
			return h.close()
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(100 * time.Millisecond)
		}
	}
}

func (h *sqlHandler) QueryDB() QueryDB {
	fmt.Fprintf(h.stderr, "%s db: returning QueryDB", Sig)

	if !isConnected(h.db) {
		panic(fmt.Sprintf("%s db: database connection is nil", Sig))
	}
	return h.db
}

func (h *sqlHandler) IsConnected() bool {
	fmt.Fprintf(h.stderr, "%s is connected: checking if db is still connected", Sig)
	return isConnected(h.db)
}
