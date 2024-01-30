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
)

var _ RedirectsDao = (*redirectsDao)(nil)

// RedirectsDao defining the dao interface
type RedirectsDao interface {
	GetByURI(ctx context.Context, uri string) (*cache.RedirectInfo, error)
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
