package assertor

import "fmt"

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
