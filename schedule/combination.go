package schedule

import (
	"github.com/creativeprojects/resticprofile/calendar"
)

type combinationItem struct {
	itemType calendar.TypeValue
	value    int
}

var (
	output [][]combinationItem
)

// generateCombination generates
// all combinations of size 'size'
// in 'elements'. This function
// mainly uses combinationUtil()
func generateCombination(elements []combinationItem, size int) [][]combinationItem {
	data := make([]combinationItem, size)
	output = make([][]combinationItem, 0)
	combinationUtil(elements, data, 0, len(elements)-1, 0, size)
	return output
}

// arr[] ---> Input Array
// data[] ---> Temporary array to store current combination
// start & end ---> Staring and ending indexes in arr[]
// index ---> Current index in data[]
// r ---> Size of a combination to be generated
func combinationUtil(
	arr []combinationItem, data []combinationItem,
	start, end int,
	index, r int) {
	// current combination is ready
	// send it out
	if index == r {
		temp := make([]combinationItem, len(data))
		// deep copy otherwise the slice always points to the same memory location
		for i, v := range data {
			temp[i] = v
		}
		output = append(output, temp)
		return
	}

	// replace index with all possible
	// elements. The condition "end-i+1 >= r-index"
	// makes sure that including one element
	// at index will make a combination with
	// remaining elements at remaining positions
	for i := start; i <= end && end-i+1 >= r-index; i++ {
		data[index] = arr[i]
		combinationUtil(arr, data, i+1, end, index+1, r)
	}
	return
}
