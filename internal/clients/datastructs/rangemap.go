package datastructs

import (
	"errors"
	"sort"
)

type pair struct {
	left  int
	right int
}

type RangeMap struct {
	ranges []pair
	values []int
}

func (r *RangeMap) Find(key int) (exists bool, value int, canBeInsertedAt int) {
	//use binary sort for searching in the range
	i := sort.Search(len(r.ranges), func(i int) bool {
		return key <= r.ranges[i].right
	})
	if i < len(r.ranges) && key >= r.ranges[i].left && key <= r.ranges[i].right {
		return true, r.values[i], i
	}
	return false, 0, i
}

func (r *RangeMap) AddRange(left int, right int, value int) error {
	//check if range does not overlap
	leftExists, _, toBeInsertedAtLeft := r.Find(left)
	if leftExists {
		return errors.New("range should not overlap")
	}
	rightExists, _, toBeInsertedAtRight := r.Find(right)
	if rightExists {
		return errors.New("range should not overlap")
	}

	if toBeInsertedAtLeft != toBeInsertedAtRight {
		return errors.New("range should not overlap")
	}
	toBeInsertedAt := toBeInsertedAtRight
	if toBeInsertedAt >= len(r.ranges) {
		r.ranges = append(r.ranges, pair{left: left, right: right})
		r.values = append(r.values, value)
		return nil
	}

	r.ranges = append(r.ranges[:toBeInsertedAt+1], r.ranges[toBeInsertedAt:]...)
	r.ranges[toBeInsertedAt] = pair{left: left, right: right}
	r.values = append(r.values[:toBeInsertedAt+1], r.values[toBeInsertedAt:]...)
	r.values[toBeInsertedAt] = value
	return nil
}

func NewRangeMap() *RangeMap {
	return &RangeMap{
		ranges: make([]pair, 0),
		values: make([]int, 0),
	}
}
