package dao

import (
	"SnapLink/internal/bloomFilter"
	"SnapLink/internal/cache"
	"SnapLink/internal/custom_err"
	"SnapLink/internal/model"
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	cacheBase "github.com/zhufuyi/sponge/pkg/cache"
	"github.com/zhufuyi/sponge/pkg/logger"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// RedirectsDao defining the dao interface
type RedirectsDao interface {
	GetByURI(ctx context.Context, uri string) (*model.Redirect, error)
	CleanUp(ctx context.Context)
}
type redirectsDao struct {
	db  *gorm.DB
	sfg *singleflight.Group
}

// NewRedirectsDao 创建数据层接口
func NewRedirectsDao(db *gorm.DB, client *redis.Client) (RedirectsDao, error) {
	var err error
	d := &redirectsDao{
		db:  db,
		sfg: new(singleflight.Group),
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

// GetByURI 根据 uri 获取对应的原始链接
func (d *redirectsDao) GetByURI(ctx context.Context, uri string) (*model.Redirect, error) {
	// 使用布隆过滤器进行数据存在性判断
	// 如果不存在，则直接返回
	// 此处使用布隆过滤器的原因是: 减少大量空值造成的缓存内存占用过大
	exist, err := bloomFilter.BFExists(ctx, "uri", uri)
	if !exist {
		return nil, custom_err.ErrRecordNotFound
	}
	// 查询缓存，是否查到数据
	record, err := cache.Redirect().Get(ctx, uri)
	if err == nil {
		if record.OriginalURL == "" {
			return nil, custom_err.ErrRecordNotFound
		}
		return record, nil
	}
	if errors.Is(err, custom_err.ErrCacheNotFound) {
		// 基于 singleflight 进行并发调用合并,主要的性能优化点在于:
		//1. 减少数据库压力：通过合并请求，减少了对数据库的总体访问次数，从而降低了数据库的负载。
		//2. 节省时间：避免了多次加锁解锁的过程，因为对于相同的资源只进行了一次查询，减少了时间消耗。
		//3. 简化缓存策略：由于所有相同的请求都等待同一次查询结果，因此不再需要二次缓存查询，简化了缓存管理。
		val, err, _ := d.sfg.Do(uri, func() (interface{}, error) { //nolint
			record = &model.Redirect{
				Uri:         uri,
				Gid:         "",
				OriginalURL: "",
			}
			// 使用 uri 进行查询,专门对 uri 做了索引优化
			err = d.db.WithContext(ctx).Table(record.TName()).Where("uri = ?", uri).First(record).Error
			if err != nil {
				// 设置空值来防御缓存穿透
				if errors.Is(err, custom_err.ErrRecordNotFound) {
					err = cache.Redirect().SetCacheWithNotFound(ctx, uri)
					if err != nil {
						return nil, err
					}
					return nil, custom_err.ErrRecordNotFound
				}
				return nil, err
			}
			// 设置缓存
			err = cache.Redirect().Set(ctx, uri, record, cache.RedirectsExpireTime)
			if err != nil {
				logger.Err(errors.Wrap(err, "设置缓存失败"))
			}
			return record, nil
		})
		if err != nil {
			return nil, err
		}
		info, ok := val.(*model.Redirect)
		if !ok {
			return nil, custom_err.ErrRecordNotFound
		}
		return info, nil
	} else if errors.Is(err, cacheBase.ErrPlaceholder) {
		return nil, custom_err.ErrRecordNotFound
	}

	// 快速失败，如果是其他错误，直接返回
	return nil, err
}

// CleanUp 缓存清理
// 1. 定时清理过期缓存
// 2. 定时将永久缓存转换为短期缓存，然后重新进行缓存预热
func (d *redirectsDao) CleanUp(ctx context.Context) {
	//todo 定时清理非热点信息
}
