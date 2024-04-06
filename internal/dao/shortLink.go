package dao

import (
	"SnapLink/internal/bloomFilter"
	"SnapLink/internal/cache"
	"SnapLink/internal/custom_err"
	"SnapLink/internal/model"
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	cacheBase "github.com/zhufuyi/sponge/pkg/cache"
	"github.com/zhufuyi/sponge/pkg/logger"
	"sync"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

var instanceShortLink struct {
	IShortLinkDao
	sync.Once
}

func ShortLinkDao() IShortLinkDao {
	instanceShortLink.Once.Do(func() {
		var err error
		instanceShortLink.IShortLinkDao, err = NewShortLinkDao(model.GetDB(), model.GetRedisCli())
		if err != nil {
			logger.Panic(errors.Wrap(ErrInitDaoFailed, err.Error()).Error())
		}
	})
	return instanceShortLink.IShortLinkDao
}

// IShortLinkDao 定义接口
type IShortLinkDao interface {
	Create(ctx context.Context, table *model.ShortLink) error
	CreateBatch(ctx context.Context, tables []*model.ShortLink) (*model.ShortLink, error)
	List(ctx context.Context, gid string, page, pageSize int) ([]*model.ShortLink, error)
	Count(ctx context.Context, gid string) (int64, error)
	Delete(ctx context.Context, uri string) error
	GeRedirectByURI(ctx context.Context, uri string) (*model.Redirect, error)
	Update(ctx context.Context, shortLink *model.ShortLink) error
	UpdateWithMove(ctx context.Context, shortLink *model.ShortLink, newGid string) error
}

type shortLinkDao struct {
	db  *gorm.DB
	sfg *singleflight.Group
}

// NewShortLinkDao 创建 shortLinkDao
func NewShortLinkDao(db *gorm.DB, client *redis.Client) (IShortLinkDao, error) {
	var err error
	dao := &shortLinkDao{
		db:  db,
		sfg: new(singleflight.Group),
	}
	if err != nil {
		return nil, errors.Wrap(err, "NewShortLinkDao Failed")
	}
	if err != nil {
		return nil, errors.Wrap(err, "NewShortLinkDao Failed")
	}
	return dao, nil
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
func (d *shortLinkDao) List(ctx context.Context, gid string, page, pageSize int) ([]*model.ShortLink, error) {
	var list []*model.ShortLink
	tableName := (&model.ShortLink{Gid: gid}).TName()
	//sql := fmt.Sprintf("SELECT * FROM %s WHERE gid =? AND id >= (SELECT id FROM %s WHERE gid = ? LIMIT 1 OFFSET ?) ORDER BY id LIMIT ?", tableName, tableName)
	//custom_err := d.db.WithContext(ctx).Table(tableName).Raw(sql, gid, gid, (page-1)*pageSize, pageSize).Find(&list).Error
	// 构建子查询
	subQuery := d.db.WithContext(ctx).
		Select("id").
		Table(tableName).
		Where("gid = ?", gid).
		Order("id").
		Limit(1).
		Offset((page - 1) * pageSize)
	// 主查询
	err := d.db.WithContext(ctx).
		Table(tableName).
		Where("gid = ?", gid).
		Where("id >= (?)", subQuery).
		Order("id").
		Limit(pageSize).
		Find(&list).Error
	return list, err
}

// Count 计算总数
func (d *shortLinkDao) Count(ctx context.Context, gid string) (int64, error) {
	count, err := cache.ShortLinkGroupCountCache().Get(ctx, gid)
	if err == nil {
		return count, nil
	}

	if errors.Is(err, custom_err.ErrCacheNotFound) {
		val, err, _ := d.sfg.Do(gid, func() (interface{}, error) { //nolint
			// 从数据库中查询
			tableName := model.ShortLink{Gid: gid}.TName()
			total := new(int64)
			err = d.db.WithContext(ctx).Table(tableName).Where("gid = ?", gid).Count(total).Error
			if err != nil {
				// 设置空值来防御缓存穿透
				if errors.Is(err, custom_err.ErrRecordNotFound) {
					err = cache.ShortLinkGroupCountCache().SetCacheWithNotFound(ctx, gid)
					if err != nil {
						return nil, err
					}
					return nil, custom_err.ErrRecordNotFound
				}
				return nil, err
			}
			// 设置缓存
			err = cache.ShortLinkGroupCountCache().Set(ctx, gid, *total)
			if err != nil {
				logger.Err(errors.Wrap(err, "设置缓存失败"))
			}
			return total, nil
		})
		if err != nil {
			return 0, err
		}
		total, ok := val.(*int64)
		if !ok {
			return 0, custom_err.ErrRecordNotFound
		}
		return *total, nil
	}
	// 其他错误
	return 0, err
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

func (d *shortLinkDao) GeRedirectByURI(ctx context.Context, uri string) (*model.Redirect, error) {
	// 使用布隆过滤器进行数据存在性判断
	// 如果不存在，则直接返回
	// 此处使用布隆过滤器的原因是: 减少大量空值造成的缓存内存占用过大
	exist, err := bloomFilter.BFExists(ctx, "uri", uri)
	if !exist {
		return nil, custom_err.ErrRecordNotFound
	}
	// 查询缓存，是否查到数据
	record, err := cache.Redirect().Get(ctx, uri)
	if err == nil {
		if record.OriginalURL == "" {
			return nil, custom_err.ErrRecordNotFound
		}
		return record, nil
	}
	if errors.Is(err, custom_err.ErrCacheNotFound) {
		// 基于 singleflight 进行并发调用合并,主要的性能优化点在于:
		//1. 减少数据库压力：通过合并请求，减少了对数据库的总体访问次数，从而降低了数据库的负载。
		//2. 节省时间：避免了多次加锁解锁的过程，因为对于相同的资源只进行了一次查询，减少了时间消耗。
		//3. 简化缓存策略：由于所有相同的请求都等待同一次查询结果，因此不再需要二次缓存查询，简化了缓存管理。
		val, err, _ := d.sfg.Do(uri, func() (interface{}, error) { //nolint
			record = &model.Redirect{
				Uri:         uri,
				Gid:         "",
				OriginalURL: "",
			}
			// 使用 uri 进行查询,专门对 uri 做了索引优化
			err = d.db.WithContext(ctx).Table(record.TName()).Where("uri = ?", uri).First(record).Error
			if err != nil {
				// 设置空值来防御缓存穿透
				if errors.Is(err, custom_err.ErrRecordNotFound) {
					err = cache.ShortLinkGroupCountCache().SetCacheWithNotFound(ctx, uri)
					if err != nil {
						return nil, err
					}
					return nil, custom_err.ErrRecordNotFound
				}
				return nil, err
			}
			// 设置缓存
			err = cache.Redirect().Set(ctx, uri, record, cache.RedirectsExpireTime)
			if err != nil {
				logger.Err(errors.Wrap(err, "设置缓存失败"))
			}
			return record, nil
		})
		if err != nil {
			return nil, err
		}
		info, ok := val.(*model.Redirect)
		if !ok {
			return nil, custom_err.ErrRecordNotFound
		}
		return info, nil
	} else if errors.Is(err, cacheBase.ErrPlaceholder) {
		return nil, custom_err.ErrRecordNotFound
	}

	// 快速失败，如果是其他错误，直接返回
	return nil, err
}

// Update 更新短链接
func (d *shortLinkDao) Update(ctx context.Context, shortLink *model.ShortLink) error {
	redirect := &model.Redirect{
		Uri:         shortLink.Uri,
		Gid:         shortLink.Gid,
		OriginalURL: shortLink.OriginUrl,
	}
	// 同时更新短链接和重定向
	err := d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(redirect.TName()).WithContext(ctx).
			Where("uri = ?", redirect.Uri).Updates(redirect).Error; err != nil {
			return err
		}
		return tx.Table(shortLink.TName()).WithContext(ctx).
			Where("uri = ?", shortLink.Uri).Updates(shortLink).Error

	})
	return err

}

// UpdateWithMove 更新短链接
// 取出短链接，移动到新的分组
func (d *shortLinkDao) UpdateWithMove(ctx context.Context, shortLink *model.ShortLink, newGid string) error {
	redirect := &model.Redirect{
		Uri:         shortLink.Uri,
		Gid:         newGid,
		OriginalURL: shortLink.OriginUrl,
	}
	// 同时更新短链接和重定向
	err := d.db.Transaction(func(tx *gorm.DB) error {
		// redirect 路由可以直接更新
		if err := tx.Table(redirect.TName()).WithContext(ctx).
			Where("uri = ?", redirect.Uri).Updates(redirect).Error; err != nil {
			return err
		}
		tableName := shortLink.TName()
		// shortLink 需要先删除，再插入
		// 定位到原来的短链接
		if err := tx.Table(tableName).WithContext(ctx).Where("uri = ?", shortLink.Uri).Find(shortLink).Error; err != nil {
			return err
		}
		// 硬删除原来的短链接
		if err := tx.Table(tableName).WithContext(ctx).Where("id = ?", shortLink.ID).Unscoped().Delete(shortLink).Error; err != nil {
			return err
		}
		// 插入新的短链接
		shortLink.ID = 0
		shortLink.Gid = newGid
		return tx.Table(shortLink.TName()).WithContext(ctx).Create(shortLink).Error
	})
	return err

}
