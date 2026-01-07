package collection

import "sort"

type int64Slice []int64

func (a int64Slice) Len() int {
	return len(a)
}

func (a int64Slice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a int64Slice) Less(i, j int) bool {
	return a[i] < a[j]
}

func SortInt(a []int64) {
	sort.Sort(int64Slice(a))
}

type uint64Slice []uint64

func (a uint64Slice) Len() int {
	return len(a)
}

func (a uint64Slice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a uint64Slice) Less(i, j int) bool {
	return a[i] < a[j]
}

func SortUInt(a []uint64) {
	sort.Sort(uint64Slice(a))
}
