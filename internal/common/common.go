package common

import "sort"

func MergeStringMap(a, b map[string]string) map[string]string {
	for k, v := range a {
		b[k] = v
	}

	return b
}

func Unique(a, b []string) []string {
	stringSlice := append(a, b...)
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	if len(list) == 0 {
		return nil
	}

	sort.Strings(list)

	return list
}
