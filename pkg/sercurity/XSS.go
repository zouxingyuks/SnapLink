package sercurity

import (
	"fmt"
	"html/template"
)

// CleanXSS 清洗字符串以防御XSS攻击
// 返回清洗后的字符串和一个布尔值，如果输入和输出不同，则可能涉及XSS攻击
func CleanXSS(s string) (string, bool) {
	// 使用template.HTMLEscapeString进行转义
	escapedStr := template.HTMLEscapeString(s)

	// 检查原始字符串和转义后的字符串是否不同
	potentialXSS := escapedStr != s

	fmt.Println("escapedStr:", escapedStr)
	fmt.Println("potentialXSS:", potentialXSS)
	return escapedStr, potentialXSS
}
