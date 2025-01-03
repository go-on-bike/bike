package assert

import "sort"

func SliceIsSorted(x any, less func(i, j int) bool) {
	condition := sort.SliceIsSorted(x, less)
	errMsg := "slice is not sorted"
	assert(condition, errMsg)
}
