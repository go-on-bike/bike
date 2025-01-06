package db

type DBOperator interface {
	Connect()
	Close()
    // TODO: cambiar direction por booleano
	RunMigrations(path string, direction bool, steps int) error
}
