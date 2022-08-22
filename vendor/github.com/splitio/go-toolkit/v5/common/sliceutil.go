package common

// Partition create partitions considering the passed amount
func Partition(items []string, maxItems int) [][]string {
	var splitted [][]string

	for i := 0; i < len(items); i += maxItems {
		end := i + maxItems

		if end > len(items) {
			end = len(items)
		}

		splitted = append(splitted, items[i:end])
	}

	return splitted
}

// DedupeStringSlice returns a new slice without duplicate strings
func DedupeStringSlice(items []string) []string {
	present := make(map[string]struct{}, len(items))
	for idx := range items {
		present[items[idx]] = struct{}{}
	}

	ret := make([]string, 0, len(present))
	for key := range present {
		ret = append(ret, key)
	}

	return ret
}
