package core

import (
	"strconv"
	"strings"
)

func ParseEpisodeRange(s string) ([]int, error) {
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		start, _ := strconv.Atoi(parts[0])
		end, _ := strconv.Atoi(parts[1])

		if start == 0 {
			start = 1
		}
		if end == 0 {
			end = 1
		}
		if start > end {
			start, end = end, start
		}

		var eps []int
		for i := start; i <= end; i++ {
			eps = append(eps, i)
		}
		return eps, nil
	}

	ep, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	if ep == 0 {
		ep = 1
	}
	return []int{ep}, nil
}
