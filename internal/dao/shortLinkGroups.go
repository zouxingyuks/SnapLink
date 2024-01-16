package dao

import (
	"context"
	"errors"
	"fmt"
	errors2 "github.com/pkg/errors"
	"time"

	"SnapLink/internal/cache"
	"SnapLink/internal/model"

	cacheBase "github.com/zhufuyi/sponge/pkg/cache"
	"github.com/zhufuyi/sponge/pkg/mysql/query"
	"github.com/zhufuyi/sponge/pkg/utils"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var _ ShortLinkGroupDao = (*shortLinkGroupsDao)(nil)

// ShortLinkGroupDao defining the dao interface
type ShortLinkGroupDao interface {
	Create(ctx context.Context, record *model.ShortLinkGroup) error
	DeleteByID(ctx context.Context, id uint64) error
	DeleteByIDs(ctx context.Context, ids []uint64) error
	UpdateByID(ctx context.Context, record *model.ShortLinkGroup) error
	GetByID(ctx context.Context, id uint64) (*model.ShortLinkGroup, error)
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.ShortLinkGroup, error)
	GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.ShortLinkGroup, error)
	GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.ShortLinkGroup, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.ShortLinkGroup, int64, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.ShortLinkGroup) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.ShortLinkGroup) error
}

type shortLinkGroupsDao struct {
	db    *gorm.DB
	cache cache.ShortLinkGroupCache
	sfg   *singleflight.Group
}

// NewShortLinkGroupDao creating the dao interface
func NewShortLinkGroupDao(db *gorm.DB, xCache cache.ShortLinkGroupCache) ShortLinkGroupDao {
	return &shortLinkGroupsDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

// Create a record, insert the record and the id value is written back to the table
func (d *shortLinkGroupsDao) Create(ctx context.Context, record *model.ShortLinkGroup) error {
	err := d.db.WithContext(ctx).Create(record).Error
	_ = d.cache.Del(ctx, uint64(record.ID))
	return err
}

// GetByColumns 函数通过列信息获取分页记录，
// 注意：当表行非常多时，由于使用了 offset，查询性能会降低。
//
// params 包括分页参数和查询参数
// 分页参数（必须）：
//
//	page: 页码，从0开始
//	size: 每页的行数
//	sort: 排序字段，默认为 id 逆序，可以在字段前加 - 号表示逆序，不加 - 号表示正序，多个字段用逗号分隔
//
// 查询参数（非必须）：
//
//	name: 列名
//	exp: 表达式，默认为 "=", 支持 =, !=, >, >=, <, <=, like, in
//	value: 列值，若 exp=in，则多个值用逗号分隔
//	logic: 逻辑类型，默认为 and 当值为空时，只支持 &(and) 和 ||(or)
//
// 示例：搜索年龄超过20岁的男性
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
func (d *shortLinkGroupsDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.ShortLinkGroup, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions()
	if err != nil {
		return nil, 0, errors2.Wrap(err, "query params error: ")
	}

	var total int64
	var records []*model.ShortLinkGroup
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	total = int64(len(records))
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByID delete a record by id
func (d *shortLinkGroupsDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ShortLinkGroup{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.cache.Del(ctx, id)

	return nil
}

// DeleteByIDs delete records by batch id
func (d *shortLinkGroupsDao) DeleteByIDs(ctx context.Context, ids []uint64) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.ShortLinkGroup{}).Error
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
func (d *shortLinkGroupsDao) UpdateByID(ctx context.Context, table *model.ShortLinkGroup) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.cache.Del(ctx, uint64(table.ID))

	return err
}

func (d *shortLinkGroupsDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.ShortLinkGroup) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.Description != "" {
		update["description"] = table.Description
	}
	if table.Gid > 0 {
		update["gid"] = table.Gid
	}
	if table.Name != "" {
		update["name"] = table.Name
	}
	if table.CUserId != "" {
		update["c_user"] = table.CUserId
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a record by id
func (d *shortLinkGroupsDao) GetByID(ctx context.Context, id uint64) (*model.ShortLinkGroup, error) {
	record, err := d.cache.Get(ctx, id)
	if err == nil {
		return record, nil
	}

	if errors.Is(err, model.ErrCacheNotFound) {
		// for the same id, prevent high concurrent simultaneous access to mysql
		val, err, _ := d.sfg.Do(utils.Uint64ToStr(id), func() (interface{}, error) { //nolint
			table := &model.ShortLinkGroup{}
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
			err = d.cache.Set(ctx, id, table, cache.ShortLinkGroupExpireTime)
			if err != nil {
				return nil, fmt.Errorf("cache.Set error: %v, id=%d", err, id)
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.ShortLinkGroup)
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
func (d *shortLinkGroupsDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.ShortLinkGroup, error) {
	queryStr, args, err := c.ConvertToGorm()
	if err != nil {
		return nil, err
	}

	table := &model.ShortLinkGroup{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByIDs get records by batch id
func (d *shortLinkGroupsDao) GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.ShortLinkGroup, error) {
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
			var missedData []*model.ShortLinkGroup
			err = d.db.WithContext(ctx).Where("id IN (?)", realMissedIDs).Find(&missedData).Error
			if err != nil {
				return nil, err
			}

			if len(missedData) > 0 {
				for _, data := range missedData {
					itemMap[uint64(data.ID)] = data
				}
				err = d.cache.MultiSet(ctx, missedData, cache.ShortLinkGroupExpireTime)
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
func (d *shortLinkGroupsDao) GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.ShortLinkGroup, error) {
	page := query.NewPage(0, limit, sort)

	records := []*model.ShortLinkGroup{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Size()).Where("id < ?", lastID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *shortLinkGroupsDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.ShortLinkGroup) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return uint64(table.ID), err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *shortLinkGroupsDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	update := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	err := tx.WithContext(ctx).Model(&model.ShortLinkGroup{}).Where("id = ?", id).Updates(update).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.cache.Del(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *shortLinkGroupsDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.ShortLinkGroup) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.cache.Del(ctx, uint64(table.ID))

	return err
}
