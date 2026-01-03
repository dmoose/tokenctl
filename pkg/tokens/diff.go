package tokens

import "reflect"

// Diff returns a new map containing only the keys from target that are different from base
func Diff(target, base map[string]interface{}) map[string]interface{} {
	diff := make(map[string]interface{})

	for key, targetVal := range target {
		baseVal, exists := base[key]

		if !exists {
			// New key in target
			diff[key] = targetVal
			continue
		}

		if !reflect.DeepEqual(targetVal, baseVal) {
			// Value changed
			diff[key] = targetVal
		}
	}

	return diff
}
