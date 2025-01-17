package logformatter

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/go-on-bike/bike/tester"
	_ "github.com/tursodatabase/go-libsql"
)

func TestLogFormatter_Integration(t *testing.T) {
	stderr := &bytes.Buffer{}

	formatter, errChan := NewLogFormatter(stderr, nil, false, 0)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Run two operations: one in a different go routine
	go func() {
		err := formatter.Start(ctx)
		if err != nil {
			t.Errorf("start failed: %v", err)
		}
	}()

    // a simulation of a normal interacion of libsql handler connecter.

    // formatter entra como stderr al connecter y connecter escribira ahi.
	connecter, _ := tester.NewTestConnector(t, formatter)
	errChan <- connecter.Connect("libsql")
	errChan <- connecter.Connect("libsql")
	connecter.DB()
	connecter.IsConnected()
	errChan <- connecter.Close()
	connecter.IsConnected()
	time.Sleep(time.Millisecond * 100)

	logs := stderr.String()
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
