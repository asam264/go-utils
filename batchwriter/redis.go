package batchwriter

import (
	"fmt"

	"github.com/asam264/go-utils/common"
)

// saveToRedis 直接推入 Redis list
func (bw *BatchWriter) saveToRedis(items [][]byte) error {
	if bw.config.RedisClient == nil {
		return fmt.Errorf("redis client not configured")
	}

	if err := bw.config.RedisClient.RPush(bw.ctx, bw.redisKey, items...); err != nil {
		return fmt.Errorf("save to redis failed: %w", err)
	}

	bw.log("info", "saved to Redis", common.Field{"count", len(items)})
	return nil
}

// recoverFromRedisOnce 启动时尝试恢复 Redis 中的内容
func (bw *BatchWriter) recoverFromRedisOnce() error {
	if bw.config.RedisClient == nil {
		return nil
	}
	return bw.replayFromRedisBackground()
}

// replayFromRedisBackground 从 redis 拉取一批并交给 handler
func (bw *BatchWriter) replayFromRedisBackground() error {
	if bw.config.RedisClient == nil {
		return nil
	}

	batch := bw.config.BatchSize
	items := make([][]byte, 0, batch)

	for i := 0; i < batch; i++ {
		res, err := bw.config.RedisClient.LPop(bw.ctx, bw.redisKey)
		if err != nil {
			break
		}
		items = append(items, res)
	}

	if len(items) == 0 {
		return nil
	}

	if err := bw.callHandlerWithRetry(items); err != nil {
		bw.log("error", "replay handler failed, pushing back to redis",
			common.Field{"count", len(items)},
			common.Field{"error", err})
		if err2 := bw.saveToRedisWithRetry(items); err2 != nil {
			bw.log("error", "replay pushback failed", common.Field{"error", err2})
		}
		return err
	}

	bw.log("info", "replay processed", common.Field{"count", len(items)})
	return nil
}

