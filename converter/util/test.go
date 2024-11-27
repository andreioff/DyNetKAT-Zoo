package util

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
	om "github.com/wk8/go-ordered-map/v2"
)

func AssertEqualMaps[K comparable, V interface{}](
	t *testing.T,
	expected, actual *om.OrderedMap[K, V],
) bool {
	if expected == nil || actual == nil {
		return assert.Fail(
			t,
			fmt.Sprintf("Nil arguments! Expected: %#v, Actual: %#v", expected, actual),
		)
	}

	if expected.Len() != actual.Len() {
		return assert.Fail(t, fmt.Sprintf("Not equal (different lengths): \n"+
			"expected: %d\n"+
			"actual  : %d", expected.Len(), actual.Len(),
		))
	}

	for aPair := actual.Oldest(); aPair != nil; aPair = aPair.Next() {
		eValue, exists := expected.Get(aPair.Key)
		if !exists {
			return assert.Fail(t, fmt.Sprintf("Not equal: \n"+
				"expected: %v\n"+
				"actual  : %v%s", expected, actual, diff(*expected, *actual),
			))
		}

		if isList(aPair.Value) {
			assert.ElementsMatch(t, eValue, aPair.Value)
		} else {
			assert.Equal(t, eValue, aPair.Value)
		}
	}

	return true
}

// Checks that the provided value is array or slice.
func isList(list interface{}) bool {
	kind := reflect.TypeOf(list).Kind()
	if kind != reflect.Array && kind != reflect.Slice {
		return false
	}
	return true
}

// copied and adapted from github.com/stretchr/testify/assert
func diff[K comparable, V interface{}](expected, actual om.OrderedMap[K, V]) string {
	e, a := fmt.Sprintf("%v", expected), fmt.Sprintf("%v", actual)

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(e),
		B:        difflib.SplitLines(a),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Actual",
		ToDate:   "",
		Context:  1,
	})

	return "\n\nDiff:\n" + diff
}

func GetOrderedMapPairFunc[K comparable, V interface{}]() func(key K, value V) om.Pair[K, V] {
	return func(key K, value V) om.Pair[K, V] {
		return om.Pair[K, V]{
			Key:   key,
			Value: value,
		}
	}
}

func GetOrderedMapFunc[K comparable, V interface{}]() func(data []om.Pair[K, V]) *om.OrderedMap[K, V] {
	return func(data []om.Pair[K, V]) *om.OrderedMap[K, V] {
		return om.New[K, V](om.WithInitialData(data...))
	}
}
