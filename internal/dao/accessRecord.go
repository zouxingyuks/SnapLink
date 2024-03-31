package dao

import (
	"SnapLink/internal/model"
	"context"
	"fmt"
	"gorm.io/gorm"
)

var _ LinkAccessRecordDao = (*linkAccessRecordDao)(nil)

// LinkAccessRecordDao 访问记录
type LinkAccessRecordDao interface {
	// Create 创建访问记录
	Create(ctx context.Context, record *model.LinkAccessRecord) error
	// CreateBatch 批量创建访问记录
	CreateBatch(ctx context.Context, records []*model.LinkAccessRecord) (int, error)
	// ListByUri 获取访问记录
	ListByUri(ctx context.Context, uri string, page, pageSize int) (int, []*model.LinkAccessRecord, error)
}

type linkAccessRecordDao struct {
	db *gorm.DB
}

// NewAccessRecord 创建访问记录
func NewAccessRecord(db *gorm.DB) LinkAccessRecordDao {
	d := &linkAccessRecordDao{
		db: db,
	}
	return d
}

// Create 创建访问记录
func (d *linkAccessRecordDao) Create(ctx context.Context, record *model.LinkAccessRecord) error {
	err := d.db.Table(record.TName()).WithContext(ctx).Create(record).Error
	return err
}

// CreateBatch 批量创建访问记录
// 返回第几个记录出错
func (d *linkAccessRecordDao) CreateBatch(ctx context.Context, records []*model.LinkAccessRecord) (int, error) {
	i, l := 0, len(records)
	err := d.db.Transaction(func(tx *gorm.DB) error {
		for i = 0; i < l; i++ {
			err := tx.Table(records[i].TName()).WithContext(ctx).Create(records[i]).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	return i, err
}

// ListByUri 获取指定uri的访问记录
// todo  需要重构,深分页的解决有问题
func (d *linkAccessRecordDao) ListByUri(ctx context.Context, uri string, page, pageSize int) (int, []*model.LinkAccessRecord, error) {
	var ids []uint
	var list []*model.LinkAccessRecord
	tableName := model.LinkAccessRecord{URI: uri}.TName()
	total := new(int64)
	sql := fmt.Sprintf("SELECT * FROM %s WHERE id IN (?)", tableName)
	// 事务
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Table(tableName).Where("uri = ?", uri).Count(total).Error
		if err != nil {
			return err
		}
		err = tx.Table(tableName).Select("id").
			Where("uri = ?", uri).
			Limit(pageSize).Offset((page - 1) * pageSize).
			Find(&ids).Error
		if err != nil {
			return err
		}
		err = tx.Raw(sql, ids).Scan(&list).Error
		return err
	})
	return int(*total), list, err
}
