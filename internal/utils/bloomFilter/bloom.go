package bloomFilter

import (
	"bufio"
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/pkg/errors"
	"os"
	"sync"
	"time"
)

var lock = new(sync.Mutex)
var instance = new(struct {
	sync.Once
	*bloom.BloomFilter
})

// Instance 布隆过滤器
// 1. 读取持久化文件，如果不存在则创建
// 2. 创建布隆过滤器
// 3. 定时存储布隆过滤器
func Instance() *bloom.BloomFilter {
	instance.Do(func() {
		var err error
		//1. 读取持久化文件，如果不存在则创建
		//2. 创建布隆过滤器
		instance.BloomFilter, err = readBloomFilter("bloomFilter")
		if err != nil {
			if !os.IsNotExist(err) {
				instance.BloomFilter = bloom.NewWithEstimates(1000000, 0.01)
				writeBloomFilter("bloomFilter", instance.BloomFilter)
			}
			instance.BloomFilter = bloom.NewWithEstimates(1000000, 0.01)
		}
		//3. 定时存储布隆过滤器
		go timerSaveBloomFilter()
	})
	return instance.BloomFilter

}

// todo 超时解锁,思考如何解决

// 定时存储布隆过滤器
func timerSaveBloomFilter() {

	clickTimer := time.NewTimer(time.Minute * 10)
	for {
		select {
		case <-clickTimer.C:
			lock.Lock()
			fmt.Println("定时存储布隆过滤器...")
			err := writeBloomFilter("bloomFilter", instance.BloomFilter)
			if err != nil {
				panic(errors.Wrap(err, "write bloomFilter...failed").Error())
			}
			lock.Unlock()
			clickTimer.Reset(time.Second * 10)
		}
	}

}

// writeBloomFilter 写入布隆过滤器
func writeBloomFilter(filename string, filter *bloom.BloomFilter) error {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		file, err = os.Create(filename)
		if err != nil {
			return err
		}
		//todo 此处应用警告级别日志
	}
	defer file.Close()

	bw := bufio.NewWriter(file)
	_, err = filter.WriteTo(bw)
	if err != nil {
		return err
	}

	return bw.Flush()
}

// readBloomFilter 读取布隆过滤器
func readBloomFilter(filename string) (*bloom.BloomFilter, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	br := bufio.NewReader(file)
	filter := bloom.NewWithEstimates(1000000, 0.01) // 使用适当的n和p值初始化
	_, err = filter.ReadFrom(br)
	if err != nil {
		return nil, err
	}

	return filter, nil
}
