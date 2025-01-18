package logformatter

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/go-on-bike/bike/interfaces"
)

type logFormatter struct {
	stderr  io.Writer
	logger  interfaces.Logger
	msgChan chan []byte
	errChan chan error
}

const Sig = "logformatter.go"
const defaultBufferSize = 1024

func NewlogFormatter(
	stderr io.Writer,
	logger interfaces.Logger,
	textFormat bool,
	bufferSize int,
) (*logFormatter, chan error) {
	if stderr == nil {
		panic("logformatter.go new: stderr cannot be nil")
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
	lf := &logFormatter{
		stderr:  stderr,
		logger:  logger,
		msgChan: make(chan []byte, bufferSize),
		errChan: make(chan error, bufferSize),
	}

	return lf, lf.errChan
}

// Write ahora simplemente envía los bytes al canal y retorna inmediatamente
func (lf *logFormatter) Write(p []byte) (n int, err error) {
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

func (lf *logFormatter) listenMessage() error {
	for {
		select {
		case err := <-lf.errChan:
			if err == nil {
				continue
			}
			// aca se puede agregar mas tags pero hay que pensar bien de que antes.
			lf.logger.Error(err.Error())
		case msg := <-lf.msgChan:
			if strMsg := processMessage(msg); strMsg != "" {
				lf.logger.Info(strMsg)
			}
		}
	}
}

func (lf *logFormatter) Start(ctx context.Context) error {
	go lf.listenMessage()
	go func() {
		<-ctx.Done()
		lf.Shutdown(ctx)
	}()
	return nil
}

func processMessage(msg []byte) string {
	msgLen := len(msg)
	for msgLen > 0 && msg[msgLen-1] == '\n' {
		msgLen--
	}
	if msgLen > 0 {
		// aca se puede agregar mas tags pero hay que pensar bien de que y como antes.
		return string(msg[:msgLen])
	}
	return ""
}

func (lf *logFormatter) Shutdown(ctx context.Context) error {
	lf.logger.Info(fmt.Sprintf("%s shutdown: closing stderr channel", Sig))

	go func() {
		// Procesamos los mensajes restantes en el canal de mensajes
		for msg := range lf.msgChan {
			msgLen := len(msg)
			for msgLen > 0 && msg[msgLen-1] == '\n' {
				msgLen--
			}
			if msgLen > 0 {
				lf.logger.Info(string(msg[:msgLen]))
			}
		}

		// Procesamos los errores restantes en el canal de errores
		for err := range lf.errChan {
			if err != nil {
				lf.logger.Error(err.Error())
			}
		}
	}()

	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	for {
		if len(lf.msgChan) == 0 && len(lf.errChan) == 0 {
			close(lf.errChan)
			close(lf.errChan)
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(100 * time.Millisecond)
		}
	}
}
