package dao

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/model"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"sync"
)

const (
	usernameBF = "username"
	uriBF      = "uri"
)

type bfMakeFn = func(db *gorm.DB) (bool, error)

// 使用类工厂模式的思想来批量执行
var bfMakerFns = map[string]bfMakeFn{
	uriBF:      makeBFWithShardingNum(uriBF, model.RedirectShardingNum, model.RedirectPrefix, "uri"),
	usernameBF: makeBFWithShardingNum(usernameBF, model.TUserShardingNum, model.TUserPrefix, "username"),
}

type fixDao struct {
	db   *gorm.DB
	sfg  *singleflight.Group
	once sync.Once
}

// RebulidBF 重建布隆过滤器
func (d *fixDao) RebulidBF() (errs []error) {
	// 后面还是需要进行限流的
	// 重建布隆过滤器,使用 singleflight 来保证只有一个协程进行重建
	val, err, _ := d.sfg.Do("RebulidBF", func() (interface{}, error) {
		// 重建布隆过滤器
		for name, fn := range bfMakerFns {
			// 删除布隆过滤器
			if err := cache.BFCache().BFDelete(context.Background(), name); err != nil {
				errs = append(errs, err)
				return errs, errors.New(fmt.Sprintf("rebuild %s bloom filter failed", name))
			}
			// 重新创建布隆过滤器
			ok, errs := fn(d.db)
			if !ok {
				return errs, errors.New(fmt.Sprintf("rebuild %s bloom filter failed", name))
			}
		}
		return nil, nil
	})
	if err != nil {
		return val.([]error)
	}
	return nil
}

// 针对分表的
func makeBFWithShardingNum(key string, shardingNum int, prefix, param string) bfMakeFn {
	return func(db *gorm.DB) (bool, error) {
		// TODO: 此处改为远程配置
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := cache.BFCache().BFCreate(ctx, key, 0.01, 1e9); err != nil {
			// 记录日志
			return false, err
		}
		errCh := make(chan error, shardingNum)
		wg := sync.WaitGroup{}
		wg.Add(shardingNum)

		for i := 0; i < shardingNum; i++ {
			tableName := fmt.Sprintf("%s-%d", prefix, i)
			go func(id int, tName string) {
				defer wg.Done()
				data, err := getAll(ctx, tableName, db, []string{param})
				if err != nil {
					errCh <- err
					return
				}
				if err := cache.BFCache().BFMAdd(ctx, key, data[param]...); err != nil {
					errCh <- err
					return
				}
				errCh <- nil
			}(i, tableName)
		}
		go func() {
			wg.Wait()
			close(errCh)
		}()
		for err := range errCh {
			if err != nil {
				return false, err
			}
		}
		return true, nil
	}
}
func makeBFWithoutShardingNum(key string, prefix, param string) bfMakeFn {
	return func(db *gorm.DB) (bool, error) {
		// TODO: 此处改为远程配置
		ctx := context.Background()
		if err := cache.BFCache().BFCreate(ctx, key, 0.01, 1e9); err != nil {
			// 记录日志
			return false, err
		}
		tableName := prefix
		data, err := getAll(ctx, tableName, db, []string{param})
		if err != nil {
			return false, err
		}
		if err := cache.BFCache().BFMAdd(ctx, key, data[param]...); err != nil {
			return false, err
		}
		return true, nil
	}
}
