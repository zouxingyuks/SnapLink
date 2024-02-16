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
	Get(ctx context.Context, originalUrl string, date string, hour int) (*model.LinkAccessStatistic, error)
	Set(ctx context.Context, values map[string]any) error
	GetByDateHour(ctx context.Context, date string, hour int) ([]*model.LinkAccessStatistic, error)
	UpdateUip(ctx context.Context, originalUrl string, date string, hour int, ip string) error
	UpdateUv(ctx context.Context, originalUrl string, date string, hour int) error
	UpdatePv(ctx context.Context, originalUrl string, date string, hour int) error
	GetRecord(ctx context.Context, originalUrl string, date string, hour int) (*model.LinkAccessRecord, error)
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

// GetStatistic 获取访问统计
func (d *LinkStatsDao) GetStatistic(ctx context.Context, uri string, startDatetime, endDatetime string, pageNum, pageSize uint64) ([]model.LinkAccessStatistic, error) {
	//todo 基于缓存的设计
	//todo 调整查询表
	statiscs := []model.LinkAccessStatistic{}
	//todo 拓展信息的查询
	//使用此方法在多次查询时，只会进行一次 join 查询
	err := d.db.
		Table(model.LinkAccessStatistic{}.TName()).
		WithContext(ctx).
		Where("uri = ?  and timestampdiff(SECOND,datetime,?) <= 0 and timestampdiff(SECOND,datetime,?) >= 0", uri, startDatetime, endDatetime).
		Order("datetime desc").
		Offset(int((pageNum - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&statiscs).Error
	return statiscs, err
}

// GetRecord 获取访问记录
// 根据原始链接和时间来去精准查询
func (d *LinkStatsDao) GetRecord(ctx context.Context, uri string, startDatetime, endDatetime string, pageNum, pageSize uint64) ([]model.LinkAccessRecord, error) {
	//todo 基于缓存的设计
	//todo  如何多表查询下的高性能设计
	records := []model.LinkAccessRecord{}
	//todo 调整查询表
	err := d.db.
		Table(model.LinkAccessRecord{}.
			TName()).
		WithContext(ctx).
		Where("uri = ?  and timestampdiff(SECOND,datetime,?) <= 0 and timestampdiff(SECOND,datetime,?) >= 0", uri, startDatetime, endDatetime).
		Order("datetime desc").
		Offset(int((pageNum - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&records).Error
	return records, err

}

// UpdateUip 更新IP访问量
func (d *LinkStatsDao) UpdateUip(ctx context.Context, originalUrl string, date string, hour int, ip string) error {
	//TODO 需要考虑缓存失效情况下的处理
	return d.cache.UpdateUip(ctx, originalUrl, date, hour, ip)
}

// UpdateUv 更新UV访问量
func (d *LinkStatsDao) UpdateUv(ctx context.Context, originalUrl string, date string, hour int) error {
	return d.cache.UpdateUv(ctx, originalUrl, date, hour)
}

// UpdatePv 更新PV访问量
func (d *LinkStatsDao) UpdatePv(ctx context.Context, originalUrl string, date string, hour int) error {
	return d.cache.UpdatePv(ctx, originalUrl, date, hour)
}

// SaveAccessRecord 保存访问记录
func (d *LinkStatsDao) SaveAccessRecord(ctx context.Context, record *model.LinkAccessRecord) error {
	//TODO 考虑如何进行二次更改
	err := d.db.Table(record.TName()).WithContext(ctx).Create(record).Error
	return err
}

// Set
// 由于缓存的原因，这里需要将数据存储到缓存中
// 由于此数据是写多读少的数据，所以采用定时任务的方式进行数据的更新
func (d *LinkStatsDao) Set(ctx context.Context, data *model.LinkAccessStatistic) error {
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
func (d *LinkStatsDao) Save(ctx context.Context, data []*model.LinkAccessStatistic) error {
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
