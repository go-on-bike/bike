package sqlhandler

import "fmt"

type options struct {
	url          *string
	path         *string
	driver       *string
	maxOpenConns *int
	maxIdleConns *int
}

type Option func(options *options)

// WithURL establece la URL de conexión a la base de datos.
// Panics si la URL está vacía, ya que esto representa un error de configuración.
func WithURL(url string) Option {
	return func(options *options) {
		if url == "" {
			panic(fmt.Sprintf("%s: database URL cannot be empty", Sig))
		}
		options.url = &url
	}
}

// WithPATH establece el path donde se encuentran las migraciones.
// Panids si path está vacío.
func WithPATH(path string) Option {
	return func(options *options) {
		if path == "" {
			panic(fmt.Sprintf("%s: migration path cannot be empty", Sig))
		}
		options.path = &path
	}
}

// no se si sea taaan necesario tener un mapa si esto se hace una sola vez y la lista no sera tan
// larga. he pensado en convertirlo en un arreglo.
var drivers = map[string]bool{"libsql": true}

// WithPATH establece el driver donde se encuentran las migraciones.
// Panids si driver está vacío.
func WithDriver(driver string) Option {
	return func(options *options) {
		if driver == "" {
			panic(fmt.Sprintf("%s: migration driver cannot be empty", Sig))
		}
		if !drivers[driver] {
			panic(fmt.Sprintf("%s with driver: migration driver %s not implemented", Sig, driver))
		}
		options.driver = &driver
	}
}

func WithMaxOpenConns(maxOpenConns int) Option {
	return func(options *options) {
		if maxOpenConns <= 0 {
			panic(fmt.Sprintf("%s with max open conns: max open connections have to be grater than 0", Sig))
		}
		options.maxOpenConns = &maxOpenConns
	}
}

func WithMaxIdleConns(maxIdleConns int) Option {
	return func(options *options) {
		if maxIdleConns <= 0 {
			panic(fmt.Sprintf("%s with max idle conns: max iddle connections have to be grater than 0", Sig))
		}
		options.maxIdleConns = &maxIdleConns
	}
}
