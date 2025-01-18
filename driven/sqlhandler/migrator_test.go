package sqlhandler

import (
	"context"
	"strings"
	"testing"
	"time"

	_ "github.com/tursodatabase/go-libsql"
)

func RunMigrations(t *testing.T, inverse bool, steps int, firstSteps int) {
	// Saltamos casos negativos de firstSteps
	if firstSteps < 0 {
		t.Skip()
	}

	// Creamos una base de datos temporal única para cada caso de prueba
	dbURL, dbPath := GenTestLibsqlDBPath(t)
	migrPath := GetMigrationPATH(t)
	stderr := &strings.Builder{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := NewDataHandler(stderr, WithURL(dbURL), WithDriver("libsql"), WithPATH(migrPath))

	// Nos aseguramos de que la conexión se cierre al finalizar
	t.Cleanup(func() {
		// Cancelamos el contexto para iniciar el shutdown
		cancel()
		// Esperamos un poco para que el shutdown se complete
		time.Sleep(100 * time.Millisecond)
	})

	// Establecemos la conexión
	err := h.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start sql handler: %v", err)
	}

	// Primera fase: ejecutamos las migraciones iniciales (setup)
	err = h.RunMigrations(firstSteps, false) // false = up
	if err != nil {
		t.Fatalf("failed initial migration: %v", err)
	}

	// Segunda fase: ejecutamos el caso de fuzzing
	err = h.RunMigrations(steps, inverse)
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
