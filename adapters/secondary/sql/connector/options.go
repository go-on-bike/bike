package connector

type options struct {
	url *string
}

type Option func(options *options)

// WithURL establece la URL de conexión a la base de datos.
// Panics si la URL está vacía, ya que esto representa un error de configuración.
func WithURL(url string) Option {
	return func(options *options) {
		if url == "" {
			panic("database URL cannot be empty")
		}
		options.url = &url
	}
}
