package datastructs_test

import (
	"github.com/nightfallai/nightfall_code_scanner/internal/datastructs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRangeMap(t *testing.T) {
	rangeMap := datastructs.NewRangeMap()
	err := rangeMap.AddRange(0, 100, 20)
	assert.NoError(t, err)
	exists, value, _ := rangeMap.Find(50)
	assert.True(t, exists)
	assert.Equal(t, 20, value)
	assert.NoError(t, err)
	err = rangeMap.AddRange(50, 150, 30)
	assert.Error(t, err)
	err = rangeMap.AddRange(102, 150, 21)
	assert.NoError(t, err)
	err = rangeMap.AddRange(101, 160, 22)
	assert.Error(t, err)
	err = rangeMap.AddRange(500, 750, 23)
	assert.NoError(t, err)
	err = rangeMap.AddRange(250, 300, 27)
	assert.NoError(t, err)
	exists, value, _ = rangeMap.Find(265)
	assert.True(t, exists)
	assert.Equal(t, 27, value)
	exists, _, _ = rangeMap.Find(200)
	assert.False(t, exists)
}
