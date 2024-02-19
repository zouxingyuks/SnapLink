package dao

import (
	"SnapLink/internal/model"
	"SnapLink/pkg/db"
	"context"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"time"
)

// 根据目前的日期进行查询
// 根据对应的id 进行更新

type LinkStatsCache interface {
	//GetByDateHour(ctx context.Context, date string, hour int) ([]*model.LinkAccessStatistic, error)
	UpdateIp(ctx context.Context, uri string, date string, hour int, ip string) error
	UpdateUv(ctx context.Context, uri string, date string, hour int) error
	UpdatePv(ctx context.Context, uri string, date string, hour int) error
	UpdateLocation(ctx context.Context, uri string, date string, hour int, location string) error
	UpdateUA(ctx context.Context, uri string, date string, hour int, browser, device string) error
	GetStatisticByDateHour(ctx context.Context, uri string, date string, hour int) (*model.LinkAccessStatistic, error)
	GetAllUri(ctx context.Context, date string, hour int) ([]string, error)
}

type LinkAccessStatisticDao struct {
	db    *gorm.DB
	cache LinkStatsCache
	sfg   *singleflight.Group
}

// NewLinkAccessStatisticDao creating the dao interface
func NewLinkAccessStatisticDao(xCache LinkStatsCache) *LinkAccessStatisticDao {
	return &LinkAccessStatisticDao{
		db:    db.DB(),
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

// GetStatistic 获取访问统计
func (d *LinkAccessStatisticDao) GetStatistic(ctx context.Context, uri string, startDatetime, endDatetime string, pageNum, pageSize uint64) ([]model.LinkAccessStatistic, error) {
	//todo 基于缓存的设计
	//todo 调整查询表
	var statistics []model.LinkAccessStatistic
	//todo 拓展信息的查询
	//使用此方法在多次查询时，只会进行一次 join 查询
	err := d.db.
		Table(model.LinkAccessStatistic{}.TName()).
		WithContext(ctx).
		Where("uri = ?  and timestampdiff(SECOND,datetime,?) <= 0 and timestampdiff(SECOND,datetime,?) >= 0", uri, startDatetime, endDatetime).
		Order("datetime desc").
		Offset(int((pageNum - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&statistics).Error
	return statistics, err
}

// GetStatisticByDay
// order format: a desc,b asc
func (d *LinkAccessStatisticDao) GetStatisticByDay(ctx context.Context, uri string, startDate, endDate string, order string, pageNum, pageSize uint64) ([]model.LinkAccessStatisticDay, error) {
	var datas []model.LinkAccessStatisticDay
	tx := d.db.WithContext(ctx).Table(model.LinkAccessStatistic{}.TName()).
		Select("uri",
			"date",
			"SUM(pv) AS today_pv",
			"SUM(uv) AS today_uv",
			"SUM(uip) AS today_uip").
		Where("DATEDIFF( date, ? ) >= 0", startDate).
		Where("DATEDIFF( date, ? ) <= 0", endDate).
		Group("uri,date").
		Order(order).
		Offset(int((pageNum - 1) * pageSize)).
		Limit(int(pageSize))
	if uri != "" {
		tx.Where("uri = ?", uri)
	}
	rows, err := tx.Rows()
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	//显示查询语句
	for rows.Next() {
		data := new(model.LinkAccessStatisticDay)
		d.db.ScanRows(rows, data)
		datas = append(datas, *data)
	}
	//处理 created_at 格式为日期
	return datas, nil
}

// GetRecord 获取访问记录
// 根据原始链接和时间来去精准查询
func (d *LinkAccessStatisticDao) GetRecord(ctx context.Context, uri string, startDatetime, endDatetime string, pageNum, pageSize uint64) ([]model.LinkAccessRecord, error) {
	//todo 基于缓存的设计
	//todo  如何多表查询下的高性能设计
	var records []model.LinkAccessRecord
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

// UpdateIp 更新IP访问量
func (d *LinkAccessStatisticDao) UpdateIp(ctx context.Context, uri string, date string, hour int, ip string) error {
	//TODO 需要考虑缓存失效情况下的处理
	return d.cache.UpdateIp(ctx, uri, date, hour, ip)
}

// UpdateUv 更新UV访问量
func (d *LinkAccessStatisticDao) UpdateUv(ctx context.Context, uri string, date string, hour int) error {
	return d.cache.UpdateUv(ctx, uri, date, hour)
}

// UpdatePv 更新PV访问量
func (d *LinkAccessStatisticDao) UpdatePv(ctx context.Context, uri string, date string, hour int) error {
	return d.cache.UpdatePv(ctx, uri, date, hour)
}

// SaveAccessRecord 保存访问记录
func (d *LinkAccessStatisticDao) SaveAccessRecord(ctx context.Context, record *model.LinkAccessRecord) error {
	err := d.db.Table(record.TName()).WithContext(ctx).Create(record).Error
	return err
}

// UpdateLocation 更新访问地理位置
func (d *LinkAccessStatisticDao) UpdateLocation(ctx context.Context, uri string, date string, hour int, location string) error {
	//todo 进行二次优化
	return d.cache.UpdateLocation(ctx, uri, date, hour, location)
}

// UpdateUA 更新访问设备信息
func (d *LinkAccessStatisticDao) UpdateUA(ctx context.Context, uri string, date string, hour int, browser, device string) error {
	//todo 进行二次优化
	return d.cache.UpdateUA(ctx, uri, date, hour, browser, device)
}

//// Set
//// 由于缓存的原因，这里需要将数据存储到缓存中
//// 由于此数据是写多读少的数据，所以采用定时任务的方式进行数据的更新
//func (d *LinkAccessStatisticDao) Set(ctx context.Context, data *model.LinkAccessStatistic) error {
//	m := make(map[string]interface{})
//	bytes, err := json.Marshal(data)
//	if err != nil {
//		return err
//	}
//	err = json.Unmarshal(bytes, &m)
//	if err != nil {
//		return err
//	}
//	err = d.cache.Set(ctx, m)
//	if err != nil {
//		return err
//	}
//	return nil
//}

// SaveToDB 将数据从缓存中提取并保存到数据库
func (d *LinkAccessStatisticDao) SaveToDB(ctx context.Context, uri string, date string, hour int) error {
	//todo 优化此处的错误处理
	//1.从缓存中调取特定时间的统计数据
	//2.将统计数据存储到数据库中
	//3.将统计数据存储到缓存中
	data, err := d.cache.GetStatisticByDateHour(ctx, uri, date, hour)
	if err != nil {
		//todo 日志处理
		return err
	}
	err = d.db.WithContext(ctx).Save(data).Error
	if err != nil {
		return err
	}
	return nil
}

// 传入的 ctx 需要带有 cancel
func cronTask(d *LinkAccessStatisticDao) {
	crontab := cron.New()
	_, err := crontab.AddFunc("0 0 * * *", func() {
		//todo 日志处理
		//1.获取所有的uri，根据时间进行更新
		date := time.Now().Format("2006-01-02")
		hour := time.Now().Hour()
		uris, err := d.cache.GetAllUri(context.Background(), date, hour)
		if err != nil {
			//todo 日志处理
			//todo 错误处理
		}
		for _, uri := range uris {
			err := d.SaveToDB(context.Background(), uri, date, hour)
			if err != nil {
				//todo 日志处理
				//todo 错误处理
			}
		}
	})
	if err != nil {
		panic(errors.Wrap(err, "cron task set error"))
	}
	crontab.Start()
}
