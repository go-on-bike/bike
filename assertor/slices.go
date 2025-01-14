package assertor

import (
	"fmt"
	"sort"
)

func SliceIsSorted(x any, less func(i, j int) bool, msg string) {
	condition := sort.SliceIsSorted(x, less)
	assert(condition, fmt.Sprintf("%s: slice is not sorted", msg))
}
