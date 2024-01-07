package shortLink

// 基于自增长算法来生成短链

import (
	"strings"
)

var chars = strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", "")

func Generate(id int64) string {
	return getString62(encode62(id))
}

func encode62(id int64) []int64 {
	tempE := []int64{}

	for id > 0 {
		tempE = append(tempE, id%62)
		id /= 62
	}
	return tempE
}

func getString62(indexA []int64) string {
	res := ""

	for _, val := range indexA {
		res += chars[val]
	}
	return reverseString(res)
}

// 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}
