package sqlhandler

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// CreateTestDB es una función auxiliar que nos ayuda a crear una nueva
// base de datos vacia de prueba para cada test
func GenTestLibsqlDBPath(t *testing.T) (dbURL string, dbPath string) {
	// Creamos un archivo único para cada prueba
	path := filepath.Join(t.TempDir(), "test.db")
	return "file:" + path, path
}

func GetMigrationPATH(t *testing.T) string {
	pwd := os.Getenv("PWD")
	if pwd == "" {
		t.Fatalf("failed to get env var PWD")
	}

	migrPath := os.Getenv("MIGRATION_TEST_PATH")

	if migrPath == "" {
		migrPath = filepath.Join(filepath.Dir(filepath.Dir(pwd)), "migrations")
	}

	return migrPath
}

// AssertDBState verifica que la base de datos está en un estado válido
func AssertDBState(t *testing.T, dbPath string) {
	stats, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("failed to stat database: %v", err)
	}
	if stats.Size() == 0 {
		t.Fatal("database is empty")
	}
	if stats.Size() < 8192 {
		t.Fatal("database appears to have no migrations")
	}
}

func NewTestsqlHandler(t *testing.T, stderr io.Writer) (*sqlHandler, string) {
	dbURL, dbPath := GenTestLibsqlDBPath(t)
	c := NewDataHandler(stderr, WithURL(dbURL))
	return c, dbPath
}
