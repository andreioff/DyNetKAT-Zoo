package util

import om "github.com/wk8/go-ordered-map/v2"

func CollectKeys[K comparable, V interface{}](m om.OrderedMap[K, V]) []K {
	keys := []K{}

	for pair := m.Oldest(); pair != nil; pair = pair.Next() {
		keys = append(keys, pair.Key)
	}

	return keys
}
