package cache

import (
	"time"
)

// ILocalCache 定义了本地缓存应有的行为。
// 它抽象了对本地缓存的主要操作，使得可以在不同本地缓存实现之间进行切换和测试。
type ILocalCache interface {
	// Get 从缓存中检索与指定键关联的值。
	// 如果找到值，返回值和 true；否则返回 nil 和 false。
	// key: 要检索的缓存键。
	// 返回值: 与键关联的值（如果存在）和一个布尔值，指示是否找到了值。
	Get(key interface{}) (interface{}, bool)

	// Set 将一个值与指定的键关联到缓存中。
	// 如果操作成功，返回 true；如果因为某种原因（如容量限制）而未能设置，返回 false。
	// key: 要在缓存中设置的键。
	// value: 与键关联的值。
	// cost: 值的成本，用于缓存淘汰决策。
	// 返回值: 指示值是否成功设置到缓存中的布尔值。
	Set(key, value interface{}, cost int64) bool

	// SetWithTTL 将一个值与键关联到缓存中，并设置一个过期时间。
	// 如果操作成功，返回 true；如果未能设置，返回 false。
	// key: 要在缓存中设置的键。
	// value: 与键关联的值。
	// cost: 值的成本，用于缓存淘汰决策。
	// ttl: 值在缓存中存活的时间。
	// 返回值: 指示值是否成功设置到缓存中的布尔值。
	SetWithTTL(key, value interface{}, cost int64, ttl time.Duration) bool

	// Del 从缓存中删除与指定键关联的值。
	// key: 要删除的缓存键。
	Del(key interface{})

	// GetTTL 检索指定键的剩余生存时间（TTL）。
	// 如果找到键并且它的值没有过期，返回剩余的 TTL 和 true；否则返回 0 和 false。
	// key: 要检查 TTL 的缓存键。
	// 返回值: 键的剩余生存时间和一个布尔值，指示键是否存在且未过期。
	GetTTL(key interface{}) (time.Duration, bool)

	// Close 关闭缓存，并释放相关资源。
	Close()

	// Clear 清除缓存中的所有键值对。
	Clear()

	// MaxCost 返回缓存的最大容量限制。
	// 返回值: 缓存的最大成本限制。
	MaxCost() int64

	// UpdateMaxCost 更新缓存的最大容量限制。
	// maxCost: 新的最大容量限制。
	UpdateMaxCost(maxCost int64)

	// Wait 等待所有缓存操作完成。
	// 通常在关闭缓存之前调用，以确保所有操作都已经完成。
	Wait()
}
