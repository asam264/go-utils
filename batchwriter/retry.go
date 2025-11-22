package batchwriter

import (
	"fmt"
	"time"

	"github.com/asam264/go-utils/common"
)

// callHandlerWithRetry 指数退避重试
func (bw *BatchWriter) callHandlerWithRetry(items [][]byte) error {
	var lastErr error
	backoff := bw.config.BaseBackoff

	for attempt := 0; attempt < bw.config.MaxRetries; attempt++ {
		if err := bw.config.Handler(items); err != nil {
			lastErr = err
			bw.log("warn", "handler failed, retrying",
				common.Field{"attempt", attempt},
				common.Field{"error", err})

			select {
			case <-time.After(backoff):
			case <-bw.ctx.Done():
				return lastErr
			}

			backoff *= 2
			if backoff > bw.config.MaxBackoff {
				backoff = bw.config.MaxBackoff
			}
			continue
		}
		return nil
	}
	return lastErr
}

// saveToRedisWithRetry 将失败批次保存到 redis，带简单 retry
func (bw *BatchWriter) saveToRedisWithRetry(items [][]byte) error {
	if bw.config.RedisClient == nil {
		return fmt.Errorf("redis client not configured")
	}

	var lastErr error
	backoff := time.Second

	for i := 0; i < 3; i++ {
		err := bw.saveToRedis(items)
		if err == nil {
			return nil
		}
		lastErr = err
		bw.log("warn", "save to redis failed, retrying",
			common.Field{"attempt", i},
			common.Field{"error", err})

		select {
		case <-time.After(backoff):
		case <-bw.ctx.Done():
			return lastErr
		}
		backoff *= 2
	}
	return lastErr
}

