package batchwriter

import (
	"fmt"
	"time"

	"github.com/asam264/go-utils/common"
)

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	if cfg.Name == "" {
		return fmt.Errorf("name is required")
	}
	if cfg.Handler == nil {
		return fmt.Errorf("handler is required")
	}
	return nil
}

// applyDefaults 应用默认配置
func applyDefaults(cfg *Config) {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 20
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 5 * time.Second
	}
	if cfg.RedisKeyPrefix == "" {
		cfg.RedisKeyPrefix = "batch_writer"
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 5
	}
	if cfg.BaseBackoff <= 0 {
		cfg.BaseBackoff = 500 * time.Millisecond
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 10 * time.Second
	}
	if cfg.RecoverPeriod <= 0 {
		cfg.RecoverPeriod = 1 * time.Minute
	}
	if cfg.Logger == nil {
		cfg.Logger = common.NewNoopLogger()
	}
}
