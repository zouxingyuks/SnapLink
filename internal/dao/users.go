package dao

import (
	"SnapLink/internal/bloomFilter"
	"SnapLink/internal/cache"
	"SnapLink/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/zhufuyi/sponge/pkg/ggorm/query"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var _ TUserDao = (*tUserDao)(nil)

// TUserDao defining the dao interface
type TUserDao interface {
	GetDB() *gorm.DB
	Create(ctx context.Context, table *model.TUser) error
	Update(ctx context.Context, table *model.TUser) error
	GetByUsername(ctx context.Context, username string) (*model.TUser, error)
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.TUser, error)
	GetByConditionWithUsername(ctx context.Context, condition *query.Conditions, username string) (*model.TUser, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.TUser, int64, error)
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

func (d *tUserDao) GetDB() *gorm.DB {
	return d.db
}

// Create 创建用户记录
func (d *tUserDao) Create(ctx context.Context, u *model.TUser) error {
	err := d.db.Table(u.TName()).WithContext(ctx).Create(u).Error
	return err
}

// Update 根据用户名更新用户信息
func (d *tUserDao) Update(ctx context.Context, table *model.TUser) error {
	err := d.updateData(ctx, d.db, table)
	// delete slCache
	// todo 解耦为异步删除
	//_ = d.deleteCache(ctx, table.ID)
	return err
}

func (d *tUserDao) updateData(ctx context.Context, db *gorm.DB, table *model.TUser) error {
	if table.Username == "" {
		return errors.New("username cannot be empty")
	}
	update := map[string]interface{}{}
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
	return db.Table(table.TName()).WithContext(ctx).Where("username = ?", table.Username).Updates(update).Error
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
	for i := 0; i < model.TUserShardingNum; i++ {
		err = d.db.WithContext(ctx).Table(fmt.Sprintf("%s-%d", model.TUserPrefix, i)).Where(queryStr, args...).First(table).Error
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
	if err != nil {
		//如果是没有找到记录,此处需要对错误进行转换,以供上层判断
		if errors.Is(err, sql.ErrNoRows) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	err = d.db.Table(user.TName()).WithContext(ctx).Where("id = ?", user.ID).First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// HasUsername 查询用户名是否存在
func (d *tUserDao) HasUsername(ctx context.Context, username string) (bool, error) {

	//1. 在布隆过滤器中查询
	result, err := bloomFilter.BFExists(ctx, "username", username)
	if err != nil {
		return true, err
	}
	// 布隆过滤器认为不存在，就是真不存在
	if !result {
		return false, nil
	}
	//2. 在数据库中查询
	u, err := d.GetByUsername(ctx, username)
	if err != nil {
		//如果err 是 record not found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
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
	for i := 0; i < model.TUserShardingNum; i++ {
		tUsernames := make([]string, 0)
		err := d.db.WithContext(ctx).Table(fmt.Sprintf("%s-%d", model.TUserPrefix, i)).Model(&model.TUser{}).Pluck("username", &tUsernames).Error
		if err != nil {
			return nil, err
		}
		usernames = append(usernames, tUsernames...)
	}
	return usernames, nil
}
