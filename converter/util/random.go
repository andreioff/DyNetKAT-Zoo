package util

import (
	"cmp"
	"errors"
	"maps"
	"math/rand"
	"slices"
)

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
		return OrdArr{}, errors.New(
			"No. of random picks is greater than the no. of unique elements in the array.",
		)
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
	noDup := make(map[E]bool)
	for _, el := range arr {
		noDup[el] = true
	}

	noDupArr := slices.Collect(maps.Keys(noDup))
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
