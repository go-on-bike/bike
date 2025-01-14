package assertor

import (
	"fmt"
)

func assert(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}

// este caso es diferente ya que no puedo evaluar err sin verificar antes que err es diferente de nil
func ErrNil(err error, msg string) {
	condition := err == nil
	if !condition {
		assert(condition, fmt.Sprintf("%s :%v", msg, err))
	}
}

func ErrNotNil(err error, msg string) {
	condition := err != nil
	if !condition {
        assert(condition, fmt.Sprintf("%s: Err is nil", msg))
	}
}

func NotNil(ref any, msg string) {
	condition := ref != nil
    assert(condition,fmt.Sprintf("%s: reference is nil", msg))
}

func Nil(ref any, msg string) {
	condition := ref != nil
    assert(condition,fmt.Sprintf("%s: reference is not nil", msg))
}
