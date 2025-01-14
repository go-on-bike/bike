package sqlhandler

import (
	"path/filepath"
	"testing"
)

// CreateTestDB es una función auxiliar que nos ayuda a crear una nueva
// base de datos de prueba para cada test
func CreateTestDB(t *testing.T) string {
	// Creamos un archivo único para cada prueba
	dbPath := filepath.Join(t.TempDir(), "test.db")
	return "file:" + dbPath
}

