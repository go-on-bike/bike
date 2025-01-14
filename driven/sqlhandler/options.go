package sqlhandler

type connOpts struct {
	url *string
}

type ConnOption func(options *connOpts)

// WithURL establece la URL de conexión a la base de datos.
// Panics si la URL está vacía, ya que esto representa un error de configuración.
func WithURL(url string) ConnOption {
	return func(options *connOpts) {
		if url == "" {
			panic("database URL cannot be empty")
		}
		options.url = &url
	}
}
