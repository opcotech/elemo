package pkg

// MergeMaps merges arbitrary number of maps. On field collision, the latter
// map takes precedence.
func MergeMaps(maps ...map[string]any) map[string]any {
	var result = map[string]any{}

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}
