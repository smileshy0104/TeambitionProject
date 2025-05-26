package repo

import (
	"context"
	"time"
)

type Cache interface {
	// Put 缓存
	Put(ctx context.Context, key, value string, expire time.Duration) error
	// Get 获取缓存
	Get(ctx context.Context, key string) (string, error)
	HSet(ctx context.Context, key string, field string, value string)
	HKeys(ctx context.Context, key string) ([]string, error)
	Delete(background context.Context, keys []string)
}
