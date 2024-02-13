package go_redis_bloomfilter

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"sync"
	"testing"
)

var instance = struct {
	bloomFilter *BloomFilter
	sync.Once
}{}

func Instance() *BloomFilter {
	instance.Do(func() {
		instance.bloomFilter = NewBloomFilter(redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
		}))
	})
	return instance.bloomFilter

}
func TestBloomFilter_ADD(t *testing.T) {
	fmt.Print("add test")
	fmt.Println(Instance().ADD(context.Background(), "test", "test"))
	fmt.Print("exist test")
	fmt.Println(Instance().EXIST(context.Background(), "test", "test"))
	fmt.Print("madd test")
	fmt.Println(Instance().MADD(context.Background(), "test", "test1", "test2"))
	fmt.Print("mexist test")
	fmt.Println(Instance().MEXIST(context.Background(), "test", "test1", "test2"))
	fmt.Print("reserve test")
	fmt.Println(Instance().RESERVE(context.Background(), "test2222", 0.01, 1000))
}
