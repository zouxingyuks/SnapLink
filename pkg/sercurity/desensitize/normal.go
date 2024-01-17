// Package desensitize Descript: 敏感信息脱敏
package desensitize

import "strings"

func PhoneNumber(phone string) string {
	if len(phone) < 10 {
		return phone
	}
	return phone[:3] + strings.Repeat("*", 4) + phone[len(phone)-3:]
}
func IDCard(id string) string {
	if len(id) < 8 {
		return id
	}
	return id[:4] + strings.Repeat("*", len(id)-8) + id[len(id)-4:]
}
