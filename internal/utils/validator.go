package utils

import "regexp"

const (
	phoneReg       = `^1[3456789]\d{9}$`
	emailReg       = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	invalidCharReg = `[^\w]`
)

// IsPhone 手机号格式校验
func IsPhone(phone string) bool {
	if !regexp.MustCompile(phoneReg).MatchString(phone) {
		return false
	}
	return true
}

// IsEmail 邮箱格式校验
func IsEmail(email string) bool {
	if !regexp.MustCompile(emailReg).MatchString(email) {
		return false
	}
	return true
}

// LengthCheck 字符串长度校验
func LengthCheck(str string, min, max int) bool {
	l := len(str)
	if l < min || l > max {
		return false
	}
	return true
}

// InvalidCharCheck 字符串非法字符校验
func InvalidCharCheck(str string) bool {
	if regexp.MustCompile(invalidCharReg).MatchString(str) {
		return false
	}
	return true
}
