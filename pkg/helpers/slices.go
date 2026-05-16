// Package helpers provides utility functions for common operations
// throughout the dinnerclub application.
package helpers

import (
	"sort"
)

// SortByFrequency sorts a slice of strings by their frequency of occurrence in descending order.
func SortByFrequency(input []string) []string {
	frequencyMap := make(map[string]int)

	// Count the frequency of each string
	for _, s := range input {
		frequencyMap[s]++
	}

	// Create a slice of the keys in the map
	keys := make([]string, 0, len(frequencyMap))
	for k := range frequencyMap {
		keys = append(keys, k)
	}

	// Sort the keys by frequency in descending order
	sort.Slice(keys, func(i, j int) bool {
		return frequencyMap[keys[i]] > frequencyMap[keys[j]]
	})

	// Create a new slice sorted by frequency
	output := make([]string, len(input))
	idx := 0
	for _, k := range keys {
		for i := 0; i < frequencyMap[k]; i++ {
			output[idx] = k
			idx++
		}
	}

	return output
}

// RemoveDuplicates removes duplicate strings from a slice while preserving order.
func RemoveDuplicates(input []string) []string {
	encountered := map[string]bool{}
	output := []string{}

	for _, s := range input {
		if !encountered[s] {
			encountered[s] = true
			output = append(output, s)
		}
	}

	return output
}
