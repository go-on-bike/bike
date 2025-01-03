package assert

import (
	"fmt"

	"github.com/matisin/bike/validator"
)

func NotEmptyString(s string, msg string) {
	condition := len(s) != 0
	assert(condition, msg)
}

func StringDigit(value string) {
	condition := validator.StringDigit(value)
	errMsg := fmt.Sprintf("String %s is not a valid digit", value)
	assert(condition, errMsg)
}

func StringUUID(value string) {
	condition := validator.StringUUID(value)
	errMsg := fmt.Sprintf("Value must be a valid uuid, got %s", value)
	assert(condition, errMsg)
}

func StringAllowedValues(value string, allowedValues ...string) {
	condition := validator.AllowedValues(value, allowedValues...)
	errMsg := fmt.Sprintf("Value %s is not in te allowed values %s", value, allowedValues)
	assert(condition, errMsg)
}

func StringArrayMin(strings []string, size int) {
	arraySize := len(strings)
	condition := arraySize >= size
	errMsg := fmt.Sprintf("String array lenght %d is smaller than min %d", arraySize, size)
	assert(condition, errMsg)
}

func StringArrayMax(strings []string, size int) {
	arraySize := len(strings)
	condition := arraySize < size
    errMsg := fmt.Sprintf("String array lenght %d bigger than max %d", arraySize, size) 
	assert(condition, errMsg)
}
