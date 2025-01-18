package interfaces

import "context"

type Starter interface {
	Start(ctx context.Context) error
}

type GracefulStarter interface {
    Starter
    Shutdown(ctx context.Context) error
}

type DataHandler interface {
	GracefulStarter
	Version() (int, error)
	RunMigrations(steps int, inverse bool) error
	IsConnected() bool
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

