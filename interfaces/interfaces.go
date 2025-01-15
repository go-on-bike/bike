package interfaces

type Migrator interface {
	Version() (int, error)
    Move(steps int, inverse bool) error
}

type Connector interface {
    Connect(driver string) error 
    Close() error
}

type DataHandler interface {
	Connector
	Migrator
}
