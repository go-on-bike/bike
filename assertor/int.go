package assertor

import "fmt"

func IntBetween(value int, less int, more int, msg string) {
	condition := value >= less && value <= more
	errMsg := fmt.Sprintf("%s: Value must be between %d and %d, got %d", msg, less, more, value)
	assert(condition, errMsg)
}

func IntNot(value int, not int, msg string) {
	condition := value != not
	errMsg := fmt.Sprintf("%s: Value must be different from %d, got %d", msg, value, not)
	assert(condition, errMsg)
}

func IntGreater(value int, limit int, msg string) {
	condition := value > limit
	errMsg := fmt.Sprintf("%s: value %d, limit %d", msg, value, limit)
	assert(condition, errMsg)
}

func IntGeq(value int, limit int, msg string) {
	condition := value >= limit
	errMsg := fmt.Sprintf("%s: value %d, limit %d", msg, value, limit)
	assert(condition, errMsg)
}

func IntEven(value int, msg string) {
	condition := value%2 == 0
	errMsg := fmt.Sprintf("%s: value %d", msg, value)
	assert(condition, errMsg)
}
