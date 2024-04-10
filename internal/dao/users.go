package dao

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/zhufuyi/sponge/pkg/ggorm/query"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"sync"
)

type tUserDao struct {
	db   *gorm.DB
	sfg  *singleflight.Group
	once sync.Once
}

// Create 创建用户记录
func (d *tUserDao) Create(ctx context.Context, u *model.TUser) error {
	err := d.db.Table(u.TName()).WithContext(ctx).Create(u).Error
	return err
}

// Update 根据用户名更新用户信息
func (d *tUserDao) Update(ctx context.Context, table *model.TUser) error {
	err := d.updateData(ctx, d.db, table)
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
	result, err := cache.BFCache().BFExists(ctx, "username", username)
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
