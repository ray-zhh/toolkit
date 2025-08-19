package str

import "strconv"

func ParseInt64(val string) int64 {
	ret, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return ret
}

func ParseFloat(val string) float64 {
	ret, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return ret
}
