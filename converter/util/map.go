package util

import (
	om "github.com/wk8/go-ordered-map/v2"
)

func CollectKeys[K comparable, V interface{}](m om.OrderedMap[K, V]) []K {
	keys := []K{}

	for pair := m.Oldest(); pair != nil; pair = pair.Next() {
		keys = append(keys, pair.Key)
	}

	return keys
}

func MapsAreEqual[K comparable, V comparable](m1 map[K]V, m2 map[K]V) bool {
	if len(m1) != len(m2) {
		return false
	}

	for key1, value1 := range m1 {
		value2, exists := m2[key1]
		if !exists || value1 != value2 {
			return false
		}
	}
	return true
}
