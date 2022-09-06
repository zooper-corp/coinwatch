package tools

import "sort"

type sortable struct {
	nums []float64
	idxs []int
}

func (s sortable) Len() int           { return len(s.nums) }
func (s sortable) Less(i, j int) bool { return s.nums[i] < s.nums[j] }
func (s sortable) Swap(i, j int) {
	s.nums[i], s.nums[j] = s.nums[j], s.nums[i]
	s.idxs[i], s.idxs[j] = s.idxs[j], s.idxs[i]
}

func SortAndReturnIndex(nums []float64) []int {
	idxs := make([]int, len(nums))
	for i := range idxs {
		idxs[i] = i
	}
	sort.Sort(sortable{nums, idxs})
	return idxs
}
