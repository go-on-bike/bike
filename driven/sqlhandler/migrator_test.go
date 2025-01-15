package sqlhandler

import (
	"testing"

	"github.com/go-on-bike/bike/interfaces"
	_ "github.com/tursodatabase/go-libsql"
)

func RunMigrations(t *testing.T, inverse bool, steps int, firstSteps int) {
	// Saltamos casos negativos de firstSteps
	if firstSteps < 0 {
		t.Skip()
	}

	// Creamos una base de datos temporal única para cada caso de prueba
	dbURL, dbPath := GetDBLocation(t)
	migrPath := GetMigrationPATH(t)
	c := NewConnector(WithURL(dbURL))

	// Nos aseguramos de que la conexión se cierre al finalizar
	t.Cleanup(func() {
		if c.db != nil {
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
	m := NewMigrator(c.db, WithPATH(migrPath))

	// Primera fase: ejecutamos las migraciones iniciales (setup)
	err = m.Move(firstSteps, false) // false = up
	if err != nil {
		t.Fatalf("failed initial migration: %v", err)
	}

	// Segunda fase: ejecutamos el caso de fuzzing
	err = m.Move(steps, inverse)
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

func TestMigratorInterface(t *testing.T) {
	t.Run("sql migrator is migrator", func(t *testing.T) {
		dbURL, dbPath := GetDBLocation(t)

		c := NewConnector(WithURL(dbURL))

		if err := c.Connect("libsql"); err != nil {
			t.Fatalf("unexpected error connecting: %v", err)
		}
		if err := c.Close(); err != nil {
			t.Fatalf("unexpected error closing: %v", err)
		}

		t.Cleanup(func() {
			if c.db != nil {
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

		migrPath := GetMigrationPATH(t)
		var m interfaces.Migrator = NewMigrator(c.db, WithPATH(migrPath))

		if err := m.Move(0, false); err != nil {
			t.Fatalf("failed initial migration: %v", err)
		}

        version, err := m.Version()
        if err != nil {
            t.Fatalf("failed to get version of db")
        }

        t.Logf("version of db is %d", version)

		// Verificamos el estado final de la base de datos
		AssertDBState(t, dbPath)
	})
}
