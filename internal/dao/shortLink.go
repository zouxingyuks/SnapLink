package dao

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/model"
	"SnapLink/pkg/db"
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var _ ShortLinkDao = (*shortLinkDao)(nil)

// ShortLinkDao 定义接口
type ShortLinkDao interface {
	Create(ctx context.Context, table *model.ShortLink) error
	CreateBatch(ctx context.Context, tables []*model.ShortLink) (*model.ShortLink, error)
	List(ctx context.Context, gid string, page, pageSize int) (int, []*model.ShortLink, error)
	Delete(ctx context.Context, uri string) error
}

type shortLinkDao struct {
	db    *gorm.DB
	cache cache.ShortLinkCache
	sfg   *singleflight.Group
}

// NewShortLinkDao 创建 shortLinkDao
func NewShortLinkDao(xCache cache.ShortLinkCache) ShortLinkDao {
	return &shortLinkDao{
		db:    db.DB(),
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

// Create 创建一条短链接
// 1. 创建短链接
// 2. 创建重定向
// 3. 理论上来说,此处创建成功即为成功,缓存的更新不在此处进行
func (d *shortLinkDao) Create(ctx context.Context, shortLink *model.ShortLink) error {
	redirect := &model.Redirect{
		Uri:         shortLink.Uri,
		Gid:         shortLink.Gid,
		OriginalURL: shortLink.OriginUrl,
	}
	// 同时创建短链接和重定向
	err := d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(redirect.TName()).WithContext(ctx).Create(redirect).Error; err != nil {
			return err
		}
		return tx.Table(shortLink.TName()).WithContext(ctx).Create(shortLink).Error

	})
	return err
}

// CreateBatch 批量创建短链接
func (d *shortLinkDao) CreateBatch(ctx context.Context, tables []*model.ShortLink) (*model.ShortLink, error) {
	// 事务
	l := len(tables)
	i := 0
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i = 0; i < l; i++ {
			redirect := &model.Redirect{
				Uri:         tables[i].Uri,
				Gid:         tables[i].Gid,
				OriginalURL: tables[i].OriginUrl,
			}
			if err := tx.Table(redirect.TName()).Create(redirect).Error; err != nil {
				return err
			}
			if err := tx.Table(tables[i].TName()).Create(tables[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return tables[i], err

	}
	return nil, nil
}

// List 分页查询
// 对于深分页情况进行优化
// 1. 基于子查询进行优化
func (d *shortLinkDao) List(ctx context.Context, gid string, page, pageSize int) (int, []*model.ShortLink, error) {
	var list []*model.ShortLink
	tableName := (&model.ShortLink{Gid: gid}).TName()
	var ids []uint
	sql := fmt.Sprintf("SELECT * FROM %s WHERE id IN (?)", tableName)
	total := new(int64)
	//使用事务
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Table(tableName).Where("gid = ?", gid).Count(total).Error
		if err != nil {
			return err
		}
		err = tx.Table(tableName).Select("id").
			Where("gid = ?", gid).
			Limit(pageSize).Offset((page - 1) * pageSize).
			Find(&ids).Error
		if err != nil {
			return err
		}
		err = tx.Raw(sql, ids).Scan(&list).Error
		return err
	})
	return int(*total), list, err
}

// todo 封装一个计算完成的方法

// Delete 删除短链接
func (d *shortLinkDao) Delete(ctx context.Context, uri string) error {
	var gid string
	redirectTableName := model.Redirect{
		Uri: uri,
	}.TName()
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := d.db.WithContext(ctx).Table(redirectTableName).Select("gid").Where("uri = ?", uri).Row().Scan(&gid)
		if err != nil {
			return err
		}
		shortLinkTableName := model.ShortLink{Uri: gid}.TName()
		err = d.db.WithContext(ctx).Table(redirectTableName).Where("uri = ?", uri).Delete(&model.Redirect{}).Error
		if err != nil {
			return err
		}
		err = d.db.WithContext(ctx).Table(shortLinkTableName).Where("uri = ?", uri).Delete(&model.ShortLink{}).Error
		return err
	})
	return err
}
