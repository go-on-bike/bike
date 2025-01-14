package assertor

import (
	"fmt"

	"github.com/go-on-bike/bike/validator"
)

func NotEmptyString(s string, msg string) {
	condition := len(s) != 0
	assert(condition, msg)
}

func StringDigit(value string, msg string) {
	condition := validator.StringDigit(value)
	errMsg := fmt.Sprintf("%s: String %s is not a valid digit", msg, value)
	assert(condition, errMsg)
}

func StringUUID(value string, msg string) {
	condition := validator.StringUUID(value)
	errMsg := fmt.Sprintf("%s: Value must be a valid uuid, got %s", msg, value)
	assert(condition, errMsg)
}

func StringAllowedValues(value string, msg string, allowedValues ...string) {
	condition := validator.AllowedValues(value, allowedValues...)
	errMsg := fmt.Sprintf("%s: Value %s is not in te allowed values %s", msg, value, allowedValues)
	assert(condition, errMsg)
}

func StringArrayMin(strings []string, size int, msg string) {
	arraySize := len(strings)
	condition := arraySize >= size
	errMsg := fmt.Sprintf("%s: String array lenght %d is smaller than min %d", msg, arraySize, size)
	assert(condition, errMsg)
}

func StringArrayMax(strings []string, size int, msg string) {
	arraySize := len(strings)
	condition := arraySize < size
	errMsg := fmt.Sprintf("%s: String array lenght %d bigger than max %d", msg, arraySize, size)
	assert(condition, errMsg)
}
