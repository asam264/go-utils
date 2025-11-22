package batchwriter

import (
	"context"
	"fmt"
	"time"

	"github.com/asam264/go-utils/common"
)

// New 创建新的批量写入器
func New(cfg Config) (*BatchWriter, error) {
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	applyDefaults(&cfg)

	ctx, cancel := context.WithCancel(context.Background())
	bw := &BatchWriter{
		config:     cfg,
		queue:      make([][]byte, 0, cfg.BatchSize),
		lastFlush:  time.Now(),
		ctx:        ctx,
		cancel:     cancel,
		flushReqCh: make(chan struct{}, 1),
		refreshT:   time.NewTicker(cfg.FlushInterval),
		recoverT:   time.NewTicker(cfg.RecoverPeriod),
		redisKey:   fmt.Sprintf("%s:%s", cfg.RedisKeyPrefix, cfg.Name),
	}
	return bw, nil
}

// Start 启动后台 worker
func (bw *BatchWriter) Start() error {
	bw.mu.Lock()
	if bw.started {
		bw.mu.Unlock()
		return fmt.Errorf("BatchWriter %s already started", bw.config.Name)
	}
	bw.started = true
	bw.mu.Unlock()

	// 尝试从 redis 恢复一次（启动时）
	if err := bw.recoverFromRedisOnce(); err != nil {
		bw.log("error", "initial recover failed", common.Field{"error", err})
	}

	bw.wg.Add(1)
	go bw.worker()

	bw.log("info", "BatchWriter started",
		common.Field{"batchSize", bw.config.BatchSize},
		common.Field{"flushInterval", bw.config.FlushInterval})
	return nil
}

// Push 将一条数据放入队列
func (bw *BatchWriter) Push(data []byte) error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if !bw.started || bw.closed {
		return fmt.Errorf("BatchWriter %s not started or closed", bw.config.Name)
	}

	bw.queue = append(bw.queue, data)

	// 如果达到阈值，非阻塞触发一次 flush 请求
	if len(bw.queue) >= bw.config.BatchSize {
		select {
		case bw.flushReqCh <- struct{}{}:
		default:
		}
	}
	return nil
}

// worker 串行化处理 flush 请求和周期性 flush、以及 redis 重放
func (bw *BatchWriter) worker() {
	defer bw.wg.Done()
	for {
		select {
		case <-bw.ctx.Done():
			bw.doFlush(true)
			return
		case <-bw.flushReqCh:
			bw.doFlush(false)
		case <-bw.refreshT.C:
			bw.doFlush(false)
		case <-bw.recoverT.C:
			bw.replayFromRedisBackground()
		}
	}
}

// doFlush 将当前队列安全交换出去并交给 handler（含重试）
func (bw *BatchWriter) doFlush(force bool) {
	bw.mu.Lock()
	qLen := len(bw.queue)
	if qLen == 0 {
		bw.mu.Unlock()
		return
	}

	if !force && qLen < bw.config.BatchSize && time.Since(bw.lastFlush) < bw.config.FlushInterval {
		bw.mu.Unlock()
		return
	}

	items := bw.queue
	bw.queue = make([][]byte, 0, bw.config.BatchSize)
	bw.lastFlush = time.Now()
	bw.mu.Unlock()

	if err := bw.callHandlerWithRetry(items); err != nil {
		bw.log("error", "handler final failed, saving to redis",
			common.Field{"count", len(items)},
			common.Field{"error", err})
		if err2 := bw.saveToRedisWithRetry(items); err2 != nil {
			bw.log("error", "save to redis failed", common.Field{"error", err2})
		}
	} else {
		bw.log("info", "flush success", common.Field{"count", len(items)})
	}
}

// Shutdown 优雅关闭
func (bw *BatchWriter) Shutdown(timeout time.Duration) error {
	bw.log("info", "shutting down")

	bw.mu.Lock()
	if !bw.started || bw.closed {
		bw.mu.Unlock()
		return nil
	}
	bw.closed = true
	bw.mu.Unlock()

	bw.cancel()

	c := make(chan struct{})
	go func() {
		bw.wg.Wait()
		close(c)
	}()

	select {
	case <-c:
	case <-time.After(timeout):
		bw.log("warn", "shutdown timeout")
	}

	// 保存剩余队列到 redis
	bw.mu.Lock()
	remaining := len(bw.queue)
	if remaining > 0 {
		items := bw.queue
		bw.queue = nil
		bw.mu.Unlock()
		if err := bw.saveToRedisWithRetry(items); err != nil {
			bw.log("error", "save remaining to redis failed", common.Field{"error", err})
			return err
		}
	} else {
		bw.mu.Unlock()
	}

	bw.log("info", "shutdown complete")
	return nil
}

// GetQueueSize 获取队列大小
func (bw *BatchWriter) GetQueueSize() int {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	return len(bw.queue)
}

// GetStats 获取状态
func (bw *BatchWriter) GetStats() map[string]interface{} {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	return map[string]interface{}{
		"name":      bw.config.Name,
		"queueSize": len(bw.queue),
		"batchSize": bw.config.BatchSize,
		"lastFlush": bw.lastFlush.Format(time.RFC3339),
		"started":   bw.started,
	}
}
