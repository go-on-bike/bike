package assert

func Bytes(bytes []byte) {
	condition := len(bytes) > 0
	errMsg := "Bytes have lenght 0"
	assert(condition, errMsg)
}
