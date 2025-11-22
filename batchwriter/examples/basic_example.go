// Package main - BatchWriter 使用示例
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/asam264/go-utils/batchwriter"
	"github.com/asam264/go-utils/common"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ==================== 示例 1: 基本使用 ====================
func Example_Basic() {
	// 创建处理函数
	handler := func(items [][]byte) error {
		fmt.Printf("Processing %d items\n", len(items))
		for i, item := range items {
			fmt.Printf("  Item %d: %s\n", i, string(item))
		}
		return nil
	}

	// 创建配置
	cfg := batchwriter.Config{
		Name:          "example_writer",
		BatchSize:     5,
		FlushInterval: 3 * time.Second,
		Handler:       handler,
		Logger:        common.NewSimpleLogger(true, true, true),
	}

	// 创建 BatchWriter
	bw, err := batchwriter.New(cfg)
	if err != nil {
		panic(err)
	}

	// 启动
	if err := bw.Start(); err != nil {
		panic(err)
	}
	defer bw.Shutdown(5 * time.Second)

	// 推送数据
	for i := 0; i < 12; i++ {
		data := []byte(fmt.Sprintf("message-%d", i))
		if err := bw.Push(data); err != nil {
			fmt.Printf("Push error: %v\n", err)
		}
	}

	// 等待处理完成
	time.Sleep(5 * time.Second)
}

// ==================== 示例 2: 使用 Redis 持久化 ====================
func Example_WithRedis() {
	// 创建 Redis 客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer redisClient.Close()

	// 创建 zap logger
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	// 创建处理函数（可能会失败）
	handler := func(items [][]byte) error {
		// 模拟处理逻辑
		fmt.Printf("Processing batch of %d items\n", len(items))

		// 模拟偶尔失败
		// if rand.Float32() < 0.3 {
		//     return fmt.Errorf("simulated failure")
		// }

		return nil
	}

	// 创建配置
	cfg := batchwriter.Config{
		Name:          "redis_writer",
		BatchSize:     10,
		FlushInterval: 5 * time.Second,
		Handler:       handler,
		RedisClient:   batchwriter.NewGoRedisAdapter(redisClient),
		Logger:        batchwriter.NewZapLoggerAdapter(zapLogger),
		MaxRetries:    3,
		RecoverPeriod: 30 * time.Second,
	}

	// 创建并启动
	bw, err := batchwriter.New(cfg)
	if err != nil {
		panic(err)
	}

	if err := bw.Start(); err != nil {
		panic(err)
	}
	defer bw.Shutdown(10 * time.Second)

	// 推送数据
	for i := 0; i < 100; i++ {
		data := []byte(fmt.Sprintf("data-%d", i))
		if err := bw.Push(data); err != nil {
			fmt.Printf("Push error: %v\n", err)
		}
	}

	time.Sleep(10 * time.Second)
}

// ==================== 示例 3: 使用 Manager 管理多个 Writer ====================
func Example_WithManager() {
	manager := batchwriter.GetDefaultManager()

	// 创建多个 writer
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("writer-%d", i)
		handler := func(writerName string) batchwriter.Handler {
			return func(items [][]byte) error {
				fmt.Printf("[%s] Processing %d items\n", writerName, len(items))
				return nil
			}
		}(name)

		cfg := batchwriter.Config{
			Name:          name,
			BatchSize:     5,
			FlushInterval: 2 * time.Second,
			Handler:       handler,
			Logger:        common.NewSimpleLogger(true, true, true),
		}

		bw, err := batchwriter.New(cfg)
		if err != nil {
			panic(err)
		}

		if err := manager.RegisterAndStart(name, bw); err != nil {
			panic(err)
		}
	}

	// 使用不同的 writer
	for i := 0; i < 20; i++ {
		writerName := fmt.Sprintf("writer-%d", i%3)
		writer, ok := manager.Get(writerName)
		if ok {
			data := []byte(fmt.Sprintf("message-%d", i))
			writer.Push(data)
		}
	}

	time.Sleep(5 * time.Second)

	// 查看统计
	stats := manager.GetAllStats()
	fmt.Printf("Stats: %+v\n", stats)

	// 关闭所有
	manager.ShutdownAll(5 * time.Second)
}

// ==================== 示例 4: 处理结构化数据 ====================
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	UserID    string    `json:"user_id"`
}

func Example_StructuredData() {
	// 批量写入日志到数据库
	handler := func(items [][]byte) error {
		var entries []LogEntry
		for _, item := range items {
			var entry LogEntry
			if err := json.Unmarshal(item, &entry); err != nil {
				return fmt.Errorf("unmarshal error: %w", err)
			}
			entries = append(entries, entry)
		}

		// 批量插入数据库
		fmt.Printf("Inserting %d log entries to database\n", len(entries))
		// db.BatchInsert(entries)

		return nil
	}

	cfg := batchwriter.Config{
		Name:          "log_writer",
		BatchSize:     50,
		FlushInterval: 10 * time.Second,
		Handler:       handler,
		Logger:        common.NewSimpleLogger(true, true, true),
	}

	bw, err := batchwriter.New(cfg)
	if err != nil {
		panic(err)
	}

	if err := bw.Start(); err != nil {
		panic(err)
	}
	defer bw.Shutdown(5 * time.Second)

	// 推送日志
	for i := 0; i < 100; i++ {
		entry := LogEntry{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   fmt.Sprintf("Log message %d", i),
			UserID:    fmt.Sprintf("user-%d", i%10),
		}

		data, _ := json.Marshal(entry)
		if err := bw.Push(data); err != nil {
			fmt.Printf("Push error: %v\n", err)
		}
	}

	time.Sleep(15 * time.Second)
}

// ==================== 示例 5: 集成到 HTTP 服务 ====================
func Example_HTTPService() {
	// 创建全局 writer
	var globalWriter *batchwriter.BatchWriter

	// 初始化
	func() {
		handler := func(items [][]byte) error {
			// 批量处理请求日志
			fmt.Printf("Processing %d request logs\n", len(items))
			return nil
		}

		cfg := batchwriter.Config{
			Name:          "request_logger",
			BatchSize:     100,
			FlushInterval: 5 * time.Second,
			Handler:       handler,
		}

		var err error
		globalWriter, err = batchwriter.New(cfg)
		if err != nil {
			panic(err)
		}
		globalWriter.Start()
	}()

	// HTTP 中间件示例
	logRequest := func(method, path string, duration time.Duration) {
		logData := fmt.Sprintf("%s %s - %v", method, path, duration)
		globalWriter.Push([]byte(logData))
	}

	// 模拟请求
	for i := 0; i < 1000; i++ {
		logRequest("GET", "/api/users", time.Millisecond*100)
	}

	time.Sleep(10 * time.Second)
	globalWriter.Shutdown(5 * time.Second)
}

func main() {
	fmt.Println("=== Example 1: Basic Usage ===")
	Example_Basic()

	fmt.Println("\n=== Example 3: Manager ===")
	Example_WithManager()

	fmt.Println("\n=== Example 4: Structured Data ===")
	Example_StructuredData()

	fmt.Println("\n=== Example 5: HTTP Service ===")
	Example_HTTPService()
}
