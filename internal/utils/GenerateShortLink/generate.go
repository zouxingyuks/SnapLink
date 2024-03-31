package GenerateShortLink

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// GenerateHash 短链接生成算法
// 1. 将长链接转换为短链接的算法是将长链接进行哈希计算，然后再进行base64编码，最后截取前8个字符作为短链接标识。
func GenerateHash(url string) (uri string) {
	var encode = sha256.New()
	data := []byte(url + time.Now().String())
	n := rand.Intn(len(data))
	encode.Write(append(data[n+1:], data[:n]...))
	sha := base64.URLEncoding.EncodeToString(encode.Sum(nil))
	// 截取前8个字符作为短链接标识
	return sha[:8]
	////method := rand.Int31n(2)
	////if method == 0 {
	////	return toOtherNumberSystem(time.Now().UnixNano(), 62)[:8]
	////}
	//return shortUrl(url)[0]
}

var (
	chars = []string{
		"a", "b", "c", "d", "e", "f", "g", "h",
		"i", "j", "k", "l", "m", "n", "o", "p",
		"q", "r", "s", "t", "u", "v", "w", "x",
		"6", "7", "8", "9", "A", "B", "C", "D",
		"E", "F", "G", "H", "I", "J", "K", "L",
		"M", "N", "O", "P", "Q", "R", "S", "T",
		"U", "V", "W", "X", "Y", "Z",
		"y", "z", "0", "1", "2", "3", "4", "5",
	}
)

func shortUrl(url string) []string {
	key := "orzW6x5Q"

	// MD5加密
	data := []byte(key + url)
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	cipherStr := md5Ctx.Sum(nil)
	md5String := hex.EncodeToString(cipherStr)

	resUrl := make([]string, 4)

	for i := 0; i < 4; i++ {
		sTempSubString := md5String[i*8 : i*8+8]
		lHexLong, _ := hex.DecodeString(sTempSubString)
		var lHexLongInt64 int64 = int64(lHexLong[0])<<24 + int64(lHexLong[1])<<16 + int64(lHexLong[2])<<8 + int64(lHexLong[3])
		lHexLongInt64 &= 0x3FFFFFFF

		var outChars string
		for j := 0; j < 6; j++ {
			index := 0x0000003D & lHexLongInt64
			outChars += string(chars[index])
			lHexLongInt64 >>= 5
		}
		resUrl[i] = outChars
	}
	return resUrl
}

const digits = "01234fghijklmnopqrstuvwRSTUCDEFopnuvwOPVWHIJKL56XxyzABCDEFG789abcdeMNOPQYZ"

// toOtherNumberSystem 将十进制数字转换为指定进制的字符串
func toOtherNumberSystem(number int64, seed int) string {
	if number < 0 {
		number = (2 * 0x7fffffff) + number + 2
	}

	var buf [64]rune
	charPos := len(buf)

	for number/int64(seed) > 0 {
		charPos--
		buf[charPos] = rune(digits[number%int64(seed)])
		number /= int64(seed)
	}
	charPos--
	buf[charPos] = rune(digits[number%int64(seed)])

	return string(buf[charPos:])
}

// toDecimalNumber 将其他进制的数字（字符串形式）转换为十进制的数字
func toDecimalNumber(number string, seed int) int64 {
	if seed == 10 {
		var result int64
		fmt.Sscanf(number, "%d", &result)
		return result
	}

	var result int64
	base := int64(1)

	for i := len(number) - 1; i >= 0; i-- {
		index := strings.IndexRune(digits, rune(number[i]))
		result += int64(index) * base
		base *= int64(seed)
	}

	return result
}
