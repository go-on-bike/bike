package sqlhandler

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-on-bike/bike/interfaces"
	_ "github.com/tursodatabase/go-libsql"
)

func TestSQLHandlerIsDataHandler(t *testing.T) {
	dbURL, dbPath := GenTestLibsqlDBPath(t)
	migrPATH := GetMigrationPATH(t)

	stderr := &strings.Builder{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var handler interfaces.DataHandler = NewDataHandler(
		stderr,
		WithURL(dbURL),
		WithDriver("libsql"),
		WithPATH(migrPATH),
	)

	// Nos aseguramos de que la conexión se cierre al finalizar
	t.Cleanup(func() {
		// Cancelamos el contexto para iniciar el shutdown
		cancel()
		// Esperamos un poco para que el shutdown se complete
		time.Sleep(100 * time.Millisecond)
	})

	if err := handler.Start(ctx); err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := handler.RunMigrations(0, false); err != nil {
		t.Fatalf("error on 1st migration: %v", err)
	}

	if err := handler.RunMigrations(0, true); err != nil {
		t.Fatalf("error on 2nd migration: %v", err)
	}

	if err := handler.RunMigrations(0, false); err != nil {
		t.Fatalf("error on 3rd migration: %v", err)
	}

	version, err := handler.Version()
	if err != nil {
		t.Fatalf("error getting db version: %v", err)
	}

	t.Logf("Version of db is %d", version)

	// En lugar de llamar Shutdown explícitamente
	cancel() // Trigger cleanup
	time.Sleep(100 * time.Millisecond)

	// Verificar que el cleanup funcionó
	if handler.IsConnected() {
		t.Error("handler still connected after context cancellation")
	}

	AssertDBState(t, dbPath)
}

// TestNewConnector verifica la creación correcta del Connector
func TestNewHandler(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		dbURL, _ := GenTestLibsqlDBPath(t)
		stderr := &strings.Builder{}

		h := NewDataHandler(stderr, WithURL(dbURL))
		if h == nil {
			t.Fatal("expected non-nil Connector")
		}
		if h.options.url == nil {
			t.Fatal("expected non-nil URL option")
		}
	})

	t.Run("empty url panics", func(t *testing.T) {
		stderr := &strings.Builder{}
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic with empty URL")
			}
		}()
		NewDataHandler(stderr, WithURL(""))
	})

	t.Run("nil stderr panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic with empty URL")
			}
		}()
		NewDataHandler(nil, WithURL(""))
	})

	t.Run("nil stderr panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic with not accepted driver")
			}
		}()
		NewDataHandler(nil, WithDriver("driver"))
	})
}

// TestConnectorConnect verifica todas las operaciones de conexión
func TestConnect(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		dbURL, _ := GenTestLibsqlDBPath(t)
		stderr := &strings.Builder{}

		h := NewDataHandler(stderr, WithURL(dbURL), WithDriver("libsql"))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Nos aseguramos de que la conexión se cierre al finalizar
		t.Cleanup(func() {
			// Cancelamos el contexto para iniciar el shutdown
			cancel()
			// Esperamos un poco para que el shutdown se complete
			time.Sleep(100 * time.Millisecond)
		})

		err := h.Start(ctx)
		if err != nil {
			t.Fatalf("unexpected error connecting: %v", err)
		}
		if h.db == nil {
			t.Fatal("expected non-nil db after connection")
		}
	})

	t.Run("default driver", func(t *testing.T) {
		dbURL, _ := GenTestLibsqlDBPath(t)
		stderr := &strings.Builder{}

		h := NewDataHandler(stderr, WithURL(dbURL))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Nos aseguramos de que la conexión se cierre al finalizar
		t.Cleanup(func() {
			// Cancelamos el contexto para iniciar el shutdown
			cancel()
			// Esperamos un poco para que el shutdown se complete
			time.Sleep(100 * time.Millisecond)
		})

		err := h.Start(ctx)
		if err != nil {
			t.Fatal("expected no error connecting")
		}
	})

	t.Run("double connection attempt", func(t *testing.T) {
		dbURL, _ := GenTestLibsqlDBPath(t)
		stderr := &strings.Builder{}

		h := NewDataHandler(stderr, WithURL(dbURL), WithDriver("libsql"))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Nos aseguramos de que la conexión se cierre al finalizar
		t.Cleanup(func() {
			// Cancelamos el contexto para iniciar el shutdown
			cancel()
			// Esperamos un poco para que el shutdown se complete
			time.Sleep(100 * time.Millisecond)
		})

		err := h.Start(ctx)
		if err != nil {
			t.Fatalf("unexpected error on first connect: %v", err)
		}

		// Intentar conectar de nuevo debería fallar
		err = h.Start(ctx)
		if err == nil {
			t.Fatal("expected error on second connect")
		}
	})

	t.Run("malformed url", func(t *testing.T) {
		stderr := &strings.Builder{}

		c := NewDataHandler(stderr, WithURL("invalid://url"))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Nos aseguramos de que la conexión se cierre al finalizar
		t.Cleanup(func() {
			// Cancelamos el contexto para iniciar el shutdown
			cancel()
			// Esperamos un poco para que el shutdown se complete
			time.Sleep(100 * time.Millisecond)
		})
		err := c.Start(ctx)
		if err == nil {
			t.Fatal("expected error with invalid URL")
		}
	})
}

// // TestConnectorClose verifica el comportamiento del cierre de conexiones
// func TestConnectorClose(t *testing.T) {
// t.Run("successful close", func(t *testing.T) {
// dbURL, _ := GenTestLibsqlDBPath(t)
// stderr := &strings.Builder{}

// h := NewDataHandler(stderr, WithURL(dbURL), WithDriver("libsql"))

// err := h.Start()
// if err != nil {
// t.Fatalf("unexpected error connecting: %v", err)
// }

// err = h.Shutdown()
// if err != nil {
// t.Fatalf("unexpected error closing: %v", err)
// }
// })

// t.Run("close without connect panics", func(t *testing.T) {
// dbURL, _ := GenTestLibsqlDBPath(t)
// stderr := &strings.Builder{}

// h := NewDataHandler(stderr, WithURL(dbURL), WithDriver("libsql"))

// defer func() {
// if r := recover(); r == nil {
// t.Fatal("expected panic when closing without connection")
// }
// }()
// h.Shutdown()
// })

// t.Run("DB() without connect panics", func(t *testing.T) {
// dbURL, _ := GenTestLibsqlDBPath(t)
// stderr := &strings.Builder{}

// h := NewDataHandler(stderr, WithURL(dbURL), WithDriver("libsql"))

// defer func() {
// if r := recover(); r == nil {
// t.Fatal("expected panic when getting DB without connection")
// }
// }()
// h.DB()
// })

// t.Run("checking functionality of IsConnected", func(t *testing.T) {
// dbURL, _ := GenTestLibsqlDBPath(t)
// stderr := &strings.Builder{}

// h := NewDataHandler(stderr, WithURL(dbURL), WithDriver("libsql"))

// connected := h.IsConnected()
// if connected {
// t.Fatal("connected should false")
// }

// err := h.Start()
// if err != nil {
// t.Fatalf("unexpected error connecting: %v", err)
// }

// connected = h.IsConnected()
// if !connected {
// t.Fatal("connected should true")
// }

// h.Shutdown()
// connected = h.IsConnected()
// if connected {
// t.Fatal("connected should false")
// }
// })

// t.Run("double close", func(t *testing.T) {
// dbURL, _ := GenTestLibsqlDBPath(t)
// stderr := &strings.Builder{}

// h := NewDataHandler(stderr, WithURL(dbURL), WithDriver("libsql"))

// err := h.Start()
// if err != nil {
// t.Fatalf("unexpected error connecting: %v", err)
// }

// err = h.Shutdown()
// if err != nil {
// t.Fatalf("unexpected error on first close: %v", err)
// }

// defer func() {
// if r := recover(); r == nil {
// t.Fatal("expected panic on second close")
// }
// }()
// h.Shutdown()
// })
// }

// TestConnectorIntegration verifica el funcionamiento completo del connector
// realizando operaciones reales en la base de datos
func TestConnectorIntegration(t *testing.T) {
	dbURL, _ := GenTestLibsqlDBPath(t)
	stderr := &strings.Builder{}

	h := NewDataHandler(stderr, WithURL(dbURL), WithDriver("libsql"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Nos aseguramos de que la conexión se cierre al finalizar
	t.Cleanup(func() {
		// Cancelamos el contexto para iniciar el shutdown
		cancel()
		// Esperamos un poco para que el shutdown se complete
		time.Sleep(100 * time.Millisecond)
	})

	err := h.Start(ctx)
	if err != nil {
		t.Fatalf("unexpected error connecting: %v", err)
	}

	// Configuramos el modo WAL para mejor rendimiento
	var journalMode string
	err = h.db.QueryRow("PRAGMA journal_mode=WAL").Scan(&journalMode)
	if err != nil {
		t.Fatalf("failed to set WAL mode: %v", err)
	}
	// El valor retornado debería ser "wal"
	if journalMode != "wal" {
		t.Fatalf("expected journal_mode to be 'wal', got '%s'", journalMode)
	}

	// Creamos una tabla de prueba
	_, err = h.db.Exec(`
        CREATE TABLE test (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL
        )
    `)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	// Insertamos varios registros de prueba
	testData := []string{"Alice", "Bob", "Charlie"}
	for _, name := range testData {
		_, err = h.db.Exec("INSERT INTO test (name) VALUES (?)", name)
		if err != nil {
			t.Fatalf("failed to insert test data '%s': %v", name, err)
		}
	}

	// Verificamos que podemos leer todos los datos insertados
	rows, err := h.db.Query("SELECT name FROM test ORDER BY id")
	if err != nil {
		t.Fatalf("failed to query test data: %v", err)
	}
	defer rows.Close()

	var retrievedNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		retrievedNames = append(retrievedNames, name)
	}

	// Verificamos que no hubo errores durante la iteración
	if err = rows.Err(); err != nil {
		t.Fatalf("error during row iteration: %v", err)
	}

	// Comparamos los resultados
	if len(retrievedNames) != len(testData) {
		t.Fatalf("expected %d names, got %d", len(testData), len(retrievedNames))
	}
	for i, expected := range testData {
		if retrievedNames[i] != expected {
			t.Fatalf("at position %d: expected '%s', got '%s'", i, expected, retrievedNames[i])
		}
	}
}
