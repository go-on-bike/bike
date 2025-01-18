package logformatter

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/go-on-bike/bike/driven/sqlhandler"
	"github.com/go-on-bike/bike/tester"
	_ "github.com/tursodatabase/go-libsql"
)

func TestLogFormatter_Integration(t *testing.T) {
	stderr := &bytes.Buffer{}

	formatter, errChan := NewlogFormatter(stderr, nil, false, 0)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dbURL, _ := tester.GenTestLibsqlDBPath(t)
	h := sqlhandler.NewDataHandler(formatter, sqlhandler.WithURL(dbURL), sqlhandler.WithDriver("libsql"))

	// Nos aseguramos de que la conexión se cierre al finalizar
	t.Cleanup(func() {
		// Cancelamos el contexto para iniciar el shutdown
		cancel()
		// Esperamos un poco para que el shutdown se complete
		time.Sleep(100 * time.Millisecond)
	})

    errChan <- formatter.Start(ctx)
	errChan <- h.Start(ctx)
	h.QueryDB()
	h.IsConnected()
	cancel()
	h.IsConnected()

	time.Sleep(time.Millisecond * 100)

	logs := stderr.String()
    t.Log(logs)
	if len(logs) == 0 {
		t.Errorf("final stream is empty")
	}
	for i, line := range strings.Split(logs, "\n") {
		// Ignorar líneas vacías
		if line == "" {
			continue
		}

		// Intentar parsear como JSON
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("línea %d no es JSON válido: %v\ncontenido: %s", i+1, err, line)
			continue
		}

		// Verificsterrar campos esperados en el JSON
		if _, ok := logEntry["time"]; !ok {
			t.Errorf("línea %d no tiene campo 'time'", i+1)
		}
		if _, ok := logEntry["level"]; !ok {
			t.Errorf("línea %d no tiene campo 'level'", i+1)
		}
		if _, ok := logEntry["msg"]; !ok {
			t.Errorf("línea %d no tiene campo 'msg'", i+1)
		}
	}
}
