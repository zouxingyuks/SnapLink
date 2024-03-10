package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"SnapLink/internal/cache"
	"SnapLink/internal/model"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var _ ShortLinkGroupDao = (*shortLinkGroupsDao)(nil)

// ShortLinkGroupDao defining the dao interface
type ShortLinkGroupDao interface {
	Create(ctx context.Context, record *model.ShortLinkGroup) error
	GetAllByCUser(ctx context.Context, cUser string) ([]*model.ShortLinkGroup, error)
	GetAll(ctx context.Context) ([]*model.ShortLinkGroup, error)
	UpdateByGid(ctx context.Context, gid string, name string) (*model.ShortLinkGroup, error)
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

// Create 创建短链接分组
func (d *shortLinkGroupsDao) Create(ctx context.Context, group *model.ShortLinkGroup) error {
	err := d.db.Table(group.TName()).WithContext(ctx).Create(group).Error
	return err
}

// GetAllByCUser 根据创建人获取所有的分组
func (d *shortLinkGroupsDao) GetAllByCUser(ctx context.Context, cUser string) ([]*model.ShortLinkGroup, error) {
	// 从缓存中获取
	records, _ := d.cache.HGetALL(ctx, cUser)
	if len(records) > 0 {
		return records, nil
	}
	// 从数据库中获取
	tableName := model.ShortLinkGroup{
		CUsername: cUser,
	}.TName()
	err := d.db.Table(tableName).WithContext(ctx).Where("c_username = ?", cUser).Find(&records).Error
	if err != nil {
		return nil, err
	}
	d.cache.HMSet(ctx, cUser, records)
	return records, nil

}

// GetAll 获取所有的分组
func (d *shortLinkGroupsDao) GetAll(ctx context.Context) (result []*model.ShortLinkGroup, err error) {
	for i := 0; i < model.SLGroupShardingNum; i++ {
		var records []*model.ShortLinkGroup
		tableName := fmt.Sprintf("short_link_group_%d", i)
		err = d.db.Table(tableName).WithContext(ctx).Find(&records).Error
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}
		result = append(result, records...)
	}
	return result, nil
}

// UpdateByGid 根据gid更新分组名称
func (d *shortLinkGroupsDao) UpdateByGid(ctx context.Context, gid string, name string) (*model.ShortLinkGroup, error) {
	//todo 用事务改写此处
	update := map[string]interface{}{}
	update["name"] = name
	update["updated_at"] = time.Now()
	group := &model.ShortLinkGroup{
		Gid: gid,
	}
	tDB := d.db.Table(group.TName()).WithContext(ctx).Where("gid = ?", gid).Updates(update)
	err := tDB.Error
	if tDB.RowsAffected == 0 {
		return nil, errors.New("no record is updated")
	}
	if err != nil {
		return nil, err
	}
	d.db.Table(group.TName()).WithContext(ctx).Where("gid = ?", gid).First(group)
	d.cache.HSet(ctx, group.CUsername, group)
	return group, nil
}
