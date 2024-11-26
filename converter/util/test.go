package util

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
)

func AssertEqualMaps[K comparable, V interface{}](t *testing.T, expected, actual map[K]V) bool {
	if expected == nil || actual == nil {
		return assert.Fail(t, fmt.Sprintf("Nil arguments! Expected: %#v, Actual: %#v",
			expected, actual))
	}

	if len(expected) != len(actual) {
		return assert.Fail(t, fmt.Sprintf("Not equal (different lengths): \n"+
			"expected: %d\n"+
			"actual  : %d", len(expected), len(actual),
		))
	}

	for aKey, aValue := range actual {
		eValue, exists := expected[aKey]
		if !exists {
			return assert.Fail(t, fmt.Sprintf("Not equal: \n"+
				"expected: %v\n"+
				"actual  : %v%s", expected, actual, diff(expected, actual),
			))
		}

		if isList(aValue) {
			assert.ElementsMatch(t, eValue, aValue)
		} else {
			assert.Equal(t, eValue, aValue)
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
func diff[K comparable, V interface{}](expected, actual map[K]V) string {
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
