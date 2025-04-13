package util

// Cast is a function.
func Cast(
	in map[string]any,
) map[string][]map[string]any {
	out := map[string][]map[string]any{}

	for key, value := range in {
		value, ok := value.([]any)
		if !ok {
			continue
		}

		innerOut := make([]map[string]any, 0, len(value))

		for _, valueArray := range value {
			valueArray, okValueArray := valueArray.(map[string]any)
			if !okValueArray {
				continue
			}

			innerMap := make(map[string]any, len(value))

			for innerKey, innerValue := range valueArray {
				innerMap[innerKey] = innerValue
			}

			innerOut = append(innerOut, innerMap)
		}

		out[key] = innerOut
	}

	return out
}
