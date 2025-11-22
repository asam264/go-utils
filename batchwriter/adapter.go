package batchwriter

import (
	"context"
	"fmt"

	"github.com/asam264/go-utils/common"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// GoRedisAdapter Redis 适配器，适配 go-redis/v9
type GoRedisAdapter struct {
	client *redis.Client
}

// NewGoRedisAdapter 创建 go-redis 适配器
func NewGoRedisAdapter(client *redis.Client) *GoRedisAdapter {
	return &GoRedisAdapter{client: client}
}

// LPop 从 Redis list 左侧弹出元素
func (a *GoRedisAdapter) LPop(ctx context.Context, key string) ([]byte, error) {
	result := a.client.LPop(ctx, key)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key not found")
		}
		return nil, err
	}
	return []byte(result.Val()), nil
}

// RPush 向 Redis list 右侧推入元素
func (a *GoRedisAdapter) RPush(ctx context.Context, key string, values ...[]byte) error {
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = v
	}
	return a.client.RPush(ctx, key, args...).Err()
}

// ZapLoggerAdapter Zap 日志适配器
type ZapLoggerAdapter struct {
	logger *zap.Logger
}

// NewZapLoggerAdapter 创建 Zap 日志适配器
func NewZapLoggerAdapter(logger *zap.Logger) *ZapLoggerAdapter {
	return &ZapLoggerAdapter{logger: logger}
}

// Info 记录 Info 级别日志
func (l *ZapLoggerAdapter) Info(msg string, fields ...common.Field) {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	l.logger.Info(msg, zapFields...)
}

// Warn 记录 Warn 级别日志
func (l *ZapLoggerAdapter) Warn(msg string, fields ...common.Field) {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	l.logger.Warn(msg, zapFields...)
}

// Error 记录 Error 级别日志
func (l *ZapLoggerAdapter) Error(msg string, fields ...common.Field) {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	l.logger.Error(msg, zapFields...)
}

