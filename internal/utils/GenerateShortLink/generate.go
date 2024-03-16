package GenerateShortLink

import (
	"crypto/sha256"
	"encoding/base64"
	"time"
)

var encode = sha256.New()

// GenerateHash 短链接生成算法
// 1. 将长链接转换为短链接的算法是将长链接进行哈希计算，然后再进行base64编码，最后截取前8个字符作为短链接标识。
func GenerateHash(uri string) string {
	encode.Write([]byte(uri + time.Now().String()))
	sha := base64.URLEncoding.EncodeToString(encode.Sum(nil))
	// 截取前8个字符作为短链接标识
	return sha[:8]
}
