package reconciler

import "sort"

func MergedSlice(a, b []string) []string {
	check := make(map[string]int)
	res := make([]string, 0)

	d := append(a, b...)
	for _, val := range d {
		check[val] = 1
	}

	for elem := range check {
		res = append(res, elem)
	}

	sort.Strings(res)

	return res
}
