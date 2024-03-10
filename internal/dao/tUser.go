package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"SnapLink/internal/cache"
	"SnapLink/internal/model"

	cacheBase "github.com/zhufuyi/sponge/pkg/cache"
	"github.com/zhufuyi/sponge/pkg/ggorm/query"
	"github.com/zhufuyi/sponge/pkg/utils"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var tUserShardingNum = 16

var _ TUserDao = (*tUserDao)(nil)

// TUserDao defining the dao interface
type TUserDao interface {
	Create(ctx context.Context, table *model.TUser) error
	DeleteByID(ctx context.Context, id uint64) error
	DeleteByIDs(ctx context.Context, ids []uint64) error
	UpdateByID(ctx context.Context, table *model.TUser) error
	GetByID(ctx context.Context, id uint64) (*model.TUser, error)
	GetByUsername(ctx context.Context, username string) (*model.TUser, error)
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.TUser, error)
	GetByConditionWithUsername(ctx context.Context, condition *query.Conditions, username string) (*model.TUser, error)
	GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.TUser, error)
	GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.TUser, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.TUser, int64, error)
	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.TUser) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.TUser) error
	HasUsername(ctx context.Context, username string) (bool, error)
	GetAllUserName(ctx context.Context) ([]string, error)
}

type tUserDao struct {
	db    *gorm.DB
	cache cache.TUserCache    // if nil, the cache is not used.
	sfg   *singleflight.Group // if cache is nil, the sfg is not used.
}

// NewTUserDao creating the dao interface
func NewTUserDao(db *gorm.DB, xCache cache.TUserCache) TUserDao {
	if xCache == nil {
		return &tUserDao{db: db}
	}
	return &tUserDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *tUserDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a record, insert the record and the id value is written back to the table
func (d *tUserDao) Create(ctx context.Context, table *model.TUser) error {
	err := d.db.WithContext(ctx).Create(table).Error
	_ = d.deleteCache(ctx, table.ID)
	return err
}

// DeleteByID delete a record by id
func (d *tUserDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.TUser{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// DeleteByIDs delete records by batch id
func (d *tUserDao) DeleteByIDs(ctx context.Context, ids []uint64) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.TUser{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, id := range ids {
		_ = d.deleteCache(ctx, id)
	}

	return nil
}

// UpdateByID update a record by id
func (d *tUserDao) UpdateByID(ctx context.Context, table *model.TUser) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *tUserDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.TUser) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.Username != "" {
		update["username"] = table.Username
	}
	if table.Password != "" {
		update["password"] = table.Password
	}
	if table.RealName != "" {
		update["real_name"] = table.RealName
	}
	if table.Phone != "" {
		update["phone"] = table.Phone
	}
	if table.Mail != "" {
		update["mail"] = table.Mail
	}
	if table.DeletionTime != 0 {
		update["deletion_time"] = table.DeletionTime
	}
	if table.CreateTime.IsZero() == false {
		update["create_time"] = table.CreateTime
	}
	if table.UpdateTime.IsZero() == false {
		update["update_time"] = table.UpdateTime
	}
	if table.DelFlag != 0 {
		update["del_flag"] = table.DelFlag
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a record by id
func (d *tUserDao) GetByID(ctx context.Context, id uint64) (*model.TUser, error) {
	// no cache
	if d.cache == nil {
		record := &model.TUser{}
		err := d.db.WithContext(ctx).Where("id = ?", id).First(record).Error
		return record, err
	}

	// get from cache or database
	record, err := d.cache.Get(ctx, id)
	if err == nil {
		return record, nil
	}

	if errors.Is(err, model.ErrCacheNotFound) {
		// for the same id, prevent high concurrent simultaneous access to database
		val, err, _ := d.sfg.Do(utils.Uint64ToStr(id), func() (interface{}, error) { //nolint
			table := &model.TUser{}
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
			err = d.cache.Set(ctx, id, table, cache.TUserExpireTime)
			if err != nil {
				return nil, fmt.Errorf("cache.Set error: %v, id=%d", err, id)
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.TUser)
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
//
// PS: 此接口由于不带有 username 字段，所以会扫描所有的表
func (d *tUserDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.TUser, error) {
	queryStr, args, err := c.ConvertToGorm()
	if err != nil {
		return nil, err
	}
	table := &model.TUser{}
	for i := 0; i < tUserShardingNum; i++ {
		err = d.db.WithContext(ctx).Table(fmt.Sprintf("t_user_%d", i)).Where(queryStr, args...).First(table).Error
		if err == nil {
			return table, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return table, nil
}

// GetByConditionWithUsername 效果同 GetByCondition，但是会带上 username 字段,因此只会扫描一个表
func (d *tUserDao) GetByConditionWithUsername(ctx context.Context, c *query.Conditions, username string) (*model.TUser, error) {
	queryStr, args, err := c.ConvertToGorm()
	if err != nil {
		return nil, err
	}
	table := &model.TUser{
		Username: username,
	}
	err = d.db.Table(table.TName()).WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}
	return table, nil
}

// GetByIDs get records by batch id
func (d *tUserDao) GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.TUser, error) {
	// no cache
	if d.cache == nil {
		var records []*model.TUser
		err := d.db.WithContext(ctx).Where("id IN (?)", ids).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[uint64]*model.TUser)
		for _, record := range records {
			itemMap[record.ID] = record
		}
		return itemMap, nil
	}

	// get form cache or database
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
		// find the id of an active placeholder, i.e. an id that does not exist in database
		var realMissedIDs []uint64
		for _, id := range missedIDs {
			_, err = d.cache.Get(ctx, id)
			if errors.Is(err, cacheBase.ErrPlaceholder) {
				continue
			}
			realMissedIDs = append(realMissedIDs, id)
		}

		if len(realMissedIDs) > 0 {
			var missedData []*model.TUser
			err = d.db.WithContext(ctx).Where("id IN (?)", realMissedIDs).Find(&missedData).Error
			if err != nil {
				return nil, err
			}

			if len(missedData) > 0 {
				for _, data := range missedData {
					itemMap[data.ID] = data
				}
				err = d.cache.MultiSet(ctx, missedData, cache.TUserExpireTime)
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
func (d *tUserDao) GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.TUser, error) {
	page := query.NewPage(0, limit, sort)

	records := []*model.TUser{}
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
//
// 此为扫描所有的表
// todo 实现全表扫描
func (d *tUserDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.TUser, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions()
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.TUser{}).Select([]string{"id"}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.TUser{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// GetByUsername 根据用户名查询用户信息
func (d *tUserDao) GetByUsername(ctx context.Context, username string) (*model.TUser, error) {
	user := model.TUser{
		Username: username,
	}
	// 此处使用索引覆盖扫描
	err := d.db.Table(user.TName()).WithContext(ctx).Where("username = ?", username).Select("id").Row().Scan(&user.ID)
	err = d.db.Table(user.TName()).WithContext(ctx).Where("id = ?", user.ID).First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *tUserDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.TUser) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *tUserDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	update := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	err := tx.WithContext(ctx).Model(&model.TUser{}).Where("id = ?", id).Updates(update).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *tUserDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.TUser) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

// HasUsername 查询用户名是否存在
func (d *tUserDao) HasUsername(ctx context.Context, username string) (bool, error) {

	//1. 在布隆过滤器中查询
	result, err := cache.Exists(ctx, "username", username)
	if err != nil {
		return true, err
	}
	// 布隆过滤器认为不存在，就是真不存在
	if !result {
		return false, nil
	}
	//2. 在数据库中查询
	u, err := d.GetByCondition(ctx, &query.Conditions{
		Columns: []query.Column{
			{
				Name:  "username",
				Value: username,
				Exp:   "=",
			},
		},
	})
	if err != nil {
		return true, err
	}
	if u == nil {
		return false, nil
	}
	return true, nil
}

// GetAllUserName 获取所有的用户名
func (d *tUserDao) GetAllUserName(ctx context.Context) ([]string, error) {
	usernames := make([]string, 0)
	for i := 0; i < tUserShardingNum; i++ {
		tUsernames := make([]string, 0)
		err := d.db.WithContext(ctx).Table(fmt.Sprintf("t_user_%d", i)).Model(&model.TUser{}).Pluck("username", &tUsernames).Error
		if err != nil {
			return nil, err
		}
		usernames = append(usernames, tUsernames...)
	}
	return usernames, nil
}
