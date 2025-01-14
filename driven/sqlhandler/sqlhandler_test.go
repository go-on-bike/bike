package sqlhandler

import "testing"

func TestSQLHandler_Works(t *testing.T) {
	handler := New(config)

	// Verifica que puedo usar ambas capacidades
	err := handler.Connect("libsql")
	require.NoError(t, err)

	err = handler.Migrate()
	require.NoError(t, err)
}
