package interfaces

type Migrator interface {
	Version() (int, error)
    Move(steps int, inverse bool) error
}

type Connector interface {
    Connect(driver string) error 
    Close() error
    IsConnected() bool
}

type DataHandler interface {
	Connector
	Migrator
}

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}
