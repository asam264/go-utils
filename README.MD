# Go Utils

A collection of lightweight Go utilities for common development tasks.

## Installation

```bash
go get github.com/yourusername/go-utils
```

## Modules

### BatchWriter

High-performance batch processing with automatic retry and Redis persistence support.

**Features:**
- Automatic batching based on size or time interval
- Exponential backoff retry on failures
- Optional Redis persistence for failed batches
- Graceful shutdown with data safety
- Built-in manager for multiple instances

**Basic Usage:**

```go
import "github.com/yourusername/go-utils/batchwriter"

// Create a handler to process batches
handler := func(items [][]byte) error {
    // Your batch processing logic
    return nil
}

// Configure and create writer
cfg := batchwriter.Config{
    Name:          "my_writer",
    BatchSize:     50,
    FlushInterval: 5 * time.Second,
    Handler:       handler,
}

bw, err := batchwriter.New(cfg)
if err != nil {
    panic(err)
}

// Start processing
bw.Start()
defer bw.Shutdown(5 * time.Second)

// Push data
bw.Push([]byte("your data"))
```

**With Redis Persistence:**

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/yourusername/go-utils/batchwriter"
)

redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

cfg := batchwriter.Config{
    Name:        "persistent_writer",
    Handler:     handler,
    RedisClient: batchwriter.NewGoRedisAdapter(redisClient),
}

bw, _ := batchwriter.New(cfg)
bw.Start()
```

**Configuration Options:**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| Name | string | required | Writer identifier |
| BatchSize | int | 20 | Items per batch |
| FlushInterval | time.Duration | 5s | Max time between flushes |
| Handler | Handler | required | Batch processing function |
| RedisClient | RedisClient | nil | Redis client for persistence |
| Logger | Logger | noop | Custom logger |
| MaxRetries | int | 5 | Retry attempts on failure |
| BaseBackoff | time.Duration | 500ms | Initial retry delay |
| MaxBackoff | time.Duration | 10s | Maximum retry delay |
| RecoverPeriod | time.Duration | 1m | Redis recovery interval |

**Logger Adapters:**

```go
// Zap logger
import "go.uber.org/zap"
logger, _ := zap.NewProduction()
cfg.Logger = batchwriter.NewZapLoggerAdapter(logger)

// Simple console logger
cfg.Logger = batchwriter.NewSimpleLogger(true, true, true)

// Custom logger (implement Logger interface)
type MyLogger struct{}
func (l *MyLogger) Info(msg string, fields ...batchwriter.Field) {}
func (l *MyLogger) Warn(msg string, fields ...batchwriter.Field) {}
func (l *MyLogger) Error(msg string, fields ...batchwriter.Field) {}
```

**Manager for Multiple Writers:**

```go
manager := batchwriter.GetDefaultManager()

// Register and start multiple writers
for i := 0; i < 3; i++ {
    bw, _ := batchwriter.New(cfg)
    manager.RegisterAndStart(fmt.Sprintf("writer-%d", i), bw)
}

// Get a specific writer
writer, ok := manager.Get("writer-0")

// Get statistics
stats := manager.GetAllStats()

// Shutdown all writers
manager.ShutdownAll(5 * time.Second)
```

## Common Use Cases

- **Database Batch Inserts**: Reduce database round trips by batching INSERT operations
- **Log Aggregation**: Collect logs and write them in batches to storage
- **Event Processing**: Buffer events and process them in batches
- **Metrics Collection**: Aggregate metrics before sending to monitoring systems
- **API Rate Limiting**: Batch API requests to stay within rate limits

## Requirements

- Go 1.21 or higher
- Optional: Redis for persistence (go-redis v9)
- Optional: Zap for structured logging

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.

## Roadmap

Future utility modules:
- SQL helpers (query builder, batch operations)
- Cache utilities (memory, distributed)
- More utilities based on community needs