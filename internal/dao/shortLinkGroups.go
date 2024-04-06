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
	UpdateSortOrderByGidAndUsername(ctx context.Context, gids []string, sortOrders []int, username string) error
	DelByGidAndUsername(ctx context.Context, gid, username string) error
}
type shortLinkGroupsDao struct {
	db  *gorm.DB
	sfg *singleflight.Group
}

// NewShortLinkGroupDao creating the dao interface
func NewShortLinkGroupDao(db *gorm.DB) ShortLinkGroupDao {
	return &shortLinkGroupsDao{
		db:  db,
		sfg: new(singleflight.Group),
	}
}

// Create 创建短链接分组
func (d *shortLinkGroupsDao) Create(ctx context.Context, group *model.ShortLinkGroup) error {
	err := d.db.Table(group.TName()).WithContext(ctx).Create(group).Error
	if err != nil {
		return err
	}
	return nil
}

// GetAllByCUser 根据创建人获取所有的分组
// 此处缓存的判定有两点:
// 1. 缓存中没有数据, 从数据库中获取数据, 并且将数据写入缓存.此情况返回的 records 虽然是空的,但是不会返回 nil
// 2. 缓存中特别指明没有数据,特别指明没有数据的情况是为了防止缓存穿透,其会返回一个 nil, 用于区别根本没有查到数据的情况
func (d *shortLinkGroupsDao) GetAllByCUser(ctx context.Context, cUser string) ([]*model.ShortLinkGroup, error) {
	// 从缓存中获取
	records, _ := cache.SLGroup().Get(ctx, cUser)
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
	err := d.db.Table(tableName).
		WithContext(ctx).
		Where("c_username = ?", cUser).
		Order("sort_order DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		// 缓存中特别指明没有数据

		return records, cache.SLGroup().SetCacheWithNotFound(ctx, cUser)
	}
	return records, cache.SLGroup().Set(ctx, cUser, records)

}

// GetAll 获取所有的分组
func (d *shortLinkGroupsDao) GetAll(ctx context.Context) (result []*model.ShortLinkGroup, err error) {
	for i := 0; i < model.SLGroupShardingNum; i++ {
		var records []*model.ShortLinkGroup
		tableName := fmt.Sprintf("t_link_group%d", i)
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
	return group, nil
}
func (d *shortLinkGroupsDao) UpdateSortOrderByGidAndUsername(ctx context.Context, gids []string, sortOrders []int, username string) error {
	lg := len(gids)
	ls := len(sortOrders)
	if lg != ls || lg == 0 {
		return errors.New("gids and sortOrders length not equal or are zero")
	}

	tableName := model.ShortLinkGroup{CUsername: username}.TName()

	// 构建CASE WHEN THEN语句用于批量更新
	caseStmt := "CASE"
	for i, gid := range gids {
		caseStmt += fmt.Sprintf(" WHEN gid = '%s' THEN %d", gid, sortOrders[i])
	}
	caseStmt += " END"

	// 构建完整的批量更新SQL语句
	query := fmt.Sprintf("UPDATE `%s` SET sort_order = %s WHERE c_username = ? AND gid IN (?)", tableName, caseStmt)

	// 在GORM中使用事务处理
	err := d.db.Transaction(func(tx *gorm.DB) error {
		// 使用WithContext确保上下文传递
		if err := tx.WithContext(ctx).Exec(query, username, gids).Error; err != nil {
			// 如果执行失败，返回错误将自动回滚
			return err
		}
		// 如果执行成功，返回nil提交事务
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update sort order by gid and username: %w", err)
	}
	return nil
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
	return cache.SLGroup().Del(ctx, username)
}
