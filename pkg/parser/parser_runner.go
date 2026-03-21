package parser

import (
	"maps"
)

// generates all combinations of variables
func GenerateVariableCombinations(variables map[string]any) []map[string]any {
	var combinations []map[string]any

	// Helper function to recursively generate combinations
	var generate func(map[string]any, map[string]any, []string)
	generate = func(current map[string]any, remaining map[string]any, keys []string) {
		if len(keys) == 0 {
			combo := make(map[string]any)
			maps.Copy(combo, current)
			combinations = append(combinations, combo)
			return
		}

		key := keys[0]
		value := remaining[key]

		switch v := value.(type) {
		case []any:
			for _, item := range v {
				current[key] = item
				generate(current, remaining, keys[1:])
			}
		case string:
			current[key] = v
			generate(current, remaining, keys[1:])
		case map[string]any:
			for subKey, subValue := range v {
				current[key] = map[string]any{subKey: subValue}
				generate(current, remaining, keys[1:])
			}
		default:
			current[key] = v
			generate(current, remaining, keys[1:])
		}
	}

	generate(map[string]any{}, variables, getKeys(variables))
	return combinations
}

func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func isExcluded(configSet map[string]any, excludedSets []map[string]any) bool {
	for _, exclusion := range excludedSets {
		matchCounter := 0
		// verify and count matching exclusion variables
		for k, v := range exclusion {
			if configSet[k] == v {
				matchCounter += 1
			}
		}
		if matchCounter == len(exclusion) { // if all conditions match
			// then exclusion condition is met
			return true
		}
	}
	return false
}
