package dao

import (
	"context"
	"github.com/go-redis/redis/v8"
	"project-user/config"
	"time"
)

// Rc 是一个全局的 RedisCache 实例，用于在全局范围内访问 Redis 缓存。
var Rc *RedisCache

// RedisCache 提供了对 Redis 数据库进行简单操作的方法，如 Put 和 Get。
type RedisCache struct {
	rdb *redis.Client
}

// init 函数在包被导入时初始化全局的 RedisCache 实例。
func init() {
	// 从配置中读取 Redis 的配置信息，并创建一个新的 Redis 客户端实例。
	rdb := redis.NewClient(config.C.ReadRedisConfig())
	Rc = &RedisCache{
		rdb: rdb,
	}
}

// Put 方法用于将一个键值对放入 Redis 缓存中，并设置其过期时间。
// - ctx: 上下文，用于传递请求相关的配置或元数据。
// - key: 要存储的键。
// - value: 与键关联的值。
// - expire: 键的过期时间。
// 返回: 如果操作失败，返回错误；否则返回 nil。
func (rc *RedisCache) Put(ctx context.Context, key, value string, expire time.Duration) error {
	err := rc.rdb.Set(ctx, key, value, expire).Err()
	return err
}

// Get 方法用于从 Redis 缓存中获取与指定键关联的值。
// - ctx: 上下文，用于传递请求相关的配置或元数据。
// - key: 要查询的键。
// 返回: 如果操作成功，返回键对应的值和 nil；如果键不存在或操作失败，返回空字符串和错误。
func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result, err := rc.rdb.Get(ctx, key).Result()
	return result, err
}
