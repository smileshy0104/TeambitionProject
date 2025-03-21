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
}
