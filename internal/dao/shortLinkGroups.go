package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	err := d.db.Table(tableName).
		WithContext(ctx).
		Where("c_username = ?", cUser).
		Order("sort_order ASC").
		Find(&records).Error
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

// UpdateSortOrderByGidAndUsername 批量更新排序
// todo 索引二次优化
func (d *shortLinkGroupsDao) UpdateSortOrderByGidAndUsername(ctx context.Context, gids []string, sortOrders []int, username string) error {
	lg := len(gids)
	ls := len(sortOrders)
	if lg != ls || lg == 0 {
		return errors.New("gids and sortOrders length not equal or are zero")
	}
	// 开始构建CASE WHEN语句
	var cases strings.Builder
	for i, gid := range gids {
		cases.WriteString(fmt.Sprintf("WHEN gid = '%s' THEN %d ", gid, sortOrders[i]))
	}

	// 构建完整的SQL语句
	sql := fmt.Sprintf("UPDATE %s SET sort_order = CASE %s END WHERE gid IN (?) AND c_username = ?", model.ShortLinkGroup{}.TName(), cases.String())

	// 在GORM中使用事务处理
	err := d.db.Transaction(func(tx *gorm.DB) error {
		// 使用WithContext确保上下文传递
		if err := tx.WithContext(ctx).Exec(sql, gids, username).Error; err != nil {
			// 如果执行失败，返回错误将自动回滚
			return err
		}
		// 如果执行成功，返回nil提交事务
		return nil
	})

	// 处理事务结果
	if err != nil {
		return fmt.Errorf("failed to update sort order by gid and username: %w", err)
	}
	// 更新缓存
	return d.cache.Del(ctx, username)
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
