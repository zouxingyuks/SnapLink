package dao

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/model"
	"SnapLink/pkg/db"
	"context"
	"github.com/pkg/errors"
	cacheBase "github.com/zhufuyi/sponge/pkg/cache"
	"github.com/zhufuyi/sponge/pkg/logger"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"time"
)

// RedirectsCache cache interface
type RedirectsCache interface {
	Set(ctx context.Context, uri string, info *model.Redirect, duration time.Duration) error
	Get(ctx context.Context, uri string) (*model.Redirect, error)
	SetCacheWithNotFound(ctx context.Context, uri string) error
}

type RedirectsDao struct {
	db    *gorm.DB
	cache RedirectsCache
	sfg   *singleflight.Group
}

// NewRedirectsDao creating the dao interface
func NewRedirectsDao(xCache RedirectsCache) *RedirectsDao {
	d := &RedirectsDao{
		db:    db.DB(),
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
	return d
}

// GetByURI 根据 uri 获取对应的原始链接
func (d *RedirectsDao) GetByURI(ctx context.Context, uri string) (*model.Redirect, error) {
	// 使用布隆过滤器进行数据存在性判断
	// 如果不存在，则直接返回
	exist, err := cache.BFExists(ctx, "uri", uri)
	if !exist {
		return nil, model.ErrRecordNotFound
	}
	// 查询缓存，是否查到数据
	record, err := d.cache.Get(ctx, uri)
	if err == nil {
		if record.OriginalURL == "" {
			return nil, model.ErrRecordNotFound
		}
		return record, nil
	}
	if errors.Is(err, model.ErrCacheNotFound) {
		// 为同一 uri，防止同时对 mysql 进行高并发访问
		val, err, _ := d.sfg.Do(uri, func() (interface{}, error) { //nolint

			// 二次查询缓存，是否查到数据
			record, err = d.cache.Get(ctx, uri)
			if err == nil {
				return record, nil
			}
			record = &model.Redirect{
				Uri:         uri,
				Gid:         "",
				OriginalURL: "",
			}
			// 使用 uri 进行查询,专门对 uri 做了索引优化
			err = d.db.WithContext(ctx).Table(record.TName()).Where("uri = ?", uri).First(record).Error
			if err != nil {
				// 设置空值来防御缓存穿透
				if errors.Is(err, model.ErrRecordNotFound) {
					err = d.cache.SetCacheWithNotFound(ctx, uri)
					if err != nil {
						return nil, err
					}
					return nil, model.ErrRecordNotFound
				}
				return nil, err
			}
			//  异步调用设置缓存
			go func() {
				//todo 设置缓存
				//
				//err = d.cache.Set(ctx, uri, record, cache.RedirectsExpireTime)
				//todo 设置缓存失败，记录日志
				logger.Err(errors.Wrap(err, "设置缓存失败"))
			}()
			return record, nil
		})
		if err != nil {
			return nil, err
		}
		info, ok := val.(*model.Redirect)
		if !ok {
			return nil, model.ErrRecordNotFound
		}
		return info, nil
	} else if errors.Is(err, cacheBase.ErrPlaceholder) {
		return nil, model.ErrRecordNotFound
	}

	// 快速失败，如果是其他错误，直接返回
	return nil, err
}

// CleanUp 缓存清理
// 1. 定时清理过期缓存
// 2. 定时将永久缓存转换为短期缓存，然后重新进行缓存预热
func (d *RedirectsDao) CleanUp(ctx context.Context) {
	//todo 定时清理非热点信息
}
