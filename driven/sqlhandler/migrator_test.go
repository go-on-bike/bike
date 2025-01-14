//go:build testing

package sqlhandler

import (
	"testing"

	_ "github.com/tursodatabase/go-libsql"
)

func RunMigrations(t *testing.T, down bool, steps int, firstSteps int) {
	migrPath := GetMigrationPATH(t)
	// Verificamos la variable de entorno PWD que necesitamos para las rutas

	// Saltamos casos negativos de firstSteps
	if firstSteps < 0 {
		t.Skip()
	}

	// Creamos una base de datos temporal única para cada caso de prueba
	dbPath, dbURL := GetDBLocation(t)

	// Creamos y configuramos el connector
	c := NewConnector(WithURL(dbURL))

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
	m := NewMigrator(&c.DB, WithPATH(migrPath))

	// Primera fase: ejecutamos las migraciones iniciales (setup)
	err = m.Move(false, firstSteps) // false = up
	if err != nil {
		t.Fatalf("failed initial migration: %v", err)
	}

	// Segunda fase: ejecutamos el caso de fuzzing
	err = m.Move(down, steps)
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
