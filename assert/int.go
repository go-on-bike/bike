package assert

import "fmt"

func IntBetween(value int, less int, more int) {
	condition := value >= less && value <= more
	errMsg := fmt.Sprintf("Value must be between %d and %d, got %d", less, more, value)
	assert(condition, errMsg)
}

func IntNot(value int, not int) {
	condition := value != not
	errMsg := fmt.Sprintf("Value must be different from %d, got %d", value, not)
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
