package schedule

import "github.com/creativeprojects/resticprofile/clog"

type combinationType int

const (
	weekDay combinationType = iota
	month
	day
	hour
	minute
)

type combinationItem struct {
	itemType combinationType
	value    int
}

// generateCombination generates
// all combinations of size 'size'
// in 'elements'. This function
// mainly uses combinationUtil()
func generateCombination(elements []combinationItem, size int) [][]combinationItem {
	data := make([]combinationItem, size)
	permutations := len(elements) ^ (size - 1)
	clog.Errorf("preparing %d permutations", permutations)
	output := make([][]combinationItem, 0, permutations)
	combinationUtil(elements, data, 0, len(elements)-1, 0, size, &output)
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
	index, r int, output *[][]combinationItem) {
	// current combination is ready
	// send it out
	if index == r {
		clog.Errorf("%+v", data)
		*output = append(*output, data)
		return
	}

	// replace index with all possible
	// elements. The condition "end-i+1 >= r-index"
	// makes sure that including one element
	// at index will make a combination with
	// remaining elements at remaining positions
	for i := start; i <= end && end-i+1 >= r-index; i++ {
		data[index] = arr[i]
		combinationUtil(arr, data, i+1, end, index+1, r, output)
	}
}
