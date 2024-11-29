package util

/*
Equally splits 'arr' into 'slicesNr' slices. If the length or 'arr'
cannot be equally divided, the remainder is distributed again, 1 element per slice
starting from the first slice.
*/
func SplitArray[T any](arr []T, slicesNr uint) [][]T {
	if slicesNr == 0 {
		return [][]T{}
	}

	slices := [][]T{}
	sliceSize := len(arr) / int(slicesNr)
	rem := len(arr) % int(slicesNr)

	for i := range int(slicesNr) {
		newSlice := []T{}
		for j := range sliceSize {
			newSlice = append(newSlice, arr[i*sliceSize+j])
		}
		slices = append(slices, newSlice)
	}
	// equally distribute the remaining elements between
	// the first 'rem' slices
	remStartIndex := int(slicesNr) * sliceSize
	for i := range rem {
		slices[i] = append(slices[i], arr[remStartIndex+i])
	}

	return slices
}

func ArePermutations[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	ma, mb := getMapFromArray(a), getMapFromArray(b)
	return MapsAreEqual(ma, mb)
}

func getMapFromArray[T comparable](arr []T) map[T]int {
	m := make(map[T]int)
	for _, el := range arr {
		if _, ok := m[el]; !ok {
			m[el] = 1
			continue
		}
		m[el] = m[el] + 1
	}

	return m
}
