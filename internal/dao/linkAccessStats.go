package dao

import (
	"SnapLink/internal/model"
	"context"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
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
		db:    model.GetDB(),
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

//// Set
//// 由于缓存的原因，这里需要将数据存储到缓存中
//// 由于此数据是写多读少的数据，所以采用定时任务的方式进行数据的更新
//func (d *LinkAccessStatisticDao) Set(ctx context.Context, data *model.LinkAccessStatistic) error {
//	m := make(map[string]interface{})
//	bytes, custom_err := json.Marshal(data)
//	if custom_err != nil {
//		return custom_err
//	}
//	custom_err = json.Unmarshal(bytes, &m)
//	if custom_err != nil {
//		return custom_err
//	}
//	custom_err = d.slCache.Set(ctx, m)
//	if custom_err != nil {
//		return custom_err
//	}
//	return nil
//}

func (d *LinkAccessStatisticDao) GetBasicByUri(ctx context.Context, gid string, uris []string) (map[string]*model.LinkAccessStatisticBasic, error) {
	var statistics []*model.LinkAccessStatisticBasic
	tableName := model.LinkAccessStatisticBasic{Gid: gid}.TName()
	err := d.db.
		WithContext(ctx).
		Table(tableName).
		Where("gid = ? and uri in (?)", gid, uris).
		Find(&statistics).Error
	m := make(map[string]*model.LinkAccessStatisticBasic)
	l := len(statistics)
	for i := 0; i < l; i++ {
		m[statistics[i].URI] = statistics[i]
	}
	return m, err
}
