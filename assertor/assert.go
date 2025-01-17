package assertor

import (
	"fmt"
)

// este caso es diferente ya que no puedo evaluar err sin verificar antes que err es diferente de nil
func ErrNil(err error, msg string) {
	condition := err == nil
	if !condition {
		assert(condition, fmt.Sprintf("%s :%v", msg, err))
	}
}

func assert(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}
