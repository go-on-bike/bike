package assertor

import "fmt"

func Bytes(bytes []byte, msg string) {
	condition := len(bytes) > 0
	assert(condition, fmt.Sprintf("%s: Bytes have lenght 0", msg))
}
