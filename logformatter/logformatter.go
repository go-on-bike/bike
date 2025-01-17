package logformatter

import (
	"context"
	"io"
	"log/slog"

	"github.com/go-on-bike/bike/interfaces"
)

type LogFormatter struct {
	stderr  io.Writer
	logger  interfaces.Logger
	msgChan chan []byte
	errChan chan error
}

const defaultBufferSize = 1024

func NewLogFormatter(
	stderr io.Writer,
	logger interfaces.Logger,
	textFormat bool,
	bufferSize int,
) (*LogFormatter, chan error) {
	if stderr == nil {
		panic("logformatter: stderr cannot be nil")
	}

	if logger == nil {
		if textFormat {
			logger = slog.New(slog.NewTextHandler(stderr, nil))
		} else {
			logger = slog.New(slog.NewJSONHandler(stderr, nil))
		}
	}

	if bufferSize == 0 {
		bufferSize = defaultBufferSize
	}

	// Creamos un buffer suficientemente grande para evitar bloqueos
	lf := &LogFormatter{
		stderr:  stderr,
		logger:  logger,
		msgChan: make(chan []byte, bufferSize),
		errChan: make(chan error, bufferSize),
	}

	return lf, lf.errChan
}

// Write ahora simplemente envía los bytes al canal y retorna inmediatamente
func (lf *LogFormatter) Write(p []byte) (n int, err error) {
	// Hacemos una copia de los bytes porque p podría ser reusado por el caller
	msg := make([]byte, len(p))
	copy(msg, p)

	// Enviamos al canal de manera no bloqueante
	select {
	case lf.msgChan <- msg:
		return len(p), nil
	default:
		// Si el canal está lleno, escribimos directamente al stderr
		// para evitar pérdida de logs
		return lf.stderr.Write(p)
	}
}

func (lf *LogFormatter) Start(ctx context.Context) error {
	for {
		select {
		case err := <-lf.errChan:
			if err == nil {
				continue
			}
			lf.logger.Error(err.Error())
		case msg := <-lf.msgChan:
			// Eliminamos los saltos de línea al final
			msgLen := len(msg)
			for msgLen > 0 && msg[msgLen-1] == '\n' {
				msgLen--
			}
			// Solo logeamos si quedó algo después de limpiar los saltos de línea
			if msgLen > 0 {
				lf.logger.Info(string(msg[:msgLen]))
			}
		case <-ctx.Done():
			return nil
		}
	}
}
