package dao

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/model"
	"SnapLink/pkg/db"
	"context"
	"fmt"
	"github.com/zhufuyi/sponge/pkg/logger"
	"gorm.io/gorm"
	"sync"
)

const (
	usernameBF = "username"
	uriBF      = "uri"
)

type bfMakeFn = func(db *gorm.DB) (bool, []error)

// 使用类工厂模式的思想来批量执行
var bfs = map[string]bfMakeFn{
	uriBF:      makeUriBF,
	usernameBF: makeUsernameBF,
}

type FixDao struct {
	db *gorm.DB
}

func NewFixDao() *FixDao {
	return &FixDao{
		db: db.DB(),
	}
}
func (d *FixDao) RebulidBF() (errs []error) {
	// 重建布隆过滤器
	for name, fn := range bfs {
		// 删除布隆过滤器
		if err := cache.BFDelete(context.Background(), name); err != nil {
			errs = append(errs, err)
			return errs
		}
		// 重新创建布隆过滤器
		ok, errs := fn(d.db)
		if !ok {
			return errs
		}
	}
	return nil
}

// makeUsernameBF 生成 username 的布隆过滤器
func makeUsernameBF(db *gorm.DB) (bool, []error) {
	// 此结构体使用的不多,因此临时构建即可
	type user struct {
		ID       uint
		Username string
	}
	//todo 此处改为远程配置
	err := cache.BFCreate(context.Background(), usernameBF, 0.001, 1e9)
	// 如果创建失败，说明已经存在，正常结束即可
	// 如果创建成功，说明不存在，需要添加数据
	if err == nil {
		//记录日志
		errs := make([]error, 0, model.TUserShardingNum)
		wg := sync.WaitGroup{}
		for i := 0; i < model.TUserShardingNum; i++ {
			tableName := fmt.Sprintf("t_user_%d", i)
			wg.Add(1)
			go func(id int, tName string) {
				defer func() {
					wg.Done()
				}()
				// 此处专门针对深分页问题进行优化,因为此处是全量查询
				// 因此使用游标法进行查询
				var cursor uint
				for {
					var records []user
					err := db.Table(tName).Select("id, username").Where("id > ?", cursor).Limit(1000).Scan(&records).Error
					if err != nil {
						logger.Err(err)
						break
					}
					l := len(records)
					if l == 0 {
						break
					}
					cursor = records[l-1].ID
					usernames := make([]string, 0, l)
					// 取出所有的 uri,注意不要使用 range 语法糖，因为 range 语法糖会创建临时变量
					for i := 0; i < l; i++ {
						usernames = append(usernames, records[i].Username)
					}
					err = cache.BFMAdd(context.Background(), usernameBF, usernames...)
					if err != nil {
						errs[id] = err
						break
					}
				}
			}(i, tableName)
		}
		wg.Wait()
		for _, err := range errs {
			if err != nil {
				return false, errs
			}
		}
	}
	return true, nil
}

// makeUriBF 生成 uri 的布隆过滤器
func makeUriBF(db *gorm.DB) (bool, []error) {
	// 此结构体使用的不多,因此临时构建即可
	type redirect struct {
		ID  uint
		URI string
	}
	//todo 此处改为远程配置
	err := cache.BFCreate(context.Background(), uriBF, 0.01, 1e9)
	if err == nil {
		//记录日志
		errs := make([]error, 0, model.RedirectShardingNum)
		wg := sync.WaitGroup{}
		for i := 0; i < model.RedirectShardingNum; i++ {
			tableName := fmt.Sprintf("redirect_%d", i)
			wg.Add(1)
			go func(id int, tName string) {
				defer func() {
					wg.Done()
				}()
				// 此处专门针对深分页问题进行优化,因为此处是全量查询
				// 因此使用游标法进行查询
				var cursor uint
				for {

					var records []redirect

					err := db.Table(tName).Select("id, uri").Where("id > ?", cursor).Limit(1000).Scan(&records).Error

					if err != nil {
						logger.Err(err)
						break
					}
					l := len(records)
					if l == 0 {
						break
					}
					cursor = records[l-1].ID
					uris := make([]string, 0, l)
					// 取出所有的 uri,注意不要使用 range 语法糖，因为 range 语法糖会创建临时变量
					for i := 0; i < l; i++ {
						uris = append(uris, records[i].URI)
					}
					err = cache.BFMAdd(context.Background(), uriBF, uris...)
					if err != nil {
						errs[id] = err
						break
					}
				}
			}(i, tableName)
		}
		wg.Wait()
		for _, err := range errs {
			if err != nil {
				return false, errs
			}
		}
	}
	return true, nil
}
