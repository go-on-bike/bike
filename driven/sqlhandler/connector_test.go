

package sqlhandler

import (
	_ "github.com/tursodatabase/go-libsql"
	"testing"
)

// TestNewConnector verifica la creación correcta del Connector
func TestNewConnector(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		dbURL, _ := GetDBLocation(t)
		c := NewConnector(WithURL(dbURL))
		if c == nil {
			t.Fatal("expected non-nil Connector")
		}
		if c.options.url == nil {
			t.Fatal("expected non-nil URL option")
		}
	})

	t.Run("empty url panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic with empty URL")
			}
		}()
		NewConnector(WithURL(""))
	})

	t.Run("no options panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic with no options")
			}
		}()
		NewConnector()
	})
}

// TestConnectorConnect verifica todas las operaciones de conexión
func TestConnect(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		dbURL, _ := GetDBLocation(t)
		c := NewConnector(WithURL(dbURL))
		err := c.Connect("libsql")
		if err != nil {
			t.Fatalf("unexpected error connecting: %v", err)
		}
		if c.DB == nil {
			t.Fatal("expected non-nil DB after connection")
		}
		c.Close()
	})

	t.Run("invalid driver", func(t *testing.T) {
		dbURL, _ := GetDBLocation(t)
		c := NewConnector(WithURL(dbURL))
		err := c.Connect("invalid_driver")
		if err == nil {
			t.Fatal("expected error with invalid driver")
			c.Close()
		}
	})

	t.Run("double connection attempt", func(t *testing.T) {
		dbURL, _ := GetDBLocation(t)
		c := NewConnector(WithURL(dbURL))
		err := c.Connect("libsql")
		if err != nil {
			t.Fatalf("unexpected error on first connect: %v", err)
		}

		// Intentar conectar de nuevo debería fallar
		err = c.Connect("libsql")
		if err == nil {
			t.Fatal("expected error on second connect")
		}
		c.Close()
	})

	t.Run("malformed url", func(t *testing.T) {
		c := NewConnector(WithURL("invalid://url"))
		err := c.Connect("libsql")
		if err == nil {
			t.Fatal("expected error with invalid URL")
		}
	})
}

// TestConnectorClose verifica el comportamiento del cierre de conexiones
func TestConnectorClose(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		dbURL, _ := GetDBLocation(t)
		c := NewConnector(WithURL(dbURL))
		err := c.Connect("libsql")
		if err != nil {
			t.Fatalf("unexpected error connecting: %v", err)
		}

		err = c.Close()
		if err != nil {
			t.Fatalf("unexpected error closing: %v", err)
		}
	})

	t.Run("close without connect panics", func(t *testing.T) {
		dbURL, _ := GetDBLocation(t)
		c := NewConnector(WithURL(dbURL))
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic when closing without connection")
			}
		}()
		c.Close()
	})

	t.Run("double close", func(t *testing.T) {
		dbURL, _ := GetDBLocation(t)
		c := NewConnector(WithURL(dbURL))
		err := c.Connect("libsql")
		if err != nil {
			t.Fatalf("unexpected error connecting: %v", err)
		}

		err = c.Close()
		if err != nil {
			t.Fatalf("unexpected error on first close: %v", err)
		}

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic on second close")
			}
		}()
		c.Close()
	})
}

// TestConnectorIntegration verifica el funcionamiento completo del connector
// realizando operaciones reales en la base de datos
func TestConnectorIntegration(t *testing.T) {
	dbURL, _ := GetDBLocation(t)
	c := NewConnector(WithURL(dbURL))

	err := c.Connect("libsql")
	if err != nil {
		t.Fatalf("unexpected error connecting: %v", err)
	}

	// Configuramos el modo WAL para mejor rendimiento
	var journalMode string
	err = c.DB.QueryRow("PRAGMA journal_mode=WAL").Scan(&journalMode)
	if err != nil {
		t.Fatalf("failed to set WAL mode: %v", err)
	}
	// El valor retornado debería ser "wal"
	if journalMode != "wal" {
		t.Fatalf("expected journal_mode to be 'wal', got '%s'", journalMode)
	}

	// Creamos una tabla de prueba
	_, err = c.DB.Exec(`
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
		_, err = c.DB.Exec("INSERT INTO test (name) VALUES (?)", name)
		if err != nil {
			t.Fatalf("failed to insert test data '%s': %v", name, err)
		}
	}

	// Verificamos que podemos leer todos los datos insertados
	rows, err := c.DB.Query("SELECT name FROM test ORDER BY id")
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

	// Cerramos la conexión
	err = c.Close()
	if err != nil {
		t.Fatalf("unexpected error closing: %v", err)
	}
}
