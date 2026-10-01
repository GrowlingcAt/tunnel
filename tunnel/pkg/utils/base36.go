package utils

import "strings"

const chars1 = "abcjklmdef01ghrs2inopq345vwxy67tuz89"

func ToBase36(num int64) string {
	result := ""
	for num > 0 {
		result = string(chars1[num%36]) + result
		num /= 36
	}
	return result
}

func ToBase10(str string) int64 {
	var res int64 = 0
	for _, s := range str {
		index := strings.IndexRune(chars1, s)
		res = res*36 + int64(index)
	}
	return res
}
