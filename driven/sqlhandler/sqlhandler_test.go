package sqlhandler

import (
	"testing"

	_ "github.com/tursodatabase/go-libsql"
)

func TestSQLHandlerComposition(t *testing.T) {
	dbURL, dbPath := GetDBLocation(t)
	migrPATH := GetMigrationPATH(t)

	handler := NewHandler([]ConnOption{WithURL(dbURL)}, []MigrOption{WithPATH(migrPATH)})

	if err := handler.Connect("libsql"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Log(handler.Connector.DB)
	t.Log(handler.Migrator.DB)
	// t.Log(*handler.db)

	// t.Log(handler.DB)
	// t.Log(&handler.DB)

	if err := handler.Move(false, 0); err != nil {
		t.Fatalf("error on 1st migration: %v", err)
	}

	// if err := handler.Move(true, 0); err != nil {
	// t.Fatalf("error on 2nd migration: %v", err)
	// }

	// if err := handler.Move(false, 0); err != nil {
	// t.Fatalf("error on 3rd migration: %v", err)
	// }

	// if err := handler.Close(); err != nil {
	// t.Fatalf("error closing database: %v", err)
	// }

	AssertDBState(t, dbPath)
}
