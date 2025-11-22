package common

import "fmt"

// Logger 日志接口，允许用户提供自己的日志实现
type Logger interface {
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

// noopLogger 空日志实现
type noopLogger struct{}

func (l *noopLogger) Info(msg string, fields ...Field)  {}
func (l *noopLogger) Warn(msg string, fields ...Field)  {}
func (l *noopLogger) Error(msg string, fields ...Field) {}

// NewNoopLogger 创建空日志器
func NewNoopLogger() Logger {
	return &noopLogger{}
}

// SimpleLogger 简单控制台日志实现
type SimpleLogger struct {
	enableInfo  bool
	enableWarn  bool
	enableError bool
}

// NewSimpleLogger 创建简单日志器
// enableInfo, enableWarn, enableError 分别控制是否输出对应级别的日志
func NewSimpleLogger(enableInfo, enableWarn, enableError bool) *SimpleLogger {
	return &SimpleLogger{
		enableInfo:  enableInfo,
		enableWarn:  enableWarn,
		enableError: enableError,
	}
}

// Info 记录 Info 级别日志
func (l *SimpleLogger) Info(msg string, fields ...Field) {
	if !l.enableInfo {
		return
	}
	fmt.Printf("[INFO] %s", msg)
	for _, f := range fields {
		fmt.Printf(" %s=%v", f.Key, f.Value)
	}
	fmt.Println()
}

// Warn 记录 Warn 级别日志
func (l *SimpleLogger) Warn(msg string, fields ...Field) {
	if !l.enableWarn {
		return
	}
	fmt.Printf("[WARN] %s", msg)
	for _, f := range fields {
		fmt.Printf(" %s=%v", f.Key, f.Value)
	}
	fmt.Println()
}

// Error 记录 Error 级别日志
func (l *SimpleLogger) Error(msg string, fields ...Field) {
	if !l.enableError {
		return
	}
	fmt.Printf("[ERROR] %s", msg)
	for _, f := range fields {
		fmt.Printf(" %s=%v", f.Key, f.Value)
	}
	fmt.Println()
}
