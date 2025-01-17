package recoverer

import "fmt"

// TODO: terminar este paquete

func TheresNothing() error {
	message := "This package is useless"

	return fmt.Errorf("%s", message)
}
