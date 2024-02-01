package dao

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/model"
	"SnapLink/pkg/db"
	"context"
	"errors"
	"fmt"
	cacheBase "github.com/zhufuyi/sponge/pkg/cache"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"slices"
)

var _ RedirectsDao = (*redirectsDao)(nil)

// RedirectsDao defining the dao interface
type RedirectsDao interface {
	GetByURI(ctx context.Context, uri string) (*cache.RedirectInfo, error)
	WarmUp(ctx context.Context)
	CleanUp(ctx context.Context)
}

type redirectsDao struct {
	db    *gorm.DB
	cache cache.RedirectsCache
	sfg   *singleflight.Group
}

// NewRedirectsDao creating the dao interface
func NewRedirectsDao(xCache cache.RedirectsCache) RedirectsDao {
	return &redirectsDao{
		db:    db.DB(),
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

// GetByURI 根据 uri 获取对应的原始链接
func (d *redirectsDao) GetByURI(ctx context.Context, uri string) (*cache.RedirectInfo, error) {
	record, err := d.cache.Get(ctx, uri)
	if err == nil {
		return record, nil
	}
	if errors.Is(err, model.ErrCacheNotFound) {
		// 为同一 uri，防止同时对 mysql 进行高并发访问
		val, err, _ := d.sfg.Do(uri, func() (interface{}, error) { //nolint
			redirect := &model.Redirect{}
			err = d.db.WithContext(ctx).Where("uri LIKE ?", uri).First(redirect).Error
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
			// 在对应的分表中查询对应的 uri
			shortLink := &model.ShortLink{
				Gid: redirect.Gid,
				Uri: redirect.Uri,
			}
			err = d.db.Table(shortLink.TName()).Where("gid = ? AND uri = ?", shortLink.Gid, shortLink.Uri).First(shortLink).Error
			if err != nil {
				//此处的错误属于异常错误，不应该出现
				return nil, err
			}
			// 根据 gid 获取对应的分表
			info := &cache.RedirectInfo{
				OriginalURL: shortLink.OriginUrl,
				Gid:         shortLink.Gid,
				Uri:         shortLink.Uri,
				Clicks:      shortLink.Clicks,
			}
			// set cache
			err = d.cache.Set(ctx, uri, info, cache.RedirectsExpireTime)
			if err != nil {
				return nil, fmt.Errorf("cache.Set error: %v, uri=%s", err, uri)
			}
			return info, nil
		})
		if err != nil {
			return nil, err
		}
		info, ok := val.(*cache.RedirectInfo)
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

// WarmUp 缓存预热
// 将 lua 脚本存于数据库中，以此实现动态修改缓存预热方案
func (d *redirectsDao) WarmUp(ctx context.Context) {
	// 获取访问量最多的 100 条记录设置为永久缓存
	var records []cache.RedirectInfo
	for i := 0; i < 25; i++ {
		// 获取访问量最多的 100 条记录
		var rows []cache.RedirectInfo
		d.db.WithContext(ctx).Raw(fmt.Sprintf(`SELECT origin_url , gid , uri, clicks FROM short_link_%d WHERE enable = 1 ORDER BY clicks DESC LIMIT 100`, i)).Scan(&rows)
		records = append(records, rows...)
	}
	//todo 配置化处理
	slices.SortFunc(records, func(a, b cache.RedirectInfo) int {
		if a.Clicks > b.Clicks {
			return -1
		} else if a.Clicks == b.Clicks {
			return 0
		} else {
			return 1
		}
	})
	l := len(records)
	// 将这些记录设置为永久缓存
	for i := 0; i < 100 && i < l; i++ {
		info := &cache.RedirectInfo{
			OriginalURL: records[i].OriginalURL,
			Gid:         records[i].Gid,
			Uri:         records[i].Uri,
			Clicks:      records[i].Clicks,
		}
		// set cache
		err := d.cache.Set(ctx, records[i].Uri, info, cache.RedirectsNeverExpireTime)
		if err != nil {
			fmt.Printf("cache.Set error: %v, uri=%s", err, records[i].Uri)
		}
	}
}

// CleanUp 缓存清理
// 1. 定时清理过期缓存
// 2. 定时将永久缓存转换为短期缓存，然后重新进行缓存预热
func (d *redirectsDao) CleanUp(ctx context.Context) {

}
