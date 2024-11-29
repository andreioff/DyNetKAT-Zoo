package util

import (
	"cmp"
	"math/rand"
	"slices"

	om "github.com/wk8/go-ordered-map/v2"
)

// const SEED int64 = 31
const SEED int64 = 3

var randGen rand.Rand

func init() {
	randGen = *rand.New(rand.NewSource(SEED))
}

/*
Picks at random 'picksNr' WITHOUT replacement from 'arr' and returns the result.
Ignores duplicate elements.
It stable sorts the elements of the array to ensure reproducible results.
*/
func RandomFromArray[OrdArr ~[]E, E cmp.Ordered](arr OrdArr, picksNr uint) (OrdArr, error) {
	arr = sortAndRemoveDuplicates(arr)
	if int(picksNr) > len(arr) {
		return OrdArr{}, NewError(ErrMorePicksThanUniqueElements)
	}

	picks := OrdArr{}
	randIndecies := randGen.Perm(len(arr))[:picksNr]

	for _, randIndex := range randIndecies {
		randValue := arr[randIndex]
		picks = append(picks, randValue)
	}

	return picks, nil
}

func sortAndRemoveDuplicates[OrdArr ~[]E, E cmp.Ordered](arr OrdArr) OrdArr {
	noDup := om.New[E, bool]()
	noDupArr := []E{}
	for _, el := range arr {
		_, exists := noDup.Get(el)
		if !exists {
			noDupArr = append(noDupArr, el)
			noDup.Set(el, true)
		}
	}

	slices.Sort(noDupArr)
	return noDupArr
}

/*
Picks at random 'picksNr' elements with replacement from 'arr' and returns the result.
It stable sorts the elements of the array to ensure reproducible results.
*/
func RandomFromArrayWithReplc[OrdArr ~[]E, E cmp.Ordered](arr OrdArr, picksNr uint) OrdArr {
	arr = sortAndRemoveDuplicates(arr)

	picks := OrdArr{}
	arrLen := len(arr)
	for range picksNr {
		randIndex := randGen.Intn(arrLen)
		randValue := arr[randIndex]

		picks = append(picks, randValue)
	}

	return picks
}

func RandomInts(n, minVal, maxVal int) []int {
	if n < 1 || maxVal < 1 || minVal < 0 || minVal >= maxVal {
		return []int{}
	}

	arr := []int{}
	for range n {
		randV := randGen.Intn(maxVal - minVal)
		arr = append(arr, minVal+randV)
	}

	return arr
}
