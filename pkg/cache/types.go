package cache

import "fmt"

// KeyGenerator 键值生成器
type KeyGenerator func(key string) string

func NewKeyGenerator(prefix string) KeyGenerator {
	return func(key string) string {
		return fmt.Sprintf("%s:%s", prefix, key)
	}
}
