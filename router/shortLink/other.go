package shortLink

import (
	"github.com/google/uuid"
	"net/url"
)

// makeFullShortURL 生成完整的短链接
func makeFullShortURL(uri string) string {
	//此处配置从配置文件中获取
	u := url.URL{
		Scheme: "http",
		Host:   "anubis.cafe",
		Path:   uri,
	}
	return u.String()
}

// ToHash  短链接转hash
func ToHash(domain, shortLink string) string {
	// 生成 hash
	// 1. 生成 hash
	// 尝试生成 10 次，直到生成不重复的hash
	uri := GenerateHash(shortLink)
	for i := 1; i <= 10; i++ {
		// 同一域名下的短链接不能重复
		data := []byte(makeFullShortURL(uri))
		if BloomFilter().Test(data) {
			GenerateHash(shortLink)
		}
		// 误判的情况有
		//todo 1. 误判为存在，但是实际不存在。这种情况暂时不考虑

		// 2. 误判为不存在，但是实际存在，这种情况可以基于数据库的唯一索引来解决
		BloomFilter().Add(data)
		break
	}
	return uri
}
func GenerateHash(uri string) string {
	// 对uri进行hash
	// 1. 生成 hash, 采用 base62
	uuid := uuid.New().String()
	return uuid
}

// todo 布隆过滤器的阀值调整任务
func checkBloomFilter() {
	//todo 定期检查布隆过滤器的阀值，如果超过阀值，需要重新初始化布隆过滤器
}
