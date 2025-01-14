//go:build testing
// +build testing
package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-on-bike/bike/adapters/secondary/sql/connector"
	_ "github.com/tursodatabase/go-libsql"
)

func RunMigrations(t *testing.T, down bool, steps int, firstSteps int) {
	// Verificamos la variable de entorno PWD que necesitamos para las rutas
	pwd := os.Getenv("PWD")
	if pwd == "" {
		t.Fatal("no PWD env var")
		return
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")

	if migrationsPath == "" {
		migrationsPath = filepath.Join(filepath.Dir(filepath.Dir(pwd)), "migrations")
	}

	// Saltamos casos negativos de firstSteps
	if firstSteps < 0 {
		t.Skip()
	}

	// Creamos una base de datos temporal única para cada caso de prueba
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("%s.db", t.Name()))
	dbURL := fmt.Sprintf("file:%s", dbPath)

	// Creamos y configuramos el connector
	c := connector.NewConnector(connector.WithURL(dbURL))

	// Nos aseguramos de que la conexión se cierre al finalizar
	t.Cleanup(func() {
		if c.DB != nil {
			if err := c.Close(); err != nil {
				t.Errorf("error closing database: %v", err)
			}
		}
	})

	// Establecemos la conexión
	err := c.Connect("libsql")
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Creamos el migrador usando la conexión establecida
	m := NewMigrator(c.DB)

	// Primera fase: ejecutamos las migraciones iniciales (setup)
	err = m.Move(migrationsPath, false, firstSteps) // false = up
	if err != nil {
		t.Fatalf("failed initial migration: %v", err)
	}

	// Segunda fase: ejecutamos el caso de fuzzing
	err = m.Move(migrationsPath, down, steps)
	if err != nil {
		// Si steps no es un número válido o hay otro error, verificamos el estado
		AssertDBState(t, dbPath)
		return
	}

	// Verificamos el estado final de la base de datos
	AssertDBState(t, dbPath)
}

func FuzzMigrator(f *testing.F) {
	// Mantenemos los mismos casos semilla que son útiles para probar diferentes escenarios
	f.Add(false, 3, 0) // up, 3 steps
	f.Add(true, 1, 0)  // down, 1 step
	f.Add(false, 2, 3) // up, 2 steps con first steps
	f.Add(false, 0, 0) // up, 0 steps

	// La función principal de fuzzing
	f.Fuzz(RunMigrations)
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
