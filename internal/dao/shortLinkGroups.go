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
	UpdateByGidAndUsername(ctx context.Context, gid string, name, username string) (*model.ShortLinkGroup, error)
	DelByGidAndUsername(ctx context.Context, gid, username string) error
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
	if err != nil {
		return err
	}
	return d.cache.HSet(ctx, group.CUsername, group)
}

// GetAllByCUser 根据创建人获取所有的分组
// 此处缓存的判定有两点:
// 1. 缓存中没有数据, 从数据库中获取数据, 并且将数据写入缓存.此情况返回的 records 虽然是空的,但是不会返回 nil
// 2. 缓存中特别指明没有数据,特别指明没有数据的情况是为了防止缓存穿透,其会返回一个 nil, 用于区别根本没有查到数据的情况
func (d *shortLinkGroupsDao) GetAllByCUser(ctx context.Context, cUser string) ([]*model.ShortLinkGroup, error) {
	// 从缓存中获取
	records, _ := d.cache.HGetALL(ctx, cUser)
	// 如果缓存中特别指明没有数据，直接返回
	if records == nil {
		records = make([]*model.ShortLinkGroup, 0)
		return records, nil
	}
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
	if len(records) == 0 {
		// 缓存中特别指明没有数据

		return records, d.cache.HSetEmpty(ctx, cUser)
	}
	return records, d.cache.HMSet(ctx, cUser, records)

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

// UpdateByGidAndUsername 根据gid更新分组名称
func (d *shortLinkGroupsDao) UpdateByGidAndUsername(ctx context.Context, gid string, name, username string) (*model.ShortLinkGroup, error) {
	//todo 用事务改写此处
	update := map[string]interface{}{}
	update["name"] = name
	update["updated_at"] = time.Now()
	group := &model.ShortLinkGroup{
		Gid:       gid,
		CUsername: username,
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

// DelByGidAndUsername 根据gid删除分组
func (d *shortLinkGroupsDao) DelByGidAndUsername(ctx context.Context, gid, username string) error {
	group := &model.ShortLinkGroup{
		Gid:       gid,
		CUsername: username,
	}
	err := d.db.Table(group.TName()).WithContext(ctx).Where("gid = ?", gid).Delete(group).Error
	if err != nil {
		return err
	}
	return d.cache.HDel(ctx, username, gid)
}
