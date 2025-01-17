package sqlhandler

import "fmt"

type ConnOption func(options *connOpts)

// WithURL establece la URL de conexión a la base de datos.
// Panics si la URL está vacía, ya que esto representa un error de configuración.
func WithURL(url string) ConnOption {
	return func(options *connOpts) {
		if url == "" {
            panic(fmt.Sprintf("%s: database URL cannot be empty", SigConn))
		}
		options.url = &url
	}
}

type MigrOption func(options *migrOpts)

// WithPATH establece el path donde se encuentran las migraciones.
// Panids si path está vacío.
func WithPATH(path string) MigrOption {
	return func(options *migrOpts) {
		if path == "" {
            panic(fmt.Sprintf("%s: migration path cannot be empty", SigConn))
		}
		options.path = &path
	}
}
