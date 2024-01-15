package dao

import (
	"SnapLink/pkg/db"
	"context"
	"errors"
	"fmt"
	"time"

	"SnapLink/internal/cache"
	"SnapLink/internal/model"

	cacheBase "github.com/zhufuyi/sponge/pkg/cache"
	"github.com/zhufuyi/sponge/pkg/mysql/query"
	"github.com/zhufuyi/sponge/pkg/utils"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var _ ShortLinkDao = (*shortLinkDao)(nil)

// ShortLinkDao 定义接口
type ShortLinkDao interface {
	Create(ctx context.Context, table *model.ShortLink) error
	DeleteByID(ctx context.Context, id uint64) error
	DeleteByIDs(ctx context.Context, ids []uint64) error
	UpdateByID(ctx context.Context, table *model.ShortLink) error
	GetByID(ctx context.Context, id uint64) (*model.ShortLink, error)
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.ShortLink, error)
	GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.ShortLink, error)
	GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.ShortLink, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.ShortLink, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.ShortLink) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.ShortLink) error
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
func (d *shortLinkDao) Create(ctx context.Context, table *model.ShortLink) error {
	err := d.db.WithContext(ctx).Create(table).Error
	_ = d.cache.Del(ctx, uint64(table.ID))
	return err
}

// DeleteByID delete a record by id
func (d *shortLinkDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ShortLink{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.cache.Del(ctx, id)

	return nil
}

// DeleteByIDs delete records by batch id
func (d *shortLinkDao) DeleteByIDs(ctx context.Context, ids []uint64) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.ShortLink{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, id := range ids {
		_ = d.cache.Del(ctx, id)
	}

	return nil
}

// UpdateByID update a record by id
func (d *shortLinkDao) UpdateByID(ctx context.Context, table *model.ShortLink) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.cache.Del(ctx, uint64(table.ID))

	return err
}

// updateDataByID 根据id更新数据
func (d *shortLinkDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.ShortLink) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.OriginUrl != "" {
		update["origin_url"] = table.OriginUrl
	}
	if table.Domain != "" {
		update["domain"] = table.Domain
	}
	if table.Gid != "" {
		update["gid"] = table.Gid
	}
	if table.CreateType != "" {
		update["create_type"] = table.CreateType
	}
	if table.ValidTime.IsZero() == false {
		update["valid_time"] = table.ValidTime
	}
	if table.Description != "" {
		update["description"] = table.Description
	}
	if table.Enable != 0 {
		update["enable"] = table.Enable
	}
	if table.Favicon != "" {
		update["favicon"] = table.Favicon
	}
	if table.Uri != "" {
		update["uri"] = table.Uri
	}
	if table.Clicks != 0 {
		update["clicks"] = table.Clicks
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a record by id
func (d *shortLinkDao) GetByID(ctx context.Context, id uint64) (*model.ShortLink, error) {
	record, err := d.cache.Get(ctx, id)
	if err == nil {
		return record, nil
	}

	if errors.Is(err, model.ErrCacheNotFound) {
		// for the same id, prevent high concurrent simultaneous access to mysql
		val, err, _ := d.sfg.Do(utils.Uint64ToStr(id), func() (interface{}, error) { //nolint
			table := &model.ShortLink{}
			err = d.db.WithContext(ctx).Where("id = ?", id).First(table).Error
			if err != nil {
				// if data is empty, set not found cache to prevent cache penetration, default expiration time 10 minutes
				if errors.Is(err, model.ErrRecordNotFound) {
					err = d.cache.SetCacheWithNotFound(ctx, id)
					if err != nil {
						return nil, err
					}
					return nil, model.ErrRecordNotFound
				}
				return nil, err
			}
			// set cache
			err = d.cache.Set(ctx, id, table, cache.ShortLinkExpireTime)
			if err != nil {
				return nil, fmt.Errorf("cache.Set error: %v, id=%d", err, id)
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.ShortLink)
		if !ok {
			return nil, model.ErrRecordNotFound
		}
		return table, nil
	} else if errors.Is(err, cacheBase.ErrPlaceholder) {
		return nil, model.ErrRecordNotFound
	}

	// fail fast, if cache error return, don't request to db
	return nil, err
}

// GetByCondition get a record by condition
// query conditions:
//
//	name: column name
//	exp: expressions, which default is "=",  support =, !=, >, >=, <, <=, like, in
//	value: column value, if exp=in, multiple values are separated by commas
//	logic: logical type, defaults to and when value is null, only &(and), ||(or)
//
// example: find a male aged 20
//
//	condition = &query.Conditions{
//	    Columns: []query.Column{
//		{
//			Name:    "age",
//			Value:   20,
//		},
//		{
//			Name:  "gender",
//			Value: "male",
//		},
//	}
func (d *shortLinkDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.ShortLink, error) {
	queryStr, args, err := c.ConvertToGorm()
	if err != nil {
		return nil, err
	}

	table := &model.ShortLink{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByIDs get records by batch id
func (d *shortLinkDao) GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.ShortLink, error) {
	itemMap, err := d.cache.MultiGet(ctx, ids)
	if err != nil {
		return nil, err
	}

	var missedIDs []uint64
	for _, id := range ids {
		_, ok := itemMap[id]
		if !ok {
			missedIDs = append(missedIDs, id)
			continue
		}
	}

	// get missed data
	if len(missedIDs) > 0 {
		// find the id of an active placeholder, i.e. an id that does not exist in mysql
		var realMissedIDs []uint64
		for _, id := range missedIDs {
			_, err = d.cache.Get(ctx, id)
			if errors.Is(err, cacheBase.ErrPlaceholder) {
				continue
			}
			realMissedIDs = append(realMissedIDs, id)
		}

		if len(realMissedIDs) > 0 {
			var missedData []*model.ShortLink
			err = d.db.WithContext(ctx).Where("id IN (?)", realMissedIDs).Find(&missedData).Error
			if err != nil {
				return nil, err
			}

			if len(missedData) > 0 {
				for _, data := range missedData {
					itemMap[uint64(data.ID)] = data
				}
				err = d.cache.MultiSet(ctx, missedData, cache.ShortLinkExpireTime)
				if err != nil {
					return nil, err
				}
			} else {
				for _, id := range realMissedIDs {
					_ = d.cache.SetCacheWithNotFound(ctx, id)
				}
			}
		}
	}

	return itemMap, nil
}

// GetByLastID get paging records by last id and limit
func (d *shortLinkDao) GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.ShortLink, error) {
	page := query.NewPage(0, limit, sort)

	var records []*model.ShortLink
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Size()).Where("id < ?", lastID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetByColumns get paging records by column information,
// Note: query performance degrades when table rows are very large because of the use of offset.
//
// params includes paging parameters and query parameters
// paging parameters (required):
//
//	page: page number, starting from 0
//	size: lines per page
//	sort: sort fields, default is id backwards, you can add - sign before the field to indicate reverse order, no - sign to indicate ascending order, multiple fields separated by comma
//
// query parameters (not required):
//
//	name: column name
//	exp: expressions, which default is "=",  support =, !=, >, >=, <, <=, like, in
//	value: column value, if exp=in, multiple values are separated by commas
//	logic: logical type, defaults to and when value is null, only &(and), ||(or)
//
// example: search for a male over 20 years of age
//
//	params = &query.Params{
//	    Page: 0,
//	    Size: 20,
//	    Columns: []query.Column{
//		{
//			Name:    "age",
//			Exp: ">",
//			Value:   20,
//		},
//		{
//			Name:  "gender",
//			Value: "male",
//		},
//	}
func (d *shortLinkDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.ShortLink, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions()
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.ShortLink{}).Select([]string{"id"}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	var records []*model.ShortLink
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// CreateByTx create a record in the database using the provided transaction
func (d *shortLinkDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.ShortLink) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return uint64(table.ID), err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *shortLinkDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	update := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	err := tx.WithContext(ctx).Model(&model.ShortLink{}).Where("id = ?", id).Updates(update).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.cache.Del(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *shortLinkDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.ShortLink) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.cache.Del(ctx, uint64(table.ID))

	return err
}
