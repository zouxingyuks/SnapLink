package GenerateShortLink

import (
	"crypto/sha256"
	"encoding/base64"
)

var hasher = sha256.New()

// GenerateHash 短链接生成算法
func GenerateHash(uri string) string {
	hasher.Write([]byte(uri))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	// 截取前8个字符作为短链接标识
	return sha[:8]
}
