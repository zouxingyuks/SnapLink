package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

// getAll 全表扫描
// selectParam 填写非 id 的部分
func getAll(ctx context.Context, tableName string, db *gorm.DB, selectParam []string) (map[string][]any, error) {
	result := make(map[string][]any, len(selectParam))
	for _, param := range selectParam {
		result[param] = []any{}
	}
	// 此处专门针对深分页问题进行优化,因为此处是全量查询
	// 因此使用游标法进行查询
	var cursor int64
	const batchSize = 1000
	for {
		var records []map[string]any
		if err := db.WithContext(ctx).
			Table(tableName).
			Select(append([]string{"id"}, selectParam...)).
			Where("id > ?", cursor).
			Limit(batchSize).
			Scan(&records).Error; err != nil {
			return nil, err
		}
		l := len(records)
		if l == 0 {
			break
		}
		// 假设 record 是从数据库查询得到的一条记录
		switch id := records[l-1]["id"].(type) {
		case int:
			cursor = int64(id)
		case int64:
			cursor = id
		case int32:
			cursor = int64(id)
		case uint:
			cursor = int64(id)
		case uint64:
			cursor = int64(id)
		case uint32:
			cursor = int64(id)
		default:
			// 处理不支持的类型
			return nil, fmt.Errorf("unsupported type for id: %T", records[l-1]["id"])
		}
		for i := 0; i < l; i++ {
			for _, param := range selectParam {
				result[param] = append(result[param], records[i][param])
			}
		}
	}
	return result, nil
}
