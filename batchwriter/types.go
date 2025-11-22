// Package batchwriter provides high-performance batch writing with retry and persistence
package batchwriter

import (
	"context"
	"sync"
	"time"

	"github.com/asam264/go-utils/common"
)

// RedisClient Redis客户端接口
type RedisClient interface {
	LPop(ctx context.Context, key string) ([]byte, error)
	RPush(ctx context.Context, key string, values ...[]byte) error
}

// Handler 批量处理函数类型
type Handler func(items [][]byte) error

// Config 批量写入配置
type Config struct {
	Name           string
	BatchSize      int
	FlushInterval  time.Duration
	RedisKeyPrefix string
	RedisClient    RedisClient
	Handler        Handler
	Logger         common.Logger

	// Retry policy
	MaxRetries    int
	BaseBackoff   time.Duration
	MaxBackoff    time.Duration
	RecoverPeriod time.Duration
}

// BatchWriter 批量写入器
type BatchWriter struct {
	config Config

	mu        sync.Mutex
	queue     [][]byte
	lastFlush time.Time

	ctx    context.Context
	cancel context.CancelFunc

	flushReqCh chan struct{}
	refreshT   *time.Ticker
	recoverT   *time.Ticker

	wg sync.WaitGroup

	redisKey string

	started bool
	closed  bool
}

