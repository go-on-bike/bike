package assertor

import "fmt"

func FloatGeq(value float64, limit float64, msg string) {
	condition := value >= limit
	errMsg := fmt.Sprintf("%s: value %f, limit %f", msg, value, limit)
	assert(condition, errMsg)
}

func FloatLeq(value float64, limit float64, msg string) {
	condition := value <= limit
	errMsg := fmt.Sprintf("%s: value %f, limit %f", msg, value, limit)
	assert(condition, errMsg)
}
