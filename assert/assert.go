package assert

import (
	"fmt"
)

func assert(condition bool, errMsg string) {
	if !condition {
		panic(errMsg)
	}
}

// este caso es diferente ya que no puedo evaluar err sin verificar antes que err es diferente de nil
func ErrNil(err error, msg string) {
	condition := err == nil
	if !condition {
		assert(condition, fmt.Sprintf("%s :%v", msg, err))
	}
}

func ErrorNotNil(err error) {
	condition := err != nil
	if !condition {
		panic("Error is nil")
	}
}

func NotNil(target any) {
	condition := target != nil
	errMsg := "Value is nil"
	assert(condition, errMsg)
}

func Nil(value any) {
	condition := value == nil
	errMsg := "Value is not nil"
	assert(condition, errMsg)
}
