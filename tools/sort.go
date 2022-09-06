package tools

import "sort"

type sortable struct {
	sort.Float64Slice
	idx []int
}

func (s sortable) Swap(i, j int) {
	s.Float64Slice.Swap(i, j)
	s.idx[i], s.idx[j] = s.idx[j], s.idx[i]
}

func newSortable(n ...float64) *sortable {
	s := &sortable{Float64Slice: sort.Float64Slice(n), idx: make([]int, len(n))}
	for i := range s.idx {
		s.idx[i] = i
	}
	return s
}

func SortAndReturnIndex(nums []float64) []int {
	s := newSortable(nums...)
	sort.Sort(s)
	return s.idx
}
