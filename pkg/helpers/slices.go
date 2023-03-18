package helpers

import (
	"sort"
)

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
