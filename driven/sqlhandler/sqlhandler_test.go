package sqlhandler

import (
	"strings"
	"testing"

	"github.com/go-on-bike/bike/interfaces"
	_ "github.com/tursodatabase/go-libsql"
)

func TestSQLHandlerIsDataHandler(t *testing.T) {
	dbURL, dbPath := GenTestLibsqlDBPath(t)
	migrPATH := GetMigrationPATH(t)

	stderr := &strings.Builder{}

	var handler interfaces.DataHandler = NewDataHandler(
		stderr,
		[]ConnOption{WithURL(dbURL)},
		[]MigrOption{WithPATH(migrPATH)},
	)

	if err := handler.Connect("libsql"); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := handler.Move(0, false); err != nil {
		t.Fatalf("error on 1st migration: %v", err)
	}

	if err := handler.Move(0, true); err != nil {
		t.Fatalf("error on 2nd migration: %v", err)
	}

	if err := handler.Move(0, false); err != nil {
		t.Fatalf("error on 3rd migration: %v", err)
	}

	version, err := handler.Version()
	if err != nil {
		t.Fatalf("error getting db version: %v", err)
	}

	t.Logf("Version of db is %d", version)

	if err := handler.Close(); err != nil {
		t.Fatalf("error closing database: %v", err)
	}

	AssertDBState(t, dbPath)
}
