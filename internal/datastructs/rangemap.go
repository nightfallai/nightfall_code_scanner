package datastructs

import (
	"errors"
	"sort"
)

type rangeValue struct {
	left  int
	right int
	value int
}

type RangeMap struct {
	rangeValues []*rangeValue
}

func (r *RangeMap) Find(key int) (exists bool, value int, canBeInsertedAt int) {
	//use binary sort for searching in the range
	i := sort.Search(len(r.rangeValues), func(i int) bool {
		return key <= r.rangeValues[i].right
	})
	if i < len(r.rangeValues) && key >= r.rangeValues[i].left && key <= r.rangeValues[i].right {
		return true, r.rangeValues[i].value, i
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
	if toBeInsertedAt >= len(r.rangeValues) {
		r.rangeValues = append(r.rangeValues, &rangeValue{left: left, right: right, value: value})
		return nil
	}

	// this will insert range value at toBeInsertedAt
	r.rangeValues = append(r.rangeValues[:toBeInsertedAt+1], r.rangeValues[toBeInsertedAt:]...)
	r.rangeValues[toBeInsertedAt] = &rangeValue{left: left, right: right, value: value}
	return nil
}

func NewRangeMap() *RangeMap {
	return &RangeMap{
		rangeValues: make([]*rangeValue, 0),
	}
}
