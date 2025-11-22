package batchwriter

import (
	"fmt"
	"sync"
	"time"
)

// Manager 管理多个 BatchWriter 实例
type Manager struct {
	mu       sync.RWMutex
	writers  map[string]*BatchWriter
	shutdown bool
}

var (
	defaultManager *Manager
	once           sync.Once
)

// GetDefaultManager 获取默认的 Manager 实例（单例）
func GetDefaultManager() *Manager {
	once.Do(func() {
		defaultManager = &Manager{
			writers: make(map[string]*BatchWriter),
		}
	})
	return defaultManager
}

// Register 注册一个 BatchWriter
func (m *Manager) Register(name string, writer *BatchWriter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shutdown {
		return fmt.Errorf("manager is shutdown")
	}

	if _, exists := m.writers[name]; exists {
		return fmt.Errorf("writer %s already registered", name)
	}

	m.writers[name] = writer
	return nil
}

// RegisterAndStart 注册并启动一个 BatchWriter
func (m *Manager) RegisterAndStart(name string, writer *BatchWriter) error {
	if err := m.Register(name, writer); err != nil {
		return err
	}
	return writer.Start()
}

// Get 获取指定名称的 BatchWriter
func (m *Manager) Get(name string) (*BatchWriter, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	writer, ok := m.writers[name]
	return writer, ok
}

// Unregister 取消注册一个 BatchWriter
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	writer, exists := m.writers[name]
	if !exists {
		return fmt.Errorf("writer %s not found", name)
	}

	delete(m.writers, name)
	return writer.Shutdown(5 * time.Second)
}

// GetAllStats 获取所有 BatchWriter 的统计信息
func (m *Manager) GetAllStats() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]map[string]interface{})
	for name, writer := range m.writers {
		stats[name] = writer.GetStats()
	}
	return stats
}

// ShutdownAll 关闭所有 BatchWriter
func (m *Manager) ShutdownAll(timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shutdown {
		return nil
	}
	m.shutdown = true

	var wg sync.WaitGroup
	errors := make(chan error, len(m.writers))

	for name, writer := range m.writers {
		wg.Add(1)
		go func(n string, w *BatchWriter) {
			defer wg.Done()
			if err := w.Shutdown(timeout); err != nil {
				errors <- fmt.Errorf("shutdown %s failed: %w", n, err)
			}
		}(name, writer)
	}

	wg.Wait()
	close(errors)

	var firstErr error
	for err := range errors {
		if firstErr == nil {
			firstErr = err
		}
	}

	m.writers = make(map[string]*BatchWriter)
	return firstErr
}

// List 列出所有已注册的 BatchWriter 名称
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.writers))
	for name := range m.writers {
		names = append(names, name)
	}
	return names
}
