package dao

import (
	"SnapLink/internal/model"
	"SnapLink/pkg/db"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// 根据目前的日期进行查询
// 根据对应的id 进行更新

type LinkStatsCache interface {
	Get(ctx context.Context, originalUrl string, date string, hour int) (*model.LinkAccessStat, error)
	Set(ctx context.Context, values map[string]any) error
	GetByDateHour(ctx context.Context, date string, hour int) ([]*model.LinkAccessStat, error)
	UpdateUip(ctx context.Context, originalUrl string, date string, hour int, ip string) error
	UpdateUv(ctx context.Context, originalUrl string, date string, hour int) error
	UpdatePv(ctx context.Context, originalUrl string, date string, hour int) error
}

type LinkStatsDao struct {
	db    *gorm.DB
	cache LinkStatsCache
	sfg   *singleflight.Group
}

// NewLinkStatsDao creating the dao interface
func NewLinkStatsDao(xCache LinkStatsCache) *LinkStatsDao {
	return &LinkStatsDao{
		db:    db.DB(),
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

// Get get a record by originalUrl, date and hour
func (d *LinkStatsDao) Get(ctx context.Context, originalUrl string, date string, hour int) (*model.LinkAccessStat, error) {
	return d.cache.Get(ctx, originalUrl, date, hour)
}

// Set
// 由于缓存的原因，这里需要将数据存储到缓存中
// 由于此数据是写多读少的数据，所以采用定时任务的方式进行数据的更新
func (d *LinkStatsDao) Set(ctx context.Context, data *model.LinkAccessStat) error {
	m := make(map[string]interface{})
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}
	err = d.cache.Set(ctx, m)
	if err != nil {
		return err
	}
	return nil
}

// Save 更新一或多条记录到数据库
func (d *LinkStatsDao) Save(ctx context.Context, data []*model.LinkAccessStat) error {
	//更新缓存
	err := d.db.WithContext(ctx).Save(data).Error
	if err != nil {
		return err
	}
	return nil
}

// 传入的 ctx 需要带有 cancel
func cronTask(ctx context.Context, d *LinkStatsDao) {
	ticker := time.NewTicker(1 * time.Hour)
	for {

		select {
		case <-ticker.C:
			{
				datas, err := d.cache.GetByDateHour(context.Background(), time.Now().Format("2006-01-02"), time.Now().Hour())
				if err != nil {
					//todo 日志处理
					fmt.Println("cronTask error", err)
					continue
				}
				d.Save(context.Background(), datas)

			}
		case <-ctx.Done():
			{
				//todo 日志处理
				fmt.Println("cronTask Done")
				break
			}
		}
	}
}

func (d *LinkStatsDao) UpdateUip(ctx context.Context, originalUrl string, date string, hour int, ip string) error {
	//TODO 需要考虑缓存失效情况下的处理
	return d.cache.UpdateUip(ctx, originalUrl, date, hour, ip)
}

func (d *LinkStatsDao) UpdateUv(ctx context.Context, originalUrl string, date string, hour int) error {
	return d.cache.UpdateUv(ctx, originalUrl, date, hour)
}

func (d *LinkStatsDao) UpdatePv(ctx context.Context, originalUrl string, date string, hour int) error {
	return d.cache.UpdatePv(ctx, originalUrl, date, hour)
}
