package utils

import "golang.org/x/crypto/bcrypt"

// salt 加密盐
const salt = 10

// Encrypt 加密密码
func Encrypt(password string) string {
	cipher, _ := bcrypt.GenerateFromPassword([]byte(password), salt)
	return string(cipher)
}

// Compare 密码比对
func Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
